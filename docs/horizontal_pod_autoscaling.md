# Horizontal Pod Autoscaling

With Horizontal Pod Autoscaling, Kubernetes automatically scales the number of
pods in a replication controller, deployment or replica set based on observed
CPU utilization (or, with alpha support, on some other, application-provided
metrics).

The current stable version, which only includes support for CPU autoscaling, can
be found in the `autoscaling/v1` API version. The alpha version, which includes
support for scaling on memory and custom metrics, can be found in
`autoscaling/v2alpha1` in 1.7 and `autoscaling/v2beta1` 1.8 and 1.9.

Kops can assist in setting up HPA and recommends Kubernetes `1.7.x` to `1.9.x`
and Kops `>=1.7`. Relevant reading you will need to go through:

* [Extending the Kubernetes API with the aggregation layer][k8s-extend-api]
* [Configure The Aggregation Layer][k8s-aggregation-layer]
* [Horizontal Pod Autoscaling][k8s-hpa]

While the above links go into details on how Kubernetes needs to be configured
to work with HPA, a lot of that work is already done for you by Kops.
Specifically:

* [x] Enable the [Aggregation Layer][k8s-aggregation-layer] via the following
  kube-apiserver flags:
   * [x] `--requestheader-client-ca-file=<path to aggregator CA cert>`
   * [x] `--requestheader-allowed-names=aggregator`
   * [x] `--requestheader-extra-headers-prefix=X-Remote-Extra-`
   * [x] `--requestheader-group-headers=X-Remote-Group`
   * [x] `--requestheader-username-headers=X-Remote-User`
   * [x] `--proxy-client-cert-file=<path to aggregator proxy cert>`
   * [x] `--proxy-client-key-file=<path to aggregator proxy key>`
* [x] Enable [Horizontal Pod Scaling][k8s-hpa] ... set the appropriate flags for
  `kube-controller-manager`:
   * [x] `--horizontal-pod-autoscaler-use-rest-clients` should be true.
   * [x] `--kubeconfig <path-to-kubeconfig>`

Ensure that you use **1.8.5 at minimum for Kubernetes 1.8.x** due to [a nasty
bug affecting the API aggregation layer][issue-1]. It was fixed in
[kubernetes/kubernetes#55259][pr-5].

## Cluster Configuration

### Support For Multiple Metrics

Enable API versions required to support scaling on cpu, memory and custom
metrics:

```yaml
# On K8s 1.7
spec:
  kubeAPIServer:
    runtimeConfig:
      autoscaling/v2alpha1: "true"
```

```yaml
# On K8s 1.8 and 1.9
spec:
  kubeAPIServer:
    runtimeConfig:
      autoscaling/v2beta1: "true"
```

If you've set the above configuration, your cluster is now ready for the
resource metrics API ([installation instruction here][k8s-metrics-server]). The
compatibility matrix is as follows:

Metrics Server | Metrics API group/version | Supported Kubernetes version
---------------|---------------------------|-----------------------------
0.3.x          | `metrics.k8s.io/v1beta1`  | 1.8+
0.2.x          | `metrics.k8s.io/v1beta1`  | 1.8+
0.1.x          | `metrics/v1alpha1`        | 1.7

### Support For Custom Metrics

Enable the horizontal pod autoscaler REST client:

```yaml
spec:
  kubeControllerManager:
    horizontalPodAutoscalerUseRestClients: true
```

If you've set the above configuration, your cluster is now ready for the custom
metrics API. Register it via the API aggregation layer. If you're using
Prometheus, checkout the [custom metrics adapter for Prometheus][k8s-prometheus-custom-metrics-adapter].

## Implementation Details

These are the PRs that enable the required configuration:

* [kubernetes/kops#3679][pr-1] - sets `--requestheader-xxx` kube-apiserver flags
  required to enable aggregation layer
  ```
  --requestheader-client-ca-file=<path to aggregator CA cert>
  --requestheader-allowed-names=aggregator
  --requestheader-extra-headers-prefix=X-Remote-Extra-
  --requestheader-group-headers=X-Remote-Group
  --requestheader-username-headers=X-Remote-User
  ```
* [kubernetes/kops#3165][pr-2] - sets `--proxy-client-xxx` kube-apiserver flags
  required to enable aggregation layer
  ```
  --proxy-client-cert-file=<path to aggregator proxy cert>
  --proxy-client-key-file=<path to aggregator proxy key>
  ```
* [kubernetes/kops#3939][pr-3] - add config option to set `--horizontal-pod-
  autoscaler-use-rest-clients` kube-controller-manager flag required to enable
  custom metrics
* [kubernetes/kops#1574][pr-4] - add config options to set `--enable-custom-
  metrics` flag on master and node kubelets required to enable custom metrics

[k8s-aggregation-layer]: https://v1-9.docs.kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/
[k8s-extend-api]: https://v1-9.docs.kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/
[k8s-hpa]: https://v1-9.docs.kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[k8s-metrics-server]: https://github.com/kubernetes/kops/blob/master/addons/metrics-server/README.md
[k8s-prometheus-custom-metrics-adapter]: https://github.com/DirectXMan12/k8s-prometheus-adapter

[pr-1]: https://github.com/kubernetes/kops/pull/3679
[pr-2]: https://github.com/kubernetes/kops/pull/3165
[pr-3]: https://github.com/kubernetes/kops/pull/3939
[pr-4]: https://github.com/kubernetes/kops/pull/1574
[pr-5]: https://github.com/kubernetes/kubernetes/pull/55259

[issue-1]: https://github.com/kubernetes/kubernetes/issues/55022
