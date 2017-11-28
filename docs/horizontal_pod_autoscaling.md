# Horizontal Pod Autoscaling

With Horizontal Pod Autoscaling, Kubernetes automatically scales the number of
pods in a replication controller, deployment or replica set based on observed
CPU utilization (or, with alpha support, on some other, application-provided
metrics).

Kops can assist in setting up HPA and recommends Kubernetes `>= 1.7` and Kops
`>=1.8`. Relevant reading you will need to go through:

* [Configure The Aggregation Layer][1]
* [Horizontal Pod Autoscaling][5]

## Required API Versions

The current stable version, which only includes support for CPU autoscaling, can
be found in the `autoscaling/v1` API version.

## Extra Capabilites

### Support For Multiple Metrics

The alpha version, which includes support for scaling on memory and custom
metrics, can be found in `autoscaling/v2alpha1` or `autoscaling/v2beta1`. Use
these if you want to specify multiple metrics for the Horizontal Pod Autoscaler
to scale on:

```yaml
# On K8s 1.7
spec:
  kubeAPIServer:
    runtimeConfig:
      autoscaling/v2alpha1: "true"
```

```yaml
# On K8s 1.8
spec:
  kubeAPIServer:
    runtimeConfig:
      autoscaling/v2beta1: "true"
```

### Support For Custom Metrics

Enable gathering custom metrics:

```yaml
spec:
  kubelet:
    enableCustomMetrics: true
```

Enable the horizontal pod autoscaler REST client:

```yaml
spec:
  kubeControllerManager:
    horizontalPodAutoscalerUseRestClients: true
```

## Relevant PRs

* [kubernetes/kops#3679][2] - set `--requestheader-xxx` kube-apiserver flags required to enable aggregation layer
  ```
  --requestheader-client-ca-file=<path to aggregator CA cert>
  --requestheader-allowed-names=aggregator
  --requestheader-extra-headers-prefix=X-Remote-Extra-
  --requestheader-group-headers=X-Remote-Group
  --requestheader-username-headers=X-Remote-User
  ```
* [kubernetes/kops#3165][3] - sets `--proxy-client-xxx` kube-apiserver flags required to enable aggregation layer
  ```
  --proxy-client-cert-file=<path to aggregator proxy cert>
  --proxy-client-key-file=<path to aggregator proxy key>
  ```
* [kubernetes/kops#3679][4] - add config option to set `--horizontal-pod-autoscaler-use-rest-clients` kube-controller-manager flag required to enable custom metrics
* [kubernetes/kops#1574][6] - add config options to set `--enable-custom-metrics` flag on master and node kubelets required to enable custom metrics

[1]: https://kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/
[2]: https://github.com/kubernetes/kops/pull/3679
[3]: https://github.com/kubernetes/kops/pull/3165
[4]: https://github.com/kubernetes/kops/pull/3939
[5]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[6]: https://github.com/kubernetes/kops/pull/1574
