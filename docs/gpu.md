# GPU Support

## kOps managed device driver

{{ kops_feature_table(kops_added_default='1.22') }}

kOps can install nvidia device drivers, plugin, and runtime, as well as configure containerd to make use of the runtime.

kOps will also install a RuntimeClass `nvidia`. As the nvidia runtime is not the default runtime, you will need to add `runtimeClassName: nvidia` to any Pod spec you want to use for GPU workloads. The RuntimeClass also configures the appropriate node selectors and tolerations to run on GPU Nodes.

kOps will add `kops.k8s.io/gpu="1"` as node selector as well as the following taint:

```yaml
  taints:
  - effect: NoSchedule
    key: nvidia.com/gpu
```

The taint will prevent you from accidentially scheduling workloads on GPU Nodes.

You can enable nvidia by adding the following to your Cluster spec:

```yaml
  containerd:
    nvidiaGPU:
      enabled: true
```

## Creating an instance group with GPU nodeN

Due to the cost of GPU instances you want to minimize the amount of pods running on them. Therefore start by provisioning a regular cluster following the [getting started documentation](https://kops.sigs.k8s.io/getting_started/aws/).

Once the cluster is running, add an instance group with GPUs:

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: <cluster name>
  name: gpu-nodes
spec:
  image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20200907
  nodeLabels:
    kops.k8s.io/instancegroup: gpu-nodes
  machineType: g4dn.xlarge
  maxSize: 1
  minSize: 1
  role: Node
  subnets:
  - eu-central-1c
```