## Kubernetes Networking Setup

Kubernetes Operations (kops) currently supports 4 networking modes:

* `kubenet` Kubernetes native networking via a CNI plugin.  This is the default.
* `cni` Container Network Interface(CNI) style networking, often installed via a Daemonset.
* `classic` Kubernetes native networking, done in-process.
* `external` networking is done via a Daemonset. This is used in some custom implementations.

### kops Default Networking

Kubernetes Operations (kops) uses `kubenet` networking by default. This sets up networking on AWS using VPC
networking, where the master allocates a /24 CIDR to each Node, drawing from the Node network.
Using `kubenet` mode routes for each node are then configured in the AWS VPC routing tables.

One important limitation when using `kubenet` networking is that an AWS routing table cannot have more than
50 entries, which sets a limit of 50 nodes per cluster. AWS support will sometimes raise the limit to 100,
but their documentation notes that routing tables over 50 may take a performance hit.

Because k8s modifies the AWS routing table, this means that realistically Kubernetes needs to own the
routing table, and thus it requires its own subnet.  It is theoretically possible to share a routing table
with other infrastructure (but not a second cluster!), but this is not really recommended.  Certain
`cni` networking solutions claim to address these problems.

Users running `--topology private` will not be able to choose `kubenet` networking because `kubenet`
requires a single routing table. These advanced users are usually running in multiple availability zones
and NAT gateways are single AZ, multiple route tables are needed to use each NAT gateway.

### Supported CNI Networking

