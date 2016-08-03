## Installing addons with kops

With kops you manage addons by using kubectl.

Addons in kubernetes are traditionally done by copying files to `/etc/kubernetes/addons` on the master.  But this
doesn't really make sense in HA master configurations.  We also have kubectl available, and addons is just a thin
wrapper over calling kubectl.

This document describes how to install some common addons.

### Dashboard

The (dashboard project)[https://github.com/kubernetes/dashboard] provides a nice administrative UI:

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/dashboard/v1.1.0.yaml
```

And then navigate to `https://<clustername>/ui`

(`/ui` is an alias to `https://<clustername>/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard`)

The login credentials are:

* Username: `admin`
* Password: get from `kops secrets expose --id kube --type secret`


### Monitoring - Standalone

Monitoring supports the horizontal pod autoscaler.

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/monitoring-standalone/v1.1.0.yaml
```
