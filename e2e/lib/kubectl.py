import datetime
import downloads
import json
import os
import e2e

# k is the channel to get; e.g. "stable", "latest", "stable-1.16", "latest-1.16"
# List at `gsutil ls gs://kubernetes-release/release | grep txt`
def get_kubernetes_version(k):
    latest_url = (
        "https://storage.googleapis.com/kubernetes-release/release/" + k + ".txt"
    )
    return downloads.read_url(latest_url).strip()


def download_kubectl(k8s_version):
    now = datetime.datetime.utcnow()
    now = now.replace(microsecond=0)

    scratch_dir = os.path.join(
        e2e.workspace_dir(), "kubectl-scratch-" + now.strftime("%Y%m%d%H%M%s")
    )

    # We build up symlinks to the downloaded binaries in the bin directory
    bin_dir = os.path.join(scratch_dir, "bin")
    os.makedirs(bin_dir, exist_ok=True)

    url = (
        "https://storage.googleapis.com/kubernetes-release/release/"
        + k8s_version
        + "/bin/linux/amd64/kubectl"
    )
    kubectl = downloads.download_hashed_url(url)
    downloads.exec(["chmod", "+x", kubectl])
    kubectl_path = os.path.join(bin_dir, "kubectl")
    os.symlink(kubectl, kubectl_path)
    # os.symlink(kubectl, "/bin/kubectl")

    return Kubectl(kubectl_path)


def local_kubectl():
    return Kubectl("kubectl")


class Kubectl(object):
    def __init__(self, bin, env=None):
        if env is None:
            env = os.environ.copy()
        self.bin = os.path.expanduser(bin)
        self.env = env

    def __repr__(self):
        s = "Kubectl:" + self.bin
        return s

    def add_to_path(self, env):
        d = os.path.dirname(self.bin)
        env["PATH"] = d + ":" + env["PATH"]

    def version(self):
        return self.exec(["version", "--client"])

    def server_version(self):
        return self.exec(["version"])

    def exec(self, args):
        return downloads.exec([self.bin] + args, env=self.env).strip()

    def exec_and_parse_json(self, args):
        j = downloads.exec([self.bin, "-ojson"] + args, env=self.env).strip()
        return json.loads(j)
