Changes:

* Switch to _not_ use a NodePort; we access through the kube-api proxy instead
* Add label `k8s-addon: kubernetes-dashboard.addons.k8s.io`