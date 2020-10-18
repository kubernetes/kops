# GPU Support

You can use [GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/overview.html) to install NVIDIA device drivers and tools to your cluster.

## Creating a cluster with GPU nodes

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
  taints:
  - nvidia.com/gpu=present:NoSchedule
```

Note the taint used above. This will prevent pods from being scheduled on GPU nodes unless we explicitly want to. The GPU Operator resources tolerate this taint by default.
Also note the node label we set. This will be used to ensure the GPU Operator resources runs on GPU nodes. 

## Install GPU Operator
GPU Operator is installed using `helm`. See the [general install instructions for GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/getting-started.html#install-gpu-operator).

In order to match the _kops_ environment, create a `values.yaml` file with the following content:

```yaml
operator:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes

driver:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes

toolkit:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes

devicePlugin:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes

dcgmExporter:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes

gfd:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes

node-feature-discovery:
  worker:
    nodeSelector:
      kops.k8s.io/instancegroup: gpu-nodes
```

Once you have installed the the _helm chart_ you should be able to see the GPU operator resources being spawned in the `gpu-operator-resources` namespace.

You should now be able to schedule other workloads on the GPU by adding the following properties to the pod spec:
```yaml
spec:
  nodeSelector:
    kops.k8s.io/instancegroup: gpu-nodes
  tolerations:
  - key: nvidia.com/gpu
    operator: Exists
```