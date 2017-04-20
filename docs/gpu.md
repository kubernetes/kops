# GPU support

```
kops create cluster gpu.example.com --zones us-east-1c --node-size p2.xlarge --node-count 1 --kubernetes-version 1.6.1
```

(Note that the p2.xlarge instance type is not cheap, but no GPU instances are)

You can use the experimental hooks feature to install the nvidia drivers:

`> kops edit cluster gpu.example.com`
```
spec:
...
  hooks:
  - execContainer:
      image: kopeio/nvidia-bootstrap:1.6
```

(TODO: Only on instance groups, or have nvidia-bootstrap detect if GPUs are present..)

In addition, you will likely want to set the `Accelerators=true` feature-flag to kubelet:

`> kops edit cluster gpu.example.com`
```
spec:
...
  kubelet:
    featureGates:
      Accelerators: "true"
```

`> kops update cluster gpu.example.com --yes`


Here is an example pod that runs tensorflow; note that it mounts libcuda from the host:

(TODO: Is there some way to have a well-known volume or similar?)

```
apiVersion: v1
kind: Pod
metadata:
  name: tf
spec:
  containers:
  - image: gcr.io/tensorflow/tensorflow:1.0.1-gpu
    imagePullPolicy: IfNotPresent
    name: gpu
    command:
    - /bin/bash
    - -c
    - "cp -d /rootfs/usr/lib/x86_64-linux-gnu/libcuda.* /usr/lib/x86_64-linux-gnu/ && cp -d /rootfs/usr/lib/x86_64-linux-gnu/libnvidia* /usr/lib/x86_64-linux-gnu/ &&/run_jupyter.sh"
    resources:
      limits:
        cpu: 2000m
        alpha.kubernetes.io/nvidia-gpu: 1
    volumeMounts:
    - name: rootfs-usr-lib
      mountPath: /rootfs/usr/lib
  volumes:
    - name: rootfs-usr-lib
      hostPath:
        path: /usr/lib
```

To use this particular tensorflow image, you should port-forward and get the URL from the log:

```
kubectl port-forward tf 8888 &
kubectl logs tf
```

And browse to the URL printed
