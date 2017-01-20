# High Availability (HA)

## Introduction

Kubernetes has two strategies for high availability:

* Run multiple independent clusters and combine them behind one management plane: [federation](https://kubernetes.io/docs/user-guide/federation/)
* Run a single cluster in multiple cloud zones, with redundant components

kops has experimental early support for federation, but it already has good support for a cluster than runs
with redundant components.  kops is able to create multiple kubernetes masters, so in the event of
a master instance failure, the kubernetes API will continue to operate.

However, when running kubernetes with a single master, if the master fails, the kubernetes API will be unavailable, but pods and services that are running on the (unaffected) nodes should continue to operate.  In this situation, we won't be able to do anything that involves the API (adding nodes, scaling pods, replacing
terminated pods), and kubectl won't work.  However your application should continue to run, and most applications
could probably tolerate an API outage of an hour or more.

Moreover, kops runs the masters in an automatic replacement mode.  Masters are run in auto-scaling groups, with
the data on an EBS volume.  If a master node is terminated, the ASG will launch a new master instance, and kops
will mount the master volume and replace the master.

In short:

* A single master kops cluster is still reasonably available; if the master instance terminates it will be automatically
  replaced.  But the use of EBS binds us to a single AZ, and in the event of a prolonged AZ outage, we might experience
  downtime.
* A multi-node kops cluster can tolerate the outage of a single AZ
* Federation will allow you to create "uber-clusters" that can tolerate a regional outage

## Using Kops HA

We can create HA clusters using kops, but only it's important to note that you must plan for this at time of cluster creation.  Currently it is not possible to change
the etcd cluster size (i.e. we cannot change an HA cluster to be non-HA, or a non-HA cluster to be HA.) [Issue #1512](https://github.com/kubernetes/kops/issues/1512)

When you first call `kops create cluster`, you specify the `--master-zones` flag listing the zones you want your masters
to run in, for example:

```
kops create cluster \
    --node-count 3 \
    --zones us-west-2a,us-west-2b,us-west-2c \
    --master-zones us-west-2a,us-west-2b,us-west-2c \
    --node-size t2.medium \
    --master-size t2.medium \
    --topology private \
    --networking kopeio-vxlan \
    hacluster.example.com
```

Kubernetes relies on a key-value store called "etcd", which uses the Quorum approach to consistency,
so it is available if 51% of the nodes are available.

As a result there are a few considerations that need to be taken into account when using kops with HA:

* Only odd number of masters instances should be created, as an even number is likely _less_ reliable than the lower odd number.
* Kops has experimental support for running multiple masters in the same AZ, but it should be used carefully.
  If we create 2 (or more) masters in the same AZ, then failure of the AZ will likely cause etcd to lose quorum
  and stop operating (with 3 nodes).  Running in the same AZ therefore increases the risk of cluster disruption,
  though it can be a valid scenario, particularly if combined with [federation](https://kubernetes.io/docs/user-guide/federation/).
