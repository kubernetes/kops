# Kubernetes Metrics Server

Compatibility matrix:

Metrics Server | Metrics API group/version | Supported Kubernetes version
---------------|---------------------------|-----------------------------
0.2.x          | `metrics.k8s.io/v1beta1`  | 1.8+
0.1.x          | `metrics/v1alpha1`        | 1.7


In order to deploy metrics-server in your cluster run the following command from
the top-level directory of this repository:

```console
# Kubernetes 1.7
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/metrics-server/v1.7.x.yaml

# Kubernetes 1.8+
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/metrics-server/v1.8.x.yaml
```
