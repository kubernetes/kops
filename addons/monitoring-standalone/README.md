# Monitoring Standalone Addon

Heapster provides basic cluster monitoring and is used running `HorizontalPodAutoscalers` and integrating metrics in kubernetes dashboard.

## Usage

### Deploy To Cluster

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/monitoring-standalone/v1.6.0.yaml
```
