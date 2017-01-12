## Allocations of CPU on the master

Note these are only _requests_, not limits.

```
50m  dns-controller
150m etcd main
100m etcd events
150m kube-apiserver
100m kube-controller-manager
100m kube-proxy
100m kube-scheduler

====

750m total

(leaving 250m for misc services e.g. CNI controller, log infrastructure etc)
```
