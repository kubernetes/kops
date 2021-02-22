# Architecture: kops-controller

kops-controller runs as a DaemonSet on the master node(s). It is a kubebuilder
controller that performs runtime reconciliation for kOps.

Controllers in kops-controller:

* NodeController
* MetricsServer


## NodeController

The NodeController watches Node objects, and applies labels to them from a
controller.  Previously this was done by the kubelet, but the fear was that this
was insecure and so [this functionality was removed](https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/0000-20170814-bounding-self-labeling-kubelets.md).

The main difficulty here is mapping from a Node to an InstanceGroup in a way
that does not render the system just as vulnerable to spoofing as it was
previously.

NodeController uses the cloud APIs to make this link (in future, cluster-api may
offer an alternative).  The theory is that we can then work to prevent spoofing
of the Node's `providerID`, and further we assume that an attacker that has
gained the ability to manipulate the underlying cloud itself has already
bypassed our protections.

On AWS, tags are not easily mutable from a Node; so we simply set a tag
with the name of the instance-group.  When we see a node, we query EC2 for the
instance defined in `providerID`, and we get the instance group name from the
tag.  We then query the instance group definition from the underlying store
(typically S3), construct the correct tags, and apply the tags.

On GCE, the metadata is more mutable.  So we query the instance, but then we
find the owning MIG, query the instances that are part of that MIG to verify
that the instance is indeed part of the MIG, and then we get the metadata from
the instance template (which is not easily mutated from the instance).  We then
get the instance group definition from the underlying store, as elsewhere.

## MetricsServer

The metrics server exposes validation results for prometheus. This is a prometheus exporter, and it collects these gauge values:

- `kops_component_healthy`
- `kops_component_unhealthy`
- `kops_node_notReady`
- `kops_node_ready`
- `kops_pod_missing`
- `kops_pod_notReady`
- `kops_pod_pending`
- `kops_pod_unknown`

Internally, these values are calculated using the same logic as `kops validate cluster`.

And you can use this metrics server to call endpoint `http://kops-controller.kube-system:3987/metrics` from your prometheus. For example:

```
$ curl http://kops-controller.kube-system:3987/metrics
# HELP kops_component_healthy Healthy components of the kubernetes cluster.
# TYPE kops_component_healthy gauge
kops_component_healthy{cluster_name="clustername.mydomain.com"} 4
# HELP kops_component_unhealthy Unhealthy components of the kubernetes cluster.
# TYPE kops_component_unhealthy gauge
kops_component_unhealthy{cluster_name="clustername.mydomain.com"} 0
# HELP kops_node_notReady Not ready nodes in the kubernetes cluster.
# TYPE kops_node_notReady gauge
kops_node_notReady{cluster_name="clustername.mydomain.com"} 0
# HELP kops_node_ready Ready nodes in the kubernetes cluster.
# TYPE kops_node_ready gauge
kops_node_ready{cluster_name="clustername.mydomain.com"} 2
# HELP kops_pod_missing Missing pods in the kubernetes cluster.
# TYPE kops_pod_missing gauge
kops_pod_missing{cluster_name="clustername.mydomain.com"} 0
# HELP kops_pod_notReady Not ready pods in the kubernetes cluster.
# TYPE kops_pod_notReady gauge
kops_pod_notReady{cluster_name="clustername.mydomain.com"} 0
# HELP kops_pod_pending Pending pods in the kubernetes cluster.
# TYPE kops_pod_pending gauge
kops_pod_pending{cluster_name="clustername.mydomain.com"} 0
# HELP kops_pod_unknown Unknown pods in the kubernetes cluster.
# TYPE kops_pod_unknown gauge
kops_pod_unknown{cluster_name="clustername.mydomain.com"} 0
```
