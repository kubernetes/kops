# Calico

[Calico](https://docs.projectcalico.org/latest/introduction/) is an open source networking and
network security solution for containers, virtual machines, and native host-based workloads.

Calico combines flexible networking capabilities with run-anywhere security enforcement to provide
a solution with native Linux kernel performance and true cloud-native scalability. Calico provides
developers and cluster operators with a consistent experience and set of capabilities whether
running in public cloud or on-prem, on a single node or across a multi-thousand node cluster.

## Installing

To use the Calico, specify the following in the cluster spec.

```yaml
  networking:
    calico: {}
```

The following command sets up a cluster using Calico.

```sh
export ZONES=mylistofzones
kops create cluster \
  --zones $ZONES \
  --networking calico \
  --yes \
  --name myclustername.mydns.io
```

## Configuring

### Enable Cross-Subnet mode in Calico (AWS only)

Calico supports a new option for IP-in-IP mode where traffic is only encapsulated
when it’s destined to subnets with intermediate infrastructure lacking Calico route awareness
– for example, across heterogeneous public clouds or on AWS where traffic is crossing availability zones/ regions.

With this mode, IP-in-IP encapsulation is only [performed selectively](https://docs.projectcalico.org/v3.10/networking/vxlan-ipip#configure-ip-in-ip-encapsulation-for-only-cross-subnet-traffic).
This provides better performance in AWS multi-AZ deployments, and in general when deploying on networks where
pools of nodes with L2 connectivity are connected via a router.

Note that Calico by default, routes between nodes within a subnet are distributed using a full node-to-node BGP mesh.
Each node automatically sets up a BGP peering with every other node within the same L2 network.
This full node-to-node mesh per L2 network has its scaling challenges for larger scale deployments.
BGP route reflectors can be used as a replacement to a full mesh, and is useful for scaling up a cluster. [BGP route reflectors are recommended once the number of nodes goes above ~50-100.](https://docs.projectcalico.org/networking/bgp#topologies-for-public-cloud)
The setup of BGP route reflectors is currently out of the scope of kops.

Read more here: [BGP route reflectors](https://docs.projectcalico.org/reference/architecture/overview#bgp-route-reflector-bird)

To enable this mode in a cluster, add the following to the cluster spec:

```yaml
  networking:
    calico:
      crossSubnet: true
```

In the case of AWS, EC2 instances have source/destination checks enabled by default.
When you enable cross-subnet mode in kops, an addon controller ([k8s-ec2-srcdst](https://github.com/ottoyiu/k8s-ec2-srcdst))
will be deployed as a Pod (which will be scheduled on one of the masters) to facilitate the disabling of said source/destination address checks.
Only the masters have the IAM policy (`ec2:*`) to allow k8s-ec2-srcdst to execute `ec2:ModifyInstanceAttribute`.

### Configuring Calico MTU

The Calico MTU is configurable by editing the cluster and setting `mtu` option in the calico configuration.
AWS VPCs support jumbo frames, so on cluster creation kops sets the calico MTU to 8912 bytes (9001 minus overhead).

For more details on Calico MTU please see the [Calico Docs](https://docs.projectcalico.org/networking/mtu#determine-mtu-size).

```yaml
spec:
  networking:
    calico:
      mtu: 8912
```

### Configuring Calico to use Typha

As of Kops 1.12 Calico uses the kube-apiserver as its datastore. The default setup does not make use of [Typha](https://github.com/projectcalico/typha) - a component intended to lower the impact of Calico on the k8s APIServer which is recommended in [clusters over 50 nodes](https://docs.projectcalico.org/latest/getting-started/kubernetes/installation/calico#installing-with-the-kubernetes-api-datastoremore-than-50-nodes) and is strongly recommended in clusters of 100+ nodes.
It is possible to configure Calico to use Typha by editing a cluster and adding a `typhaReplicas` option to the Calico spec:

```yaml
  networking:
    calico:
      typhaReplicas: 3
```

## Getting help

For help with Calico or to report any issues:
* [Calico Github](https://github.com/projectcalico/calico)
* [Calico Users Slack](https://calicousers.slack.com)

For more general information on options available with Calico see the official [Calico docs](https://docs.projectcalico.org/latest/introduction/):
* See [Calico Network Policy](https://docs.projectcalico.org/latest/security/calico-network-policy)
  for details on the additional features not available with Kubernetes Network Policy.
* See [Determining best Calico networking option](https://docs.projectcalico.org/latest/networking/determine-best-networking)
  for help with the network options available with Calico.



## Troubleshooting

### New nodes are taking minutes for syncing ip routes and new pods on them can't reach kubedns

This is caused by nodes in the Calico etcd nodestore no longer existing. Due to the ephemeral nature of AWS EC2 instances, new nodes are brought up with different hostnames, and nodes that are taken offline remain in the Calico nodestore. This is unlike most datacentre deployments where the hostnames are mostly static in a cluster. Read more about this issue at https://github.com/kubernetes/kops/issues/3224
This has been solved in kops 1.9.0, when creating a new cluster no action is needed, but if the cluster was created with a prior kops version the following actions should be taken:

  * Use kops to update the cluster ```kops update cluster <name> --yes``` and wait for calico-kube-controllers deployment and calico-node daemonset pods to be updated
  * Decommission all invalid nodes, [see here](https://docs.projectcalico.org/v2.6/usage/decommissioning-a-node)
  * All nodes that are deleted from the cluster after this actions should be cleaned from calico's etcd storage and the delay programming routes should be solved.
