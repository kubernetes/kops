# Labels

There are two main types of labels that kops can create:

* `CloudLabels` become tags in AWS on the instances
* `NodeLabels` become labels on the k8s Node objects

Both are specified at the InstanceGroup level.

A nice use for CloudLabels is to specify [AWS cost allocation tags](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html)

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
