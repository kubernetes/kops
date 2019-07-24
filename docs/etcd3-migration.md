# Moving to etcd3 and/or adopting etcd-manager

- [Moving to etcd3 and/or adopting etcd-manager](#moving-to-etcd3-andor-adopting-etcd-manager)
  - [Background Info](#background-info)
  - [Default upgrades](#default-upgrades)
    - [Non-calico users](#non-calico-users)
    - [Calico users](#calico-users)
  - [Gradual updates](#gradual-updates)
    - [Adopt etcd-manager with kops 1.11 / kubernetes 1.11](#adopt-etcd-manager-with-kops-111--kubernetes-111)
    - [Delay adopting etcd-manager with kops 1.12](#delay-adopting-etcd-manager-with-kops-112)
    - [Delay adopting etcd3 with kops 1.12](#delay-adopting-etcd3-with-kops-112)
  - [Mitigating Workload Downtime](#mitigating-workload-downtime)

## Background Info

Kubernetes is moving from etcd2 to etcd3, which is an upgrade that involves Kubernetes API Server
downtime. Technically there is no usable upgrade path from etcd2 to etcd3 that
supports HA scenarios, but kops has enabled it using etcd-manager.

Nonetheless, this remains a *higher-risk upgrade* than most other kubernetes
upgrades - you are strongly recommended to plan accordingly: back up critical
data, schedule the upgrade during a maintenance window, think about how you
could recover onto a new cluster, try it on non-production clusters first.

To minimize the pain of this migration, we are making some other supporting changes at the same time:

* We enable TLS for both clients & peers with etcd3
* Calico configurations move from talking to etcd directly to using CRDs
  (talking to etcd is considered deprecated)

This does introduce the risk that we are changing more at the same time, and we
provide some mitigation steps for breaking up the upgrade, though most of these
therefore involve multiple disruptive upgrades (e.g. etc2 -> etcd3 is
disruptive, non-TLS to TLS is disruptive).

**Note:** Even if you are already using etcd3 and have TLS enabled, it is
recommended to use to etcd-manager and the steps in this document still apply to
you. If you would like to delay using etcd-manager, there are steps at the
bottom of this doc that outlines how to do that.

## Default upgrades

When upgrading to kubernetes 1.12 with kops 1.12, by default:

* Calico will be updated to a configuration that uses CRDs
* We will automatically start using etcd-manager
* Using etcd-manager will default to etcd3
* Using etcd3 will default to using TLS for all etcd communications

### Non-calico users

The upgrade is therefore disruptive to the masters.  The recommended procedure is to quickly roll your masters, and then do a normal roll of your nodes:

> **DANGER:** Using the procedure to quickly roll your masters can result in downtime for any workloads using Service LoadBalancers. (The "Hammer 🔨" Method)  
> Any time you kill off all three masters with `--cloudonly` and `--master-interval=1s`, you may experience worker nodes go into a `NotReady` state when the new masters come online and reconcile the cluster state. This can lead to Kubernetes Service LoadBalancers removing nodes in a `NotReady` state. In some cases, larger clusters have all nodes in a `NotReady` state, causing a cluster-wide Service LoadBalancer disruption. See [Mitigating Workload Downtime](#mitigating-workload-downtime) for workarounds.

```bash
# Roll masters as quickly as possible
kops rolling-update cluster --cloudonly --instance-group-roles master --master-interval=1s
kops rolling-update cluster --cloudonly --instance-group-roles master --master-interval=1s --yes

# Roll nodes normally
kops rolling-update cluster
kops rolling-update cluster --yes
```


### Calico users

If you are using calico the switch to CRDs will effectively cause a network partition during the rolling-update.  Your application might tolerate this, but it probably won't.  We therefore recommend rolling your nodes as fast as possible also:

> **DANGER:** Using the procedure to quickly roll your masters can result in downtime for any workloads using Service LoadBalancers. (The "Hammer 🔨" Method)  
> Any time you kill off all three masters with `--cloudonly` and `--master-interval=1s`, you may experience worker nodes go into a `NotReady` state when the new masters come online and reconcile the cluster state. This can lead to Kubernetes Service LoadBalancers removing nodes in a `NotReady` state. In some cases, larger clusters have all nodes in a `NotReady` state, causing a cluster-wide Service LoadBalancer disruption. See [Mitigating Workload Downtime](#mitigating-workload-downtime) for workarounds.

```bash
# Roll masters and nodes as quickly as possible
kops rolling-update cluster --cloudonly --master-interval=1s --node-interval=1s
kops rolling-update cluster --cloudonly --master-interval=1s --node-interval=1s --yes
```

## Gradual updates

If you would like to upgrade more gradually, we offer the following strategies
to spread out the disruption over several steps.  Note that they likely involve
more disruption and are not necessarily lower risk.

### Adopt etcd-manager with kops 1.11 / kubernetes 1.11

If you don't already have TLS enabled with etcd, you can adopt etcd-manager before
kops 1.12 & kubernetes 1.12 by running:

```bash
kops set cluster cluster.spec.etcdClusters[*].provider=manager
```

Then you can proceed to update to kops 1.12 & kubernetes 1.12, as this becomes the default.

### Delay adopting etcd-manager with kops 1.12

To delay adopting etcd-manager with kops 1.12, specify the provider as type `legacy`:

```bash
kops set cluster cluster.spec.etcdClusters[*].provider=legacy
```

To remove, `kops edit` your cluster and delete the `provider: Legacy` lines from both etcdCluster blocks.

### Delay adopting etcd3 with kops 1.12

To delay adopting etcd3 with kops 1.12, specify the etcd version as 2.2.1

```bash
kops set cluster cluster.spec.etcdClusters[*].version=2.2.1
```

To remove, `kops edit` your cluster and delete the `version: 2.2.1` lines from both etcdCluster blocks.

## Mitigating Workload Downtime
### AWS ELB Mitigation
When quickly rolling all your masters, you can hit conditions which lead to nodes entering a `NotReady` state. Kubernetes, by default, will remove any `NotReady` nodes from ELBs managed by Services. To avoid possible ELB service interruption, you can add a temporary IAM policy which blocks the masters from removing `NotReady` nodes from LoadBalancer type services. This policy only needs to be in play while you are performing this upgrade and can be removed once the nodes (masters and workers) are all in a `Ready` state. Make sure you remove the policy once the cluster is upgraded and stable, otherwise Kubernetes will not be able to effectively manage your nodes in ELBs.

**Adding the Temporary Policy**
```bash
# Configure your master_node_role_name (Generally "masters.your.cluster.name")
masters_role_name="masters.<your.cluster.name>"

# Install a temporary IAM policy. This avoids nodes being removed from LoadBalancer type services while masters reconcile the state of the cluster.
aws iam put-role-policy \
  --role-name "${masters_role_name}" \
  --policy-name temporary-etcd-upgrade-deny-lb-changes \
  --policy-document \
    '{"Version": "2012-10-17", "Statement": [{"Action": ["elasticloadbalancing:DeregisterInstancesFromLoadBalancer", "elasticloadbalancing:DeregisterTargets"], "Resource": ["*"], "Effect": "Deny"}]}'
```

**Removing the Temporary Policy**
```bash
# Configure your master_node_role_name
masters_role_name="masters.<your.cluster.name>"

# Once your cluster node states have stabilized from `NotReady` to `Ready` you can remove the temporary policy from your master nodes
aws iam delete-role-policy \
  --role-name "${masters_role_name}" \
  --policy-name temporary-etcd-upgrade-deny-lb-changes
```
