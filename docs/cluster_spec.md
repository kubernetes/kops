# Description of Keys in `config` and `cluster.spec`

This list is not complete, but aims to document any keys that are less than self-explanatory.

## spec

### adminAccess

This array configures the CIDRs that are able to ssh into nodes. On AWS this is manifested as inbound security group rules on the `nodes` and `master` security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  adminAccess:
    - 12.34.56.78/32
```

### kubeAPIServer

This block contains configuration for the `kube-apiserver`.

#### runtimeConfig

Keys and values here are translated into `--runtime-config` values for `kube-apiserver`, separated by commas.

Use this to enable alpha features, for example:

```yaml
spec:
  kubeAPIServer:
    runtimeConfig:
      batch/v2alpha1: "true"
      apps/v1alpha1: "true"
```

Will result in the flag `--runtime-config=batch/v2alpha1=true,apps/v1alpha1=true`. Note that `kube-apiserver` accepts `true` as a value for switch-like flags.

### networkID

On AWS, this is the id of the VPC the cluster is created in. If creating a cluster from scratch, this field doesn't need to be specified at create time; `kops` will create a `VPC` for you.

```yaml
spec:
  networkID: vpc-abcdefg1
```
