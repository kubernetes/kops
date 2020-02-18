import datetime
import os
import os.path
import subprocess
import tempfile
import xmltodict

import gcloud

_artifacts_dir = None

def artifacts_dir():
  global _artifacts_dir
  if _artifacts_dir is None:
    wd = workspace_dir()
    _artifacts_dir = os.path.join(wd, "artifacts")
    os.makedirs(_artifacts_dir, exist_ok=True)
  return _artifacts_dir

_workspace_dir = None

def workspace_dir():
  workspace = os.environ.get("WORKSPACE")
  if workspace:
    _workspace_dir = workspace
  else:
    _workspace_dir = tempfile.mkdtemp(prefix='tmp-e2e')
  return _workspace_dir

import downloads

def download_kubetest(k8s_version):
    now = datetime.datetime.utcnow()
    now = now.replace(microsecond=0)

    scratch_dir = os.path.join(workspace_dir(), "e2e-scratch-" + now.strftime("%Y%m%d%H%M%s"))
    results_dir = os.path.join(artifacts_dir(), "e2e-results-" + now.strftime("%Y%m%d%H%M%s"))
    os.makedirs(results_dir, exist_ok=True)

    # We build up symlinks to the downloaded binaries in the bin directory
    bin_dir = os.path.join(scratch_dir, "bin")
    os.makedirs(bin_dir, exist_ok=True)

    url = "https://storage.googleapis.com/kubernetes-release/release/" + k8s_version + "/kubernetes-test-linux-amd64.tar.gz"
    tarfile = downloads.download_hashed_url(url)
    expanded = downloads.expand_tar(tarfile)
    os.symlink(os.path.join(expanded, "kubernetes/test/bin/e2e.test"), os.path.join(bin_dir, "e2e.test"))
    os.symlink(os.path.join(expanded, "kubernetes/test/bin/ginkgo"), os.path.join(bin_dir, "ginkgo"))

    url = "https://storage.googleapis.com/kubernetes-release/release/" + k8s_version + "/bin/linux/amd64/kubectl"
    kubectl = downloads.download_hashed_url(url)
    downloads.exec(["chmod", "+x", kubectl])
    os.symlink(kubectl, os.path.join(bin_dir, "kubectl"))
    #os.symlink(kubectl, "/bin/kubectl")

    return E2E(k8s_version, bin_dir=bin_dir, results_dir=results_dir)

class TestRun(object):
  def __init__(self):
    self.all_tests = []
    self.passed_tests = []
    self.skipped_tests = []
    self.failed_tests = []

  def __repr__(self):
    return "TestRun: %s passed, %d failed, %d skipped" % (len(self.passed_tests), len(self.failed_tests), len(self.skipped_tests))

  def merge(self, r):
    o = TestRun()
    o.all_tests = self.all_tests + r.all_tests
    o.passed_tests = self.passed_tests + r.passed_tests
    o.skipped_tests = self.skipped_tests + r.skipped_tests
    o.failed_tests = self.failed_tests + r.failed_tests
    return o
  
  def _parse_results(self, junit):
    testsuite = junit.get('testsuite')
    for testcase in testsuite.get('testcase'):
      name = testcase.get('@name')
      o = { "name": name, "data": testcase }

      if '@time' in testcase:
        o['time'] = float(testcase.get('@time'))

      self.all_tests.append(o)
      if 'skipped' in testcase:
        o['state'] = 'skipped'
        self.skipped_tests.append(o)
      elif 'failure' in testcase:
        o['state'] = 'fail'
        self.failed_tests.append(o)
      else:
        o['state'] = 'pass'
        self.passed_tests.append(o)

class E2E(object):
  def __init__(self, k8s_version, bin_dir, results_dir):
    self.k8s_version = k8s_version
    self.bin_dir = bin_dir
    self.results_dir = results_dir

  def __repr__(self):
    return "E2E:" + self.k8s_version

  def run(self, focus=None, skip=None, parallel=None):
    report_dir = self.results_dir
    logfile = os.path.join(self.results_dir, "e2e.log")

    args = []
    kubeconfig = os.environ.get("KUBECONFIG")
    if not kubeconfig:
      kubeconfig = os.path.expanduser("~/.kube/config")

    args += [ "--report-dir=" + report_dir ]
    args += [ "--v=4" ]
    
    env = os.environ.copy()
    env["KUBECONFIG"] = kubeconfig
    env["PATH"] = self.bin_dir + ":" + env["PATH"]
    
    # Verify kubectl is available on the PATH
    r = downloads.exec([os.path.join(self.bin_dir, "kubectl"), "version"], env=env)

    bin_e2e = os.path.join(self.bin_dir, "e2e.test")
    bin_ginkgo = os.path.join(self.bin_dir, "ginkgo")

    ginkgo_args = [bin_ginkgo]
    ginkgo_args += ["--noColor", "--succinct"]

    if focus:
      ginkgo_args += ["--focus", focus]
    if skip:
      ginkgo_args += ["--skip", skip]
    if parallel:
      ginkgo_args += ["--nodes", str(parallel)]
      
    #		if options.DryRun {
    #			ginkgoArgs = append(ginkgoArgs, "--dryRun")
    #		}

    ginkgo_args += [bin_e2e, "--"]
    ginkgo_args += args

    print("running %s, logging to %s" % (ginkgo_args, logfile))
    with open(logfile, "w+") as f:
      r = subprocess.run(ginkgo_args, env=env, stdout=f, stderr=f)
    if r.returncode == 1:
      # TODO: Is this always due to test failure?
      print("ginkgo exited with exit code 1")
    elif r.returncode != 0:
      #print("stdout")
      #print(r.stdout.decode())
      #print("stderr")
      #print(r.stderr.decode())
      r.check_returncode()
    #print(r.stdout.decode())

    tr = TestRun()
    for f in os.listdir(report_dir):
      if f.startswith("junit_") and f.endswith(".xml"):
        print(f)
        with open(os.path.join(report_dir, f)) as fd:
          junit = xmltodict.parse(fd.read())
          tr._parse_results(junit)
    print(tr)
    return tr
