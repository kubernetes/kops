# Prometheus Operator Addon

[Prometheus Operator](https://coreos.com/operators/prometheus) creates/configures/manages Prometheus clusters atop Kubernetes. This addon deploy prometheus-operator and [kube-prometheus](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus/README.md) in a kops cluster.

## Usage

### Deploy To Cluster

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/prometheus-operator/v0.19.0.yaml
```
### Updating the addon

Run the script bellow.

```console
addons/prometheus-operator/sync-repo.sh "v0.19.0"
```