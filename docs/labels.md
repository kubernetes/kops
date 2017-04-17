# Labels

There are two main types of labels that kops can create:

* `CloudLabels` become tags in AWS on the instances
* `NodeLabels` become labels on the k8s Node objects

Both are specified at the InstanceGroup level.

A nice use for CloudLabels is to specify [AWS cost allocation tags](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html).

A good use for nodeLables is to implement [nodeSelector labels](https://kubernetes
.io/docs/concepts/configuration/assign-pod-node/#step-two-add-a-nodeselector-field-to-your-pod-configuration) that 
survive [AWS EC2 auto scaling groups](https://aws.amazon.com/autoscaling/) replacing unhealthy or unreachable instances.

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

Note for AWS if `kops rolling-update cluster --instance-group nodes` returns "No rolling-update required." you'll need 
to manually terminate the EC2 node for the auto scaling group to propagate the new labels. 