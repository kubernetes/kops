## Installing Kubernetes Addons

With kops you manage addons by using kubectl.

(For a description of the addon-manager, please see [addon_manager.md](addon_manager.md).)

Addons in Kubernetes are traditionally done by copying files to `/etc/kubernetes/addons` on the master.  But this
doesn't really make sense in HA master configurations.  We also have kubectl available, and addons are just a thin
wrapper over calling kubectl.

The command `kops create cluster` does not support specifying addons to be added to the cluster when it is created. Instead they can be added after cluster creation using kubectl. Alternatively when creating a cluster from a yaml manifest, addons can be specified using `spec.addons`.
```yaml
spec:
  addons:
  - manifest: kubernetes-dashboard
  - manifest: s3://kops-addons/addon.yaml
```

This document describes how to install some common addons and how to create your own custom ones. 

### Custom addons

The docs about the [addon manager](addon_manager.md) describe in more detail how to define a addon resource with regards to versioning.
Here is a minimal example of an addon manifest that would install two different addons.

```yaml
kind: Addons
metadata:
  name: example
spec:
  addons:
  - name: foo.addons.org.io
    version: 0.0.1
    selector:
      k8s-addon: foo.addons.org.io
    manifest: foo.addons.org.io/v0.0.1.yaml
  - name: bar.addons.org.io
    version: 0.0.1
    selector:
      k8s-addon: bar.addons.org.io
    manifest: bar.addons.org.io/v0.0.1.yaml
```

In this this example the folder structure should look like this;

```
addon.yaml
  foo.addons.org.io
    v0.0.1.yaml
  bar.addons.org.io
    v0.0.1.yaml
```

The yaml files in the foo/bar folders can be any kubernetes resource. Typically this file structure would be pushed to S3 or another of the supported backends and then referenced as above in `spec.addons`. In order for master nodes to be able to access the S3 bucket containing the addon manifests, one might have to add additional iam policies to the master nodes using `spec.additionalPolicies`, like so;
```yaml
spec:
  additionalPolicies:
    master: |
      [
        {
          "Effect": "Allow",
          "Action": [
            "s3:GetObject"
          ],
          "Resource": ["arn:aws:s3:::kops-addons/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "s3:GetBucketLocation",
            "s3:ListBucket"
          ],
          "Resource": ["arn:aws:s3:::kops-addons"]
        }
      ]
```
The masters will poll for changes in the bucket and keep the addons up to date.


### Dashboard

The [dashboard project](https://github.com/kubernetes/dashboard) provides a nice administrative UI:

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/kubernetes-dashboard/v1.10.1.yaml
```

And then follow the instructions in the [dashboard documentation](https://github.com/kubernetes/dashboard/wiki/Accessing-Dashboard---1.7.X-and-above) to access the dashboard.

The login credentials are:

* Username: `admin`
* Password: get by running `kops get secrets kube --type secret -oplaintext` or `kubectl config view --minify`

#### RBAC

For k8s version > 1.6 and [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) enabled it's necessary to add your own permission to the dashboard. Please read the [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) docs before applying permissions. 

Below you see an example giving **full access** to the dashboard.

```
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubernetes-dashboard
  labels:
    k8s-app: kubernetes-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: kubernetes-dashboard
  namespace: kube-system
  ```

### Monitoring with Heapster - Standalone

Monitoring supports the horizontal pod autoscaler.

Install using:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/monitoring-standalone/v1.11.0.yaml
```
Please note that [heapster is retired](https://github.com/kubernetes/heapster/blob/master/docs/deprecation.md). Consider using [metrics-server](https://github.com/kubernetes-incubator/metrics-server) and a third party metrics pipeline to gather Prometheus-format metrics instead.

### Monitoring with Prometheus Operator + kube-prometheus

The [Prometheus Operator](https://github.com/coreos/prometheus-operator/) makes the Prometheus configuration Kubernetes native and manages and operates Prometheus and Alertmanager clusters. It is a piece of the puzzle regarding full end-to-end monitoring.

[kube-prometheus](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus) combines the Prometheus Operator with a collection of manifests to help getting started with monitoring Kubernetes itself and applications running on top of it.

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/prometheus-operator/v0.26.0.yaml
```

### Route53 Mapper

Please note that kops installs a Route53 DNS controller automatically (it is required for cluster discovery).
The functionality of the route53-mapper overlaps with the dns-controller, but some users will prefer to
use one or the other.
[README for the included dns-controller](https://github.com/kubernetes/kops/blob/master/dns-controller/README.md)

route53-mapper automates creation and updating of entries on Route53 with `A` records pointing
to ELB-backed `LoadBalancer` services created by Kubernetes. Install using:

The project is created by wearemolecule, and maintained at
[wearemolecule/route53-kubernetes](https://github.com/wearemolecule/route53-kubernetes).
[Usage instructions](https://github.com/kubernetes/kops/blob/master/addons/route53-mapper/README.md)

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/route53-mapper/v1.3.0.yml
```
