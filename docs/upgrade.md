## Upgrading Kubernetes

Upgrading Kubernetes is easy with kops. The cluster spec contains a `KubernetesVersion`, so you can simply edit it with `kops edit`, and apply the updated configuration to your cluster.

The `kops upgrade` command also automates checking for and applying updates.

Note: if you want to upgrade from a `kube-up` installation, please see the instructions for [how to upgrade kubernetes installed with kube-up](cluster_upgrades_and_migrations.md).

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

Upgrade uses the latest Kubernetes version considered stable by kops, defined in `https://github.com/kubernetes/kops/blob/master/channels/stable`.

NOTE: rolling-update does not yet perform a real rolling update - it just shuts down machines in sequence with a delay;
 there will be downtime [Issue #37](https://github.com/kubernetes/kops/issues/37)
We have implemented a new feature that does drain and validate nodes.  This feature is experimental, and you can use the new feature by setting `export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"`.

