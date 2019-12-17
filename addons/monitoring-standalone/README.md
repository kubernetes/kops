# Monitoring Standalone Addon

***RETIRED***:Heapster provides basic cluster monitoring and is used for running HorizontalPodAutoscalers and integrating metrics into the kubernetes dashboard.
The following are potential migration paths for Heapster functionality:

- **For basic CPU/memory HPA metrics**: Use [metrics-server](https://github.com/kubernetes-incubator/metrics-server).

- **For general monitoring**: Consider a third-party monitoring pipeline that can gather Prometheus-formatted metrics.
  The kubelet exposes all the metrics exported by Heapster in Prometheus format.
  One such monitoring pipeline can be set up using the [Prometheus Operator](https://github.com/coreos/prometheus-operator), which
  deploys Prometheus itself for this purpose.

## Usage

### Deploy To Cluster

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/monitoring-standalone/v1.7.0.yaml
```
