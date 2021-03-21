# Scaling kOps

## Scaling the control plane

### Dedicated API Server nodes

{{ kops_feature_table(kops_added_default='1.21') }}

A common bottleneck of the control plane is the API server. As the number of pods and nodes grow, you will want to add more resources to handle the load.

You can scale the API server horizontally by adding instance groups dedicated to running API server nodes. You can do so by adding an instance group with the `APIServer` role:

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: null
  labels:
    kops.k8s.io/cluster: <cluster name>
  name: apiserver-eu-central-1a
spec:
  image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20201026
  machineType: t3.small
  maxSize: 3
  minSize: 3
  nodeLabels:
    kops.k8s.io/instancegroup: apiserver-eu-central-1a
  role: APIServer
  subnets:
  - eu-central-1a

```

or run `kops create ig --name=<cluster name> apiserver-eu-central-1a --subnet=eu-central-1a`

Because the labels, taints, and domains can change, this feature is currently behind a feature gate.
```sh
export KOPS_FEATURE_FLAGS="+APIServerNodes"
```