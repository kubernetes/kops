## Upgrading Kubernetes

Upgrading Kubernetes is easy with kops.  The cluster spec contains a KubernetesVersion, so you
can simply edit it with `kops edit`, and apply the updated configuration to your cluster.
The `kops upgrade` command also automates checking for and applying updates.

Note: if you want to upgrade from a `kube-up` installation, please see the instructions for
[how to upgrade kubernetes installed with kube-up](upgrade_from_k8s_12.md).

### Manual update

* `kops edit cluster $NAME`
* set the KubernetesVersion to the target version (e.g. `v1.3.5`)
* `kops update cluster $NAME` to preview, then `kops update cluster $NAME --yes`
* `kops rolling-update cluster $NAME` to preview, then `kops rolling-update cluster $NAME --yes`

### Automated update

* `kops upgrade cluster $NAME` to preview, then `kops upgrade cluster $NAME --yes`

In future the upgrade step will likely perform the update immediately (and possibly even without a
node restart), but currently you must:

* `kops update cluster $NAME` to preview, then `kops update cluster $NAME --yes`
* `kops rolling-update cluster $NAME` to preview, then `kops rolling-update cluster $NAME --yes`


upgrade uses the latest Kubernetes stable release, published at `https://storage.googleapis.com/kubernetes-release/release/stable.txt`
