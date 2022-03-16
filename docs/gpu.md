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

## GPUs in OpenStack

OpenStack does not support enabling containerd configuration in cluster level. It needs to be done in instance group:

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
  containerd:
    nvidiaGPU:
      enabled: true
```

## Verifying GPUs

1. after new GPU nodes are coming up, you should see them in `kubectl get nodes`
2. nodes should have `kops.k8s.io/gpu` label and `nvidia.com/gpu:NoSchedule` taint
3. `kube-system` namespace should have nvidia-device-plugin-daemonset pod provisioned to GPU node(s)
4. if you see `nvidia.com/gpu` in kubectl describe node <node> everything should work.

```
Capacity:
  cpu:                4
  ephemeral-storage:  9983232Ki
  hugepages-1Gi:      0
  hugepages-2Mi:      0
  memory:             32796292Ki
  nvidia.com/gpu:     1 <- this one
  pods:               110
```
