# Kubernetes Metrics Server

## User guide

You can find the user guide in
[the official Kubernetes documentation](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-metrics-pipeline/).

## Design

The detailed design of the project can be found in the following docs:

- [Metrics API](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/resource-metrics-api.md)
- [Metrics Server](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/metrics-server.md)

For the broader view of monitoring in Kubernetes take a look into
[Monitoring architecture](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/monitoring_architecture.md)

## Prerequisites
you must allow service account tokens to communicate with kubelet, edit your cluster configuration
```console
$ kops edit cluster
```

add configuration below to your cluster configuration.
```
kubelet:
    anonymousAuth: false
    authorizationMode: Webhook
    authenticationTokenWebhook: true
```

update your cluster
```console
$ kops update cluster --yes
$ kops rolling-update cluster --yes
```

## Deployment

Compatibility matrix:

Metrics Server | Metrics API group/version | Supported Kubernetes version
---------------|---------------------------|-----------------------------
0.3.x          | `metrics.k8s.io/v1beta1`  | 1.8+
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

## Flags

Metrics Server supports all the standard Kubernetes API server flags, as
well as the standard Kubernetes `glog` logging flags.  The most
commonly-used ones are:

- `--logtostderr`: log to standard error instead of files in the
  container.  You generally want this on.

- `--v=<X>`: set log verbosity.  It's generally a good idea to run a log
  level 1 or 2 unless you're encountering errors.  At log level 10, large
  amounts of diagnostic information will be reported, include API request
  and response bodies, and raw metric results from Kubelet.

- `--secure-port=<port>`: set the secure port.  If you're not running as
  root, you'll want to set this to something other than the default (port
  443).

- `--tls-cert-file`, `--tls-private-key-file`: the serving certificate and
  key files.  If not specified, self-signed certificates will be
  generated, but it's recommended that you use non-self-signed
  certificates in production.

Additionally, Metrics Server defines a number of flags for configuring its
behavior:

- `--metric-resolution=<duration>`: the interval at which metrics will be
  scraped from Kubelets (defaults to 60s).

- `--kubelet-insecure-tls`: skip verifying Kubelet CA certificates.  Not
  recommended for production usage, but can be useful in test clusters
  with self-signed Kubelet serving certificates.

- `--kubelet-port`: the port to use to connect to the Kubelet (defaults to
  the default secure Kubelet port, 10250).

- `--kubelet-preferred-address-types`: the order in which to consider
  different Kubelet node address types when connecting to Kubelet.
  Functions similarly to the flag of the same name on the API server.