# Horizontal Pod Autoscaling

With Horizontal Pod Autoscaling, Kubernetes automatically scales the number of
pods in a replication controller, deployment, or replica set based on observed
CPU utilization (or, with alpha support, on some other, application-provided
metrics).

The current stable version, which only includes support for CPU autoscaling, can
be found in the `autoscaling/v1` API version. The alpha version, which includes
support for scaling on memory and custom metrics, can be found in
`autoscaling/v2beta1` (and `autoscaling/v2beta2` in 1.12 and later).

Kops can assist in setting up HPA. Relevant reading you will need to go through:

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

## Cluster Configuration

### Support For Multiple Metrics

To enable the resource metrics API for scaling on CPU and memory, install metrics-server
([installation instruction here][k8s-metrics-server]). The
compatibility matrix is as follows:

Metrics Server | Metrics API group/version | Supported Kubernetes version
---------------|---------------------------|-----------------------------
0.3.x          | `metrics.k8s.io/v1beta1`  | 1.8+

### Support For Custom Metrics

To enable the custom metrics API, register it via the API aggregation layer. If you're using
Prometheus, checkout the [custom metrics adapter for Prometheus][k8s-prometheus-custom-metrics-adapter].

[k8s-aggregation-layer]: https://kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/
[k8s-extend-api]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/
[k8s-hpa]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[k8s-metrics-server]: https://github.com/kubernetes/kops/blob/master/addons/metrics-server/README.md
[k8s-prometheus-custom-metrics-adapter]: https://github.com/DirectXMan12/k8s-prometheus-adapter
