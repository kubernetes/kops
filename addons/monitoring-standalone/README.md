# Monitoring Standalone Addon

Heapster provides basic cluster monitoring and is used for running HorizontalPodAutoscalers and integrating metrics into the kubernetes dashboard.

## Usage

### Deploy To Cluster

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/monitoring-standalone/v1.6.0.yaml
```
