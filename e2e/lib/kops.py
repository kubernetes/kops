import subprocess, os, json
#import k8s

import downloads

def get_kops_ci_latest():
    latest_url = "https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt"
    return downloads.read_url(latest_url)

def get_kubernetes_version(k):
    latest_url = "https://storage.googleapis.com/kubernetes-release/release/" + k + ".txt"
    return downloads.read_url(latest_url).strip()

def download_kops(kops_base_url):
    kops_bin_url = kops_base_url + '/linux/amd64/kops'
    kops_bin = downloads.download_hashed_url(kops_bin_url)
    downloads.exec(["chmod", "+x", kops_bin])

    env = os.environ.copy()
    env["KOPS_BASE_URL"] = kops_base_url

    return Kops(kops_bin, env)


class Kops(object):
    def __init__(self, bin, env = None):
        if env is None:
          env = os.environ.copy()
        env["KOPS_FEATURE_FLAGS"] = "AlphaAllowGCE,SpecOverrideFlag"
        self.bin = os.path.expanduser(bin)
        self.env = env

    def version(self):
        version = self.exec(["version"])
        version = version.strip()
        return version

    def short_version(self):
        version = self.exec(["version", "--short"])
        version = version.strip()
        return version

    def clusters(self):
        stdout = self.exec(["get", "clusters", "-ojson"])
        clusters = []
        for line in stdout.splitlines():
            j = json.loads(line)
            clusters.append(KopsCluster(self, j))
        return clusters

    def cluster(self, name):
        stdout = self.exec(["get", "cluster", name, "-ojson"])
        j = json.loads(stdout)
        return KopsCluster(self, j)

    def create_cluster(self, spec):
      name = spec["name"]
      args = ["create", "cluster", name]
      for k, v in spec.items():
        if k == "name":
          continue
        if k == "cluster":
          continue
        if k == "zones":
          args = args + [ "--zones", ",".join(v) ]
        else:
          args = args + [ "--" + k.replace("_", "-"), "%s" % (v)]
      cluster_override = spec.get("cluster", None)
      if cluster_override:
        for k, v in _flatten("cluster", cluster_override).items():
          args = args + [ "--override", k + "=" + v ]
      stdout = self.exec(args)
      return self.cluster(name)

    def exec(self, args):
      return downloads.exec([self.bin] + args, env=self.env)

def _flatten(prefix, m):
  out = {}
  for k, v in m.items():
    child_k = (prefix + "." + k).strip(".")
    if isinstance(v, dict):
      for k, v in _flatten(child_k, v).items():
        out[k] = v
    else:
      out[child_k] = v
  return out

class KopsCluster(object):
    def __init__(self, kops, j):
        self.kops = kops
        self.metadata = j.get('metadata')
        self.kind = j.get('kind')
        self.spec = j.get('spec')

    def name(self):
        return self.metadata.get('name')

    def __repr__(self):
        return "KopsCluster:" + self.name()

    def instance_groups(self):
        stdout = self.kops.exec(["get", "instancegroups", "--name", self.name(), "-ojson"])
        igs = []
        objs = json.loads(stdout)
        for j in objs:
            igs.append(KopsInstanceGroup(self, j))
        return igs

    def objects(self):
        stdout = self.kops.exec(["get", "--name", self.name(), "-ojson"])
        objs = json.loads(stdout)
        return objs

    def instance_group(self, name):
        stdout = self.kops.exec(["get", "instancegroups", "--name", self.name(), name, "-ojson"])
        j = json.loads(stdout)
        return KopsInstanceGroup(self, j)

    def delete(self):
        stdout = self.kops.exec(["delete", "cluster", self.name(), "--yes"])
        print(stdout)

    def set(self, k, v):
        stdout = self.kops.exec(["set", "cluster", self.name(), k + "=" + v])
        print(stdout)

    def preview_update(self):
        stdout = self.kops.exec(["update", "cluster", self.name()])
        print(stdout)

    def update(self):
        stdout = self.kops.exec(["update", "cluster", self.name(), "--yes"])
        print(stdout)

    def preview_upgrade(self):
        stdout = self.kops.exec(["upgrade", "cluster", self.name()])
        print(stdout)

    def upgrade(self):
        stdout = self.kops.exec(["upgrade", "cluster", self.name(), "--yes"])
        print(stdout)

    def validate(self):
        stdout = self.kops.exec(["validate", "cluster", self.name()])
        print(stdout)

    def preview_rolling_update(self):
        stdout = self.kops.exec(["rolling-update", "cluster", self.name()])
        print(stdout)

    def rolling_update(self):
        stdout = self.kops.exec(["rolling-update", "cluster", self.name(), "--yes"])
        print(stdout)

    def wait(self):
        stdout = self.kops.exec(["validate", "cluster", self.name(), "--wait=10m"])
        print(stdout)

    def dump(self):
        stdout = self.kops.exec(["toolbox", "dump", self.name(), "-ojson"])
        return json.loads(stdout)

    def k8s(self):
      k = k8s.KubernetesClient(context=self.name())
      return k

    def apply(self):
      self.preview_update()
      self.update()
      self.wait()
      self.preview_rolling_update()
      self.rolling_update()

    def reconfigure(self, spec):
      if spec.get("kubernetes_version"):
        self.set("spec.kubernetesVersion", spec.get("kubernetes_version"))

class KopsInstanceGroup(object):
    def __init__(self, cluster, j):
        self.cluster = cluster
        self.metadata = j.get('metadata')
        self.kind = j.get('kind')
        self.spec = j.get('spec')
        self._json = j

    def name(self):
        return self.metadata.get('name')

    def __repr__(self):
        return "KopsInstanceGroup:" + self.name()
