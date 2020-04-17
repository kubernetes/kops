# NVIDIA GPU Driver and DevicePlugin Installation

## Summary

This kops hook container may be used to enable nodes with GPUs to work with
Kubernetes.  It is targeted specifically for AWS GPU [instance
types](https://aws.amazon.com/ec2/instance-types/).

It installs the following from web sources.

1. [NVIDIA Device Drivers](http://www.nvidia.com/Download/index.aspx)
2. [CUDA Libraries](https://developer.nvidia.com/cuda-downloads)
3. [nvidia-docker](https://github.com/NVIDIA/nvidia-docker)
4. [docker-ce](https://www.docker.com/community-edition)

Using this hook indicates that you agree to the NVIDIA
[license](http://www.nvidia.com/content/DriverDownload-March2009/licence.php?lang=us).

## How it works

* This kops hook container runs on a kubernetes node upon every boot.
* It installs onto the host system a systemd oneshot service unit
  `nvidia-device-plugin.service` along with setup scripts.
* The systemd unit `nvidia-device-plugin.service` runs and executes the setup
  scripts in the host directory `/nvidia-device-plugin`.
* The scripts install the NVIDIA device drivers, CUDA libs, nvidia-docker along
  with the matching version of docker-ce.
* The scheduling of work in a separate systemd unit outside of this kops hook
  is required because it is not possible to upgrade docker-ce on the host from
  within a docker container.

## Prerequisites

Although this hook *may* work among many combinatorial versions of software and
images, it has only been tested with the following:

* kops: **1.9**, **1.15.0**
* kubernetes: 1.10, **1.11**, **1.15.5**
* OS Image: **`kope.io/k8s-1.10-debian-stretch-amd64-hvm-ebs-2018-05-27`**, **`kope.io/k8s-1.15-debian-stretch-amd64-hvm-ebs-2020-01-17`**
  * This is most certainly not the default image for kops.  The OS image must
    be explicitly overridden in the cluster or instancegroup spec.
  * Debian stretch is needed because `nvidia-docker` requires a newer version
    of `docker-ce >= 18.0`, which is not available in the Debian jessie package
    repository.  In addition, the Debian jessie kernel was compiled with gcc-7,
    while the system packages install gcc-4, thus making the nvidia driver
    compilation fail.
* cloud: **AWS**
  * This hook will only work on AWS at this moment.
  * This is due to the fact that it uses an AWS discovery mechanism to
    determine node instancetype, and subsequently install the correct drivers and
    configure the optimal settings for the GPU chipsets.

### Test Matrix

This kops hook was developed against the following version combinations.

| Kops Version  | Kubernetes Version | GPU Mode     | OS Image |
| ------------- | ------------------ | ------------ | -------- |
| 1.15.0        | 1.15.5             | deviceplugin | kope.io/k8s-1.15-debian-stretch-amd64-hvm-ebs-2020-01-17
| 1.10-beta.1   | 1.10               | deviceplugin | kope.io/k8s-1.10-debian-stretch-amd64-hvm-ebs-2018-05-27
| 1.9.1         | 1.11               | deviceplugin | kope.io/k8s-1.10-debian-stretch-amd64-hvm-ebs-2018-05-27
| 1.9.1         | 1.10               | legacy       | kope.io/k8s-1.10-debian-stretch-amd64-hvm-ebs-2018-05-27

## Using this DevicePlugin

### Create a Cluster with GPU Nodes

```bash
kops create cluster \
  --name gpu.example.k8s.local \
  --zones us-east-1c \
  --node-size p2.xlarge \
  --node-count 1 \
  --image kope.io/k8s-1.15-debian-stretch-amd64-hvm-ebs-2020-01-17 \
  --kubernetes-version 1.15.5
```

### Enable the Kops Installation Hook and DevicePlugins

This should be safe to do for all machines, because the hook auto-detects if
the machine is an AWS GPU instancetype and will NO-OP otherwise. Choose
between the DevicePlugin GPU Mode or Legacy Accelerators GPU Mode.

#### (Preferred) DevicePlugin GPU Mode

This mode is:

* Required for kubernetes >= 1.11.0
* Optional for 1.8.0 =< kubernetes <= 1.11.0

For Kubernetes >= 1.11.0 or clusters supporting DevicePlugins

```yaml
# > kops edit instancegroup nodes
# CUDA 10.0
spec:
  image: kope.io/k8s-1.15-debian-stretch-amd64-hvm-ebs-2020-01-17
  hooks:
  - execContainer:
      image: fifarzh/nvidia-device-plugin:0.2.0-cuda10.0

# CUDA 9.1
spec:
  image: kope.io/k8s-1.10-debian-stretch-amd64-hvm-ebs-2018-05-27
  hooks:
  - execContainer:
      image: dcwangmit01/nvidia-device-plugin:0.1.0

### The settings below are only necessary for kubernetes <= 1.11.0, where
###   deviceplugins are not enabled by default.
# kubelet:
#   featureGates:
#     # Enable DevicePlugins
#     DevicePlugins: "true"
#     # Disable Accelerators (may interfere with DevicePlugins)
#     Accelerators: "false"
```

#### (Deprecated) Legacy Accelerators GPU Mode

The legacy `accelerator`
GPU mode is equivalent to the original [GPU hook](/docs/gpu.md).
Accelerators are deprecated in `Kubernetes >= 1.11.0`.

```yaml
# > kops edit instancegroup nodes

spec:
  image: kope.io/k8s-1.10-debian-stretch-amd64-hvm-ebs-2018-05-27
  hooks:
  - execContainer:
      image: dcwangmit01/nvidia-device-plugin:0.1.0
      environment:
        NVIDIA_DEVICE_PLUGIN_MODE: legacy
  kubelet:
    featureGates:
      # Disable DevicePlugins (may interfere with DevicePlugins)
      DevicePlugins: "false"
      # Enable Accelerators
      Accelerators: "true"
```

### Update the cluster

```bash
kops update cluster gpu.example.k8s.local --yes
kops rolling-update cluster gpu.example.k8s.local --yes
```

### Deploy the Daemonset for the NVIDIA DevicePlugin

Only for DevicePlugin GPU Mode, load the deviceplugin daemonset for your
specific environment. This is not required for the Legacy Accelerators GPU
Mode.

```bash
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/master/nvidia-device-plugin.yml

# (Optional) Set permissive toleration to allow daemonset to run anywhere.
#   By default this is permissive in case you have tainted your GPU nodes.
kubectl patch daemonset nvidia-device-plugin-daemonset --namespace kube-system \
  -p '{ "spec": { "template": { "spec": { "tolerations": [ { "operator": "Exists" } ] } } } }'
```

### Validate that GPUs are Working

#### Deploy a Test Pod

```bash
# CUDA 10.0
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: tf-gpu
spec:
  containers:
  - name: gpu
    image: tensorflow/tensorflow:2.0.1-gpu
    command: [ "/bin/bash", "-ce", "tail -f /dev/null" ]
    # ^ From 2.0.0 onwards, CMD is not set
    imagePullPolicy: IfNotPresent
    resources:
      limits:
        memory: 1024Mi
        nvidia.com/gpu: 1 # requesting 1 GPU
EOF

# CUDA 9.1
cat << EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: tf-gpu
spec:
  containers:
  - name: gpu
    image: tensorflow/tensorflow:1.9.0-gpu
    imagePullPolicy: IfNotPresent
    resources:
      limits:
        memory: 1024Mi
        # ^ Set memory in case default limits are set low
        nvidia.com/gpu: 1 # requesting 1 GPUs
        # ^ For Legacy Accelerators mode this key must be renamed
        #   'alpha.kubernetes.io/nvidia-gpu'
  tolerations:
  # This toleration will allow the gpu hook to run anywhere
  #   By default this is permissive in case you have tainted your GPU nodes.
  - operator: "Exists"
  # ^ If you have a specific taint to target, comment out the above and modify
  #   the example below

### Example tolerations
# - key: "dedicated"
#   operator: "Equal"
#   value: "gpu"
#   effect: "NoExecute"
EOF
```

#### Validate that GPUs are working

```bash
# Check that nodes are detected to have GPUs
kubectl describe nodes | grep -E 'gpu:\s.*[1-9]'

# Check the logs of the Tensorflow Container to ensure that it ran
kubectl logs tf-gpu # no output since 2.0.0

# Show GPU info from within the pod
#   Only works in DevicePlugin mode
kubectl exec -it tf-gpu nvidia-smi

# Show Tensorflow detects GPUs from within the pod.
#   Only works in DevicePlugin mode
kubectl exec -it tf-gpu -- \
  python -c 'from tensorflow.python.client import device_lib; print(device_lib.list_local_devices())'
```

### Teardown the Test Cluster

```yaml
kubectl delete pod/tf-gpu
kubectl delete -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/master/nvidia-device-plugin.yml
kops delete cluster --name gpu.example.k8s.local --yes
```
