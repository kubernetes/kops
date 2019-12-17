# Calico Pod CIDR Migration Procedure
Prior to kops 1.7, calico, and other CNI providers was misconfigured to use the 
`.NonMasqueradeCIDR` field as the CIDR range for Pod IPs. As a result, IP
conflict may occur when a Service is allocated an IP that has already been
assigned to a Pod, or vice versa. To prevent this from occurring, manual steps
are necessary before upgrading your cluster using kops 1.7 onwards.


## Background
The field in the clusterSpec, `.NonMasqueradeCIDR`, captures the IP
range of the cluster.

Within this IP range, smaller IP ranges are then carved out for:

* Service IPs - as defined as `.serviceClusterIPRange`
* Pod IPs - as defined as `.kubeControllerManager.clusterCIDR`

It was found out in Issue [#1171](https://github.com/kubernetes/kops/issues/1171),
that weave and calico were misconfigured to use the wider IP range rather than
the range dedicated to Pods only. This was fixed in PR [#2717](https://github.com/kubernetes/kops/pull/2717)
and [#2768](https://github.com/kubernetes/kops/pull/2768) for the two CNIs, by
switching over to using the `.kubeControllerManager.clusterCIDR` field instead.

With the `--ip-alloc-range` flag changes for weave, it effectively creates a
new network. Pods in the existing network will not necessarily be able to talk
to those in the new network. Restarting of all nodes will need to happen
to guarantee that all Pods spin up with IPs in the new network. See [here](
https://github.com/weaveworks/weave/issues/2874) for more details.

Just like weave, the above change alone is not enough to mitigate the problem
on existing clusters running calico. Effectively, a new network will need to be
created first (in the form of an IP Pool in Calico terms), and a restart of all 
nodes will need to happen to have Pods be allocated the proper IP addresses.

## Prerequisites

* `kops` >= 1.7
* `jq` for retrieving the field values from the clusterSpec
* Kubernetes cluster with calico as the CNI, created prior to kops 1.7
* Scheduled maintenance window, this procedure *WILL* result in cluster degregation.

## Procedure
**WARNING** - This procedure will cause disruption to Pods running on the cluster.
New Pods may not be able to resolve DNS through kube-dns or other services
through its service IP during the rolling restart.
Attempt this migration procedure on a staging cluster prior to doing this in production.

---
Calico only uses the `CALICO_IPV4POOL_CIDR` to create a default IPv4 pool if a
pool doesn't exist already:
https://github.com/projectcalico/calicoctl/blob/v1.3.0/calico_node/startup/startup.go#L463

Therefore, we need to run two jobs. We have provided a manifest and a bash script.
job creates a new IPv4 pool that we want, and deletes the existing IP
pool that we no longer want. This is to be executed after a
`kops update cluster --yes` using kops 1.7 and beyond,
and before a `kops rolling-upgrade cluster`.

1. Using kops >= 1.7, update your cluster using `kops update cluster [--yes]`.
2. Specify your cluster name in a `NAME` variable, download the template and bash script, and then run the bash script:
```bash
export NAME="YOUR_CLUSTER_NAME"
wget https://raw.githubusercontent.com/kubernetes/kops/master/docs/calico_cidr_migration/create_migration_manifest.sh -O create_migration_manifest.sh
wget https://raw.githubusercontent.com/kubernetes/kops/master/docs/calico_cidr_migration/jobs.yaml.template -O jobs.yaml.template
chmod +x create_migration_manifest.sh
./create_migration_manifest.sh
```
This will create a `jobs.yaml` manifest file that is used by the next step.

3. Make sure the current-context in the kubeconfig is the cluster you want to perform this migration.
Run the job: `kubectl apply -f jobs.yaml`
4. Run `kops rolling-update cluster --force --yes` to initiate a rolling restart on the cluster.
This forces a restart of all nodes in the cluster.

That's it, you should see new Pods be allocated IPs in the new IP range!