[Container Network Interface](https://github.com/containernetworking/cni)  provides a specification
and libraries for writing plugins to configure network interfaces in Linux containers.  Kubernetes
has built in support for CNI networking components.

Several different CNI providers are currently built into kops:

* [Calico](https://docs.projectcalico.org/latest/introduction) - use `--networking calico` (See [below](#calico-example-for-cni-and-network-policy) for additional configuration options.)
* [Canal (Flannel + Calico)](https://github.com/projectcalico/canal)
* [flannel](https://github.com/coreos/flannel) - use `--networking flannel-vxlan` (recommended) or `--networking flannel-udp` (legacy).  `--networking flannel` now selects `flannel-vxlan`.
* [kopeio-vxlan](https://github.com/kopeio/networking)
* [kube-router](./networking.md#kube-router-example-for-cni-ipvs-based-service-proxy-and-network-policy-enforcer)
* [romana](https://github.com/romana/romana)
* [weave](https://github.com/weaveworks/weave)
* [amazon-vpc-routed-eni](./networking.md#amazon-vpc-backend)
* [Cilium](http://docs.cilium.io)
* [Lyft cni-ipvlan-vpc-k8s](https://github.com/lyft/cni-ipvlan-vpc-k8s)

The manifests for the providers are included with kops, and you simply use `--networking provider-name`.
Replace the provider name with the names listed above with you `kops cluster create`.  For instance
to install `kopeio-vxlan` execute the following:

```console
$ kops create cluster --networking kopeio-vxlan
```

This project has no bias over the CNI provider that you run, we care that we provide the correct setup to run CNI providers.

Both `kubenet` and `classic` networking options are completely baked into kops, while since
CNI networking providers are not part of the Kubernetes project, we do not maintain
their installation processes.  With that in mind, we do not support problems with
different CNI providers but support configuring Kubernetes to run CNI providers.

## Specifying network option for cluster creation

You are able to specify your networking type via command line switch or in your yaml file.
The `--networking` option accepts the four different values defined above: `kubenet`, `cni`,
`classic`, and `external`. If `--networking` is left undefined `kubenet` is installed.

### Weave Example for CNI

#### Installation Weave on a new Cluster

The following command sets up a cluster, in HA mode, that is ready for a CNI installation.

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking cni \
  --yes \
  --name myclustername.mydns.io
```

Once the cluster is stable, which you can check with a `kubectl cluster-info` command, the next
step is to install CNI networking. Most of the CNI network providers are
moving to installing their components plugins via a Daemonset.  For instance weave will
install with the following command:

Daemonset installation for K8s 1.6.x or above.
```console
$ kubectl create -f https://git.io/weave-kube-1.6
```

Daemonset installation for K8s 1.4.x or 1.5.x.
```console
$ kubectl create -f https://git.io/weave-kube
```

### Configuring Weave MTU

The Weave MTU is configurable by editing the cluster and setting `mtu` option in the weave configuration.
AWS VPCs support jumbo frames, so on cluster creation kops sets the weave MTU to 8912 bytes (9001 minus overhead).

```
spec:
  networking:
    weave:
      mtu: 8912
```

### Configuring Weave Net EXTRA_ARGS

Weave allows you to pass command line arguments to weave by adding those arguments to the EXTRA_ARGS environmental variable.
This can be used for debugging or for customizing the logging level of weave net.

```
spec:
  networking:
    weave:
      netExtraArgs: "--log-level=info"
```

Note that it is possible to break the cluster networking if flags are improperly used and as such this option should be used with caution.

### Configuring Weave NPC EXTRA_ARGS

Weave-npc (the Weave network policy controller) allows you to customize arguments of the running binary by setting the EXTRA_ARGS environmental variable.
This can be used for debugging or for customizing the logging level of weave npc.

```
spec:
  networking:
    weave:
      npcExtraArgs: "--log-level=info"
```

Note that it is possible to break the cluster networking if flags are improperly used and as such this option should be used with caution.

### Configuring Weave network encryption

The Weave network encryption is configurable by creating a weave network secret password.
Weaveworks recommends choosing a secret with [at least 50 bits of entropy](https://www.weave.works/docs/net/latest/tasks/manage/security-untrusted-networks/).
If no password is supplied, kops will generate one at random.

```console
$ cat /dev/urandom | tr -dc A-Za-z0-9 | head -c9 > password
$ kops create secret weavepassword -f password
$ kops update cluster
```

Since unencrypted nodes will not be able to connect to nodes configured with encryption enabled, this configuration cannot be changed easily without downtime.

### Calico Example for CNI and Network Policy

#### Installing Calico on a new Cluster

The following command sets up a cluster, in HA mode, with Calico as the CNI and Network Policy provider.

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking calico \
  --yes \
  --name myclustername.mydns.io
```

The above will deploy a daemonset installation which requires K8s 1.4.x or above.

##### Enable Cross-Subnet mode in Calico (AWS only)

Calico [since 2.1](https://www.projectcalico.org/project-calico-2-1-released/) supports a new option for IP-in-IP mode where traffic is only encapsulated
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

To enable this mode in a cluster, with Calico as the CNI and Network Policy provider, you must edit the cluster after the previous `kops create ...` command.

`kops edit cluster`  will show you a block like this:

```
  networking:
    calico:
      majorVersion: v3
```

You will need to change that block, and add an additional field, to look like this:

```
  networking:
    calico:
      majorVersion: v3
      crossSubnet: true
```

This `crossSubnet` field can also be defined within a cluster specification file, and the entire cluster can be create by running:
`kops create -f k8s-cluster.example.com.yaml`

In the case of AWS, EC2 instances have source/destination checks enabled by default.
When you enable cross-subnet mode in kops, an addon controller ([k8s-ec2-srcdst](https://github.com/ottoyiu/k8s-ec2-srcdst))
will be deployed as a Pod (which will be scheduled on one of the masters) to facilitate the disabling of said source/destination address checks.
Only the masters have the IAM policy (`ec2:*`) to allow k8s-ec2-srcdst to execute `ec2:ModifyInstanceAttribute`.

### Configuring Calico MTU

The Calico MTU is configurable by editing the cluster and setting `mtu` option in the calico configuration.
AWS VPCs support jumbo frames, so on cluster creation kops sets the calico MTU to 8912 bytes (9001 minus overhead).

For more details on Calico MTU please see the [Calico Docs](https://docs.projectcalico.org/networking/mtu#determine-mtu-size).

```
spec:
  networking:
    calico:
      mtu: 8912
```

#### More information about Calico

For Calico specific documentation please visit the [Calico Docs](http://docs.projectcalico.org/latest/getting-started/kubernetes/).

For details on upgrading a Calico v2 deployment see [Calico Version 3](calico-v3.md#upgrading-an-existing-cluster).

#### Getting help with Calico

For help with Calico or to report any issues:

- [Calico Github](https://github.com/projectcalico/calico)
- [Calico Users Slack](https://calicousers.slack.com)

#### Calico Backend

In kops 1.12.0 and later Calico uses the k8s APIServer as its datastore.

In versions <1.12.0 of kops Calico uses etcd as a backend for storing information about workloads and policies. Calico does not interfere with normal etcd operations and does not require special handling when upgrading etcd.  For more information please visit the [etcd Docs](https://coreos.com/etcd/docs/latest/)

#### Configuring Calico to use Typha

As of Kops 1.12 Calico uses the kube-apiserver as its datastore. The default setup does not make use of [Typha](https://github.com/projectcalico/typha) - a component intended to lower the impact of Calico on the k8s APIServer which is recommended in [clusters over 50 nodes](https://docs.projectcalico.org/latest/getting-started/kubernetes/installation/calico#installing-with-the-kubernetes-api-datastoremore-than-50-nodes) and is strongly recommended in clusters of 100+ nodes.
It is possible to configure Calico to use Typha by editing a cluster and adding a
`typhaReplicas` option to the Calico spec:

```
  networking:
    calico:
      typhaReplicas: 3
```

#### Calico troubleshooting

##### New nodes are taking minutes for syncing ip routes and new pods on them can't reach kubedns

This is caused by nodes in the Calico etcd nodestore no longer existing. Due to the ephemeral nature of AWS EC2 instances, new nodes are brought up with different hostnames, and nodes that are taken offline remain in the Calico nodestore. This is unlike most datacentre deployments where the hostnames are mostly static in a cluster. Read more about this issue at https://github.com/kubernetes/kops/issues/3224
This has been solved in kops 1.9.0, when creating a new cluster no action is needed, but if the cluster was created with a prior kops version the following actions should be taken:

  * Use kops to update the cluster ```kops update cluster <name> --yes``` and wait for calico-kube-controllers deployment and calico-node daemonset pods to be updated
  * Decommission all invalid nodes, [see here](https://docs.projectcalico.org/v2.6/usage/decommissioning-a-node)
  * All nodes that are deleted from the cluster after this actions should be cleaned from calico's etcd storage and the delay programming routes should be solved.

### Canal Example for CNI and Network Policy

Canal is a project that combines [Flannel](https://github.com/coreos/flannel) and [Calico](http://docs.projectcalico.org/latest/getting-started/kubernetes/installation/) for CNI Networking.  It uses Flannel for networking pod traffic between hosts via VXLAN and Calico for network policy enforcement and pod to pod traffic.

#### Installing Canal on a new Cluster

The following command sets up a cluster, in HA mode, with Canal as the CNI and networking policy provider

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking canal \
  --yes \
  --name myclustername.mydns.io
```

The above will deploy a daemonset installation which requires K8s 1.4.x or above.

#### Getting help with Canal

For problems with deploying Canal please post an issue to Github:

- [Canal Issues](https://github.com/projectcalico/canal/issues)

For support with Calico Policies you can reach out on Slack or Github:

- [Calico Github](https://github.com/projectcalico/calico)
- [Calico Users Slack](https://calicousers.slack.com)

For support with Flannel you can submit an issue on Github:

- [Flannel](https://github.com/coreos/flannel/issues)

### Kube-router example for CNI, IPVS based service proxy and Network Policy enforcer

[Kube-router](https://github.com/cloudnativelabs/kube-router) is project that provides one cohesive solution that provides CNI networking for pods, an IPVS based network service proxy and iptables based network policy enforcement.

#### Installing kube-router on a new Cluster

The following command sets up a cluster with Kube-router as the CNI, service proxy and networking policy provider

```
$ kops create cluster \
  --node-count 2 \
  --zones us-west-2a \
  --master-zones us-west-2a \
  --dns-zone aws.cloudnativelabs.net \
  --node-size t2.medium \
  --master-size t2.medium \
  --networking kube-router \
  --yes \
  --name myclustername.mydns.io
```

Currently kube-router supports 1.6 and above. Please note that kube-router will also provide service proxy, so kube-proxy will not be deployed in to the cluster.

No additional configurations are required to be done by user. Kube-router automatically disables source-destination check on all AWS EC2 instances. For the traffic within a subnet there is no overlay or tunneling used. For cross-subnet pod traffic ip-ip tunneling is used implicitly and no configuration is required.

### Romana Example for CNI

#### Installing Romana on a new Cluster

The following command sets up a cluster with Romana as the CNI.

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking romana \
  --yes \
  --name myclustername.mydns.io
```

Currently Romana supports Kubernetes 1.6 and above.

#### Getting help with Romana

For problems with deploying Romana please post an issue to Github:

- [Romana Issues](https://github.com/romana/romana/issues)

You can also contact the Romana team on Slack

- [Romana Slack](https://romana.slack.com) (invite required - email [info@romana.io](mailto:info@romana.io))

#### Romana Backend

Romana uses the cluster's etcd as a backend for storing information about routes, hosts, host-groups and IP allocations.
This does not affect normal etcd operations or require special treatment when upgrading etcd.
The etcd port (4001) is opened between masters and nodes when using this networking option.

#### Amazon VPC Backend

The [Amazon VPC CNI](https://github.com/aws/amazon-vpc-cni-k8s) plugin
requires no additional configurations to be done by user.

To use the Amazon VPC CNI plugin you specify

```
  networking:
    amazonvpc: {}
```

in the cluster spec file or pass the `--networking amazon-vpc-routed-eni` option on the command line to kops:

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking amazon-vpc-routed-eni \
  --yes \
  --name myclustername.mydns.io
```

**Important:** pods use the VPC CIDR, i.e. there is no isolation between the master, node/s and the internal k8s network.

**Note:** The following permissions are added to all nodes by kops to run the provider:

```json
  {
    "Sid": "kopsK8sEC2NodeAmazonVPCPerms",
    "Effect": "Allow",
    "Action": [
      "ec2:CreateNetworkInterface",
      "ec2:AttachNetworkInterface",
      "ec2:DeleteNetworkInterface",
      "ec2:DetachNetworkInterface",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribeInstances",
      "ec2:ModifyNetworkInterfaceAttribute",
      "ec2:AssignPrivateIpAddresses",
      "ec2:UnassignPrivateIpAddresses",
      "tag:TagResources"
    ],
    "Resource": [
      "*"
    ]
  },
  {
    "Effect": "Allow",
    "Action": "ec2:CreateTags",
    "Resource": "arn:aws:ec2:*:*:network-interface/*"
  }
```

In case of any issues the directory `/var/log/aws-routed-eni` contains the log files of the CNI plugin. This directory is located in all the nodes in the cluster.

[Configuration options for the Amazon VPC CNI plugin](https://github.com/aws/amazon-vpc-cni-k8s/tree/master#cni-configuration-variables) can be set through env vars defined in the cluster spec:

```yaml
  networking:
    amazonvpc:
      env:
      - name: WARM_IP_TARGET
        value: "10"
      - name: AWS_VPC_K8S_CNI_LOGLEVEL
        value: debug
```

### Cilium Example for CNI and Network Policy

The Cilium CNI uses a Linux kernel technology called BPF, which enables the dynamic insertion of powerful security visibility and control logic within the Linux kernel.

#### Installing Cilium on a new Cluster

The following command sets up a cluster, in HA mode, with Cilium as the CNI and networking policy provider

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --networking cilium\
  --yes \
  --name cilium.example.com
```

The above will deploy a Cilium daemonset installation which requires K8s 1.10.x or above.

#### Configuring Cilium

The following command registers a cluster, but doesn't create it yet

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --networking cilium\
  --name cilium.example.com
```

`kops edit cluster`  will show you a block like this:

```
  networking:
    cilium: {}
```

You can adjust Cilium agent configuration with most options that are available in [cilium-agent command reference](http://cilium.readthedocs.io/en/stable/cmdref/cilium-agent/).

The following command will launch your cluster with desired Cilium configuration

```console
$ kops update cluster myclustername.mydns.io --yes
```

##### Using etcd for agent state sync

By default, Cilium will use CRDs for synchronizing agent state. This can cause performance problems on larger clusters. As of kops 1.18, kops can manage an etcd cluster using etcd-manager dedicated for cilium agent state sync. The [Cilium docs](https://docs.cilium.io/en/stable/gettingstarted/k8s-install-external-etcd/) contains recommendations for this must be enabled.

Add the following to `spec.etcdClusters`:
Make sure `instanceGroup` match the other etcd clusters.

```
  - etcdMembers:
    - instanceGroup: master-az-1a
      name: a
    - instanceGroup: master-az-1b
      name: b
    - instanceGroup: master-az-1c
      name: c
    name: cilium
```

Then enable etcd as kvstore:

```
  networking:
    cilium:
      etcdManaged: true
```

##### Enabling BPF NodePort

As of Kops 1.18 you can safely enable Cilium NodePort.

In this mode, the cluster is fully functional without kube-proxy, with Cilium replacing kube-proxy's NodePort implementation using BPF.
Read more about this in the [Cilium docs](https://docs.cilium.io/en/stable/gettingstarted/nodeport/)

Be aware that you need to use an AMI with at least Linux 4.19.57 for this feature to work.

```
  kubeProxy:
    enabled: false
  networking:
    cilium:
      enableNodePort: true
```

##### Enabling Cilium ENI IPAM

As of Kops 1.18, you can have Cilium provision AWS managed adresses and attach them directly to Pods much like Lyft VPC and AWS VPC. See [the Cilium docs for more information](https://docs.cilium.io/en/v1.6/concepts/ipam/eni/)

When using ENI IPAM you need to disable masquerading in Cilium as well.

```
  networking:
    cilium:
      disableMasquerade: true
      ipam: eni
```

Note that since Cilium Operator is the entity that interacts with the EC2 API to provision and attaching ENIs, we force it to run on the master nodes when this IPAM is used.

Also note that this feature has only been tested on the default kops AMIs.

#### Getting help with Cilium

For problems with deploying Cilium please post an issue to Github:

- [Cilium Issues](https://github.com/cilium/cilium/issues)

For support with Cilium Network Policies you can reach out on Slack or Github:

- [Cilium Github](https://github.com/cilium/cilium)
- [Cilium Slack](https://cilium.io/slack)

### Flannel Example for CNI

#### Configuring Flannel iptables resync period

As of Kops 1.12.0, Flannel iptables resync option is configurable via editing a cluster and adding
`iptablesResyncSeconds` option to spec:

```
  networking:
    flannel:
      iptablesResyncSeconds: 360
```

### Validating CNI Installation

You will notice that `kube-dns` fails to start properly until you deploy your CNI provider.
Pod networking and IP addresses are provided by the CNI provider.

Here are some steps items that will confirm a good CNI install:

- `kubelet` is running with the with `--network-plugin=cni` option.
- The CNS provider started without errors.
- `kube-dns` daemonset starts.
- Logging on a node will display messages on pod create and delete.

The sig-networking and sig-cluster-lifecycle channels on K8s slack are always good starting places
for Kubernetes specific CNI challenges.

#### Lyft CNI

The [lyft cni-ipvlan-vpc-k8s](https://github.com/lyft/cni-ipvlan-vpc-k8s) plugin uses Amazon Elastic Network Interfaces (ENI) to assign AWS-managed IPs to Pods using the Linux kernel's IPvlan driver in L2 mode.

Read the [prerequisites](https://github.com/lyft/cni-ipvlan-vpc-k8s#prerequisites) before starting. In addition to that, you need to specify the VPC ID as `spec.networkID` in the cluster spec file.

To use the Lyft CNI plugin you specify

```
  networking:
    lyftvpc: {}
```

in the cluster spec file or pass the `--networking lyftvpc` option on the command line to kops:

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking lyftvpc \
  --yes \
  --name myclustername.mydns.io
```

You can specify which subnets to use for allocating Pod IPs by specifying

```
  networking:
    lyftvpc:
      subnetTags:
        KubernetesCluster: myclustername.mydns.io
```

In this example, new interfaces will be attached to subnets tagged with `kubernetes_kubelet = true`.

**Note:** The following permissions are added to all nodes by kops to run the provider:

```json
  {
    "Sid": "kopsK8sEC2NodeAmazonVPCPerms",
    "Effect": "Allow",
    "Action": [
      "ec2:CreateNetworkInterface",
      "ec2:AttachNetworkInterface",
      "ec2:DeleteNetworkInterface",
      "ec2:DetachNetworkInterface",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribeInstances",
      "ec2:DescribeInstanceTypes",
      "ec2:ModifyNetworkInterfaceAttribute",
      "ec2:AssignPrivateIpAddresses",
      "ec2:UnassignPrivateIpAddresses",
      "tag:TagResources"
    ],
    "Resource": [
      "*"
    ]
  },
  {
    "Effect": "Allow",
    "Action": "ec2:CreateTags",
    "Resource": "arn:aws:ec2:*:*:network-interface/*"
  }
```

In case of any issues the directory `/var/log/aws-routed-eni` contains the log files of the CNI plugin. This directory is located in all the nodes in the cluster.

## Switching between networking providers

`kops edit cluster` and you will see a block like:

```
  networking:
    classic: {}
```

That means you are running with `classic` networking.  The `{}` means there are
no configuration options, beyond the setting `classic`.

To switch to kubenet, change the word classic to kubenet.

```
  networking:
    kubenet: {}
```

Now follow the normal update / rolling-update procedure:

```console
$ kops update cluster # to preview
$ kops update cluster --yes # to apply
$ kops rolling-update cluster # to preview the rolling-update
$ kops rolling-update cluster --yes # to roll all your instances
```
Your cluster should be ready in a few minutes. It is not trivial to see that this
has worked; the easiest way seems to be to SSH to the master and verify
that kubelet has been run with `--network-plugin=kubenet`.

Switching from `kubenet` to a CNI network provider has not been tested at this time.
