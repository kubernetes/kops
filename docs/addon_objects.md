# Passing additional configuration objects

kOps has initial support for passing additional objects to the cluster, and recognizes a few "well known" objects.

This support is currently gated behind the `ClusterAddons` feature-flag (i.e. `export KOPS_FEATURE_FLAGS=ClusterAddons`)

Objects that are not well-known will be applied to the cluster.  Well-known objects will have special handling.

# Well-Known Objects

## KubeSchedulerConfiguration (group: kubescheduler.config.k8s.io)

KubeSchedulerConfiguration objects allow for custom configuration of
kube-scheduler, the component responsible for assigning Pods to Nodes.

Special handling:  the configuration will be written to a file on control plane nodes,
and the kube-scheduler component will be configured to read from that file.

Example usage:
```
export KOPS_FEATURE_FLAGS=ClusterAddons
kops create cluster --name=kubescheduler.k8s.local --zones us-east-2a --add docs/examples/addons/kubescheduler.yaml
kops update cluster --name=kubescheduler.k8s.local --yes --admin
kops validate cluster --name=kubescheduler.k8s.local
```