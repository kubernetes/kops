High Availability (HA)
======================

Introduction
-------------

Kubernetes has two strategies for high availability:

* Run multiple independent clusters and combine them behind one management plane: [federation](https://kubernetes.io/docs/user-guide/federation/)
* Run a single cluster in multiple cloud zones, with redundant components

kops has good support for a cluster than runs
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


Using Kops HA
-------------

We can create HA clusters using kops, but only it's important to note that migrating from a single-master
cluster to a multi-master cluster is a complicated operation (described [here](../single-to-multi-master.md)).
If possible, try to plan this at time of cluster creation.

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


Advanced Example
----------------

Another example `create cluster` invocation for HA with [a private network topology](../topology.md):

```
kops create cluster \
    --node-count 3 \
    --zones us-west-2a,us-west-2b,us-west-2c \
    --master-zones us-west-2a,us-west-2b,us-west-2c \
    --dns-zone example.com \
    --node-size t2.medium \
    --master-size t2.medium \
    --node-security-groups sg-12345678 \
    --master-security-groups sg-12345678,i-abcd1234 \
    --topology private \
    --networking weave \
    --cloud-labels "Team=Dev,Owner=John Doe" \
    --image 293135079892/k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-11-16 \
    ${NAME}
```

Notes (Best Practice)
----
* In regions with 2 Availability Zones, deploy the 3 masters in one zone and the nodes can be distributed between the 2
zones. This can be done by specifying the flags:
```
     --master-count=3
     --master-zones=$MASTER_ZONE
     --zones=$NODE_ZONES
```
