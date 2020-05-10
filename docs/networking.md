## Kubernetes Networking Options

The networking options determines how the pod and service networking is implemented and managed.

Kubernetes Operations (kops) currently supports 4 networking modes:

* `kubenet` Kubernetes native networking via a CNI plugin.  This is the default.
* `cni` Container Network Interface(CNI) style networking, often installed via a Daemonset.
* `classic` Kubernetes native networking, done in-process.
* `external` networking is done via a Daemonset. This is used in some custom implementations.

### Specifying network option for cluster creation

You can specify the network provider via the `--networking` command line switch. However, this will only give a default configuration of the provider. Typically you would often modify the `spec.networking` section of the cluster spec to configure the provider further.

### Kubenet (default)

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

### CNI

[Container Network Interface](https://github.com/containernetworking/cni)  provides a specification
and libraries for writing plugins to configure network interfaces in Linux containers.  Kubernetes
has built in support for CNI networking components.

Several CNI providers are currently built into kops:

* [AWS VPC](networking/aws-vpc.md)
* [Calico](networking/calico.md)
* [Canal](networking/canal.md)
* [Cilium](networking/cilium.md)
* [Flannel](networking/flannel.md)
* [Kube-router](networking/kube-router.md)
* [Lyft VPC](networking/lyft-vpc.md)
* [Romana](networking/romana.md)
* [Weave](networking/weave.md)

The manifests for the providers are included with kops, and you simply use `--networking <provider-name>`.
Replace the provider name with the names listed above with you `kops cluster create`.  For instance
to install `calico` execute the following:

```console
kops create cluster --networking calico
```

### External CNI

When using the flag `--networking cni` on `kops create cluster`  or `spec.networking: cni {}` Kops will not install any CNI at all, but expect that you install it.

When launching a cluster in this mode, the master nodes will come up in `not ready` state. You will then be able to deploy any CNI daemonset by following vanilla kubernetes install instructions. Once the CNI daemonset has been deployed, the master nodes should enter `ready` state and the remaining nodes should join the cluster shortly after.


## Validating CNI Installation

You will notice that `kube-dns` and similar pods that depend on pod networks fails to start properly until you deploy your CNI provider.

Here are some steps items that will confirm a good CNI install:

- `kubelet` is running with the with `--network-plugin=cni` option.
- The CNS provider started without errors.
- `kube-dns` daemonset starts.
- Logging on a node will display messages on pod create and delete.

The sig-networking and sig-cluster-lifecycle channels on K8s slack are always good starting places
for Kubernetes specific CNI challenges.

## How to pick the correct network provider

Kops maintainers have no bias over the CNI provider that you run, we only aim to be flexible and provide a working setup of the CNIs.

We do recommended something other than `kubenet` for production clusters due to `kubenet`'s limitations.

## Switching between networking providers

Switching between from `classic` and `kubenet` providers to a CNI provider is considered safe. Just update the config and roll the cluster.

It is also possible to switch between CNI providers, but this usually is a distruptive change. Kops will also not clean up any resources left behind by the previous CNI, _including_ then CNI daemonset.
