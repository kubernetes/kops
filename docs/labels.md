# Labels

There are two main types of labels that kops can create:

* `CloudLabels` become tags in AWS on the instances
* `NodeLabels` become labels on the k8s Node objects

Both are specified at the InstanceGroup level.

A nice use for cloudLabels is to specify [AWS cost allocation tags](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html).

A good use for nodeLables is to implement [nodeSelector labels](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#step-two-add-a-nodeselector-field-to-your-pod-configuration). 
CloudLabels and nodeLabels are automatically applied to new nodes created by [AWS EC2 auto scaling groups](https://aws.amazon.com/autoscaling/).

An example:

`kops edit ig nodes`

```
...
spec: 
  nodeLabels:
    spot: "false"
  cloudLabels:
    team: me
    project: ion
...
```

Note that keys and values are strings, so you need quotes around values that YAML
 would otherwise treat as numbers or booleans.

To apply changes, you'll need to do a `kops update cluster` and then likely a `kops rolling-update cluster`

For AWS if `kops rolling-update cluster --instance-group nodes` returns "No rolling-update required." the 
[kops rolling-update cluster](https://github.com/kubernetes/kops/blob/8bc48ef10a44a3e481b604f5dbb663420c68dcab/docs/cli/kops_rolling-update_cluster.md) `--force` flag can be used to force a rolling update, even when no changes are identified.

Example:

`kops rolling-update cluster --instance-group nodes --force`
