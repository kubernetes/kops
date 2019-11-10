# Labels

There are two main types of labels that kops can create:

* `cloudLabels` become tags in AWS on the instances
* `nodeLabels` become labels on the k8s Node objects

## cloudLabels

cloudLabels can be specified at the cluster and instance group level.

Labels specified at the cluster level will be copied to Load Balancers and to Master EBS volumes. They will also be merged to instance groups cloudLabels, with labels specified at the instance group overriding the cluster level.

A nice use for cloudLabels is to specify [AWS cost allocation tags](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html).

An example:

`kops edit ig nodes`

```
...
spec: 
  cloudLabels:
    team: me
    project: ion
...
```

`kops edit cluster`

```
...
spec: 
  cloudLabels:
    team: me
    project: ion
...
```

## nodeLabels

nodeLabels are specified at the instance group.

A good use for nodeLabels is to implement [nodeSelector labels](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#step-two-add-a-nodeselector-field-to-your-pod-configuration).
cloudLabels and nodeLabels are automatically applied to new nodes created by [AWS EC2 auto scaling groups](https://aws.amazon.com/autoscaling/).

An example:

`kops edit ig nodes`

```
...
spec: 
  nodeLabels:
    spot: "false"
...
```

Note that keys and values are strings, so you need quotes around values that YAML would otherwise treat as numbers or booleans.

## Applying Label Updates

To apply changes, you'll need to do a `kops update cluster` and then likely a `kops rolling-update cluster`

For AWS if `kops rolling-update cluster --instance-group nodes` returns "No rolling-update required." the [kops rolling-update cluster](cli/kops_rolling-update_cluster.md) `--force` flag can be used to force a rolling update, even when no changes are identified.

Example:

`kops rolling-update cluster --instance-group nodes --force`
