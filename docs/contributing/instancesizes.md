## Allocations of CPU on the master

Note these are only _requests_, not limits.

```
50m  dns-controller
200m etcd main
100m etcd events
150m kube-apiserver
100m kube-controller-manager
100m kube-proxy
100m kube-scheduler

====

800m total
```

* One a 1 core master, this leaves 200m for misc services e.g. CNI controller, log infrastructure etc.  That will be
less if we start reserving capacity on the master.

* kube-dns is relatively CPU hungry, and runs on the nodes.

* We restrict CNI controllers to 100m.  If a controller needs more, it can support a user-settable option.

* Setting a resource limit is a bad idea: https://github.com/kubernetes/kubernetes/issues/51135
