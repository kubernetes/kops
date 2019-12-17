# Prometheus Operator Addon

[Prometheus Operator](https://coreos.com/operators/prometheus) creates/configures/manages Prometheus clusters atop Kubernetes. This addon deploy prometheus-operator and [kube-prometheus](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus/README.md) in a kops cluster.

## Prerequisites

Version `>=0.18.0` of the Prometheus Operator requires a Kubernetes
cluster of version `>=1.8.0`. If you are just starting out with the
Prometheus Operator, it is highly recommended to use the latest version.

If you have an older version of Kubernetes and the Prometheus Operator running,
we recommend upgrading Kubernetes first and then the Prometheus Operator.

## Usage

### Deploy To Cluster

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/prometheus-operator/v0.26.0.yaml
```
### Updating the addon

Run the script below.

```console
addons/prometheus-operator/sync-repo.sh "v0.26.0"
```
