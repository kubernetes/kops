## etcd-manager

etcd-manager is a kubernetes-associated project that kops will use to manage
etcd (at least that is the plan per the roadmap).

etcd-manager uses many of the same ideas as the existing etcd implementation
built into kops, but it addresses some limitations also:

* separate from kops - can be used by other projects
* allows etcd2 -> etcd3 upgrade (along with minor upgrades)
* allows cluster resizing (e.g. going from 1 to 3 nodes)

If using kubernetes >= 1.12 (which will not formally be supported until kops 1.12), note that etcd-manager will be used by default.  You can override this with the `cluster.spec.etcdClusters[*].provider=Legacy` override.  This can be specified:

* as an argument to `kops create cluster`: `--overrides cluster.spec.etcdClusters[*].provider=Legacy`
* on an existing cluster with `kops set cluster cluster.spec.etcdClusters[*].provider=Legacy`
* by setting the field using `kops edit` or `kops replace`, manually making the same change as `kops set cluster ...`

(Note you will probably have to `export KOPS_FEATURE_FLAGS=SpecOverrideFlag`)

## How to use etcd-manager

Reminder: etcd-manager is alpha, and we encourage you to back up the data in
your kubernetes cluster.

To create a test cluster:
```bash
kops create cluster test.k8s.local --zones us-east-1c --master-count 3
kops update cluster test.k8s.local --yes

# Wait for cluster to boot up
kubectl get nodes
```

You can enable the etcd-manager - it will adopt the existing etcd data, though
it won't change the configuration:

```bash
# Enable etcd-manager
kops set cluster cluster.spec.etcdClusters[*].provider=Manager

kops update cluster --yes
kops rolling-update cluster --yes
```

After the masters restart, you will still be running etcd 2.2.  You can change
the version of etcd:

```bash
kops set cluster cluster.spec.etcdClusters[*].version=3.2.18

kops update cluster --yes
kops rolling-update cluster --yes
```

It should be safe to combine the etcd-manager adoption and etcd upgrade into a
single restart, but we are working on boosting test coverage.

When you're done, you can shut down the cluster:

```bash
kops delete cluster example.k8s.local --yes
```

You can also do this for existing clusters. though remember that this is still
young software, so please back up important cluster data first.  You just run the
two `kops set cluster` commands.  Note that `kops set cluster` is just an easy
command line way to set some fields in the cluster spec - if you're using a
GitOps approach you can change the manifest files directly. You can also `kops
edit cluster`.

