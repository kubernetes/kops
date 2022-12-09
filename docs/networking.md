# Kubernetes Networking Options

## Introduction

Kubernetes has a networking model in which Pods and Services have their own IP
addresses. As Pods and Services run on servers with their own IP addresses and
networking, the Kubernetes networking model is an abstraction that sits
separately from the underlying servers and networks. A number of options,
listed below, are available which implement and manage this abstraction.

## Supported networking options

The following table provides the support status for various networking providers with regards to kOps version:

As of kOps 1.26 the default network provider is Cilium. Prior to that the default is Kubenet.

| Network provider | Experimental | Stable | Deprecated |         Removed |
|------------------|-------------:|-------:|-----------:|----------------:|
| AWS VPC          |          1.9 |   1.21 |          - |               - |
| Calico           |          1.6 |   1.11 |          - |               - |
| Canal            |         1.12 |      - |          - | Kubernetes 1.26 |
| Cilium           |          1.9 |   1.15 |          - |               - |
| Cilium ENI       |         1.18 |   1.26 |          - |               - |
| Flannel udp      |        1.5.2 |      - |          - |               - |
| Flannel vxlan    |        1.8.0 |      - |          - |               - |
| Kopeio           |          1.5 |      - |          - |               - |
| Kube-router      |        1.6.2 |      - |          - |               - |
| Kubenet          |          1.5 |    1.5 |          - |               - |
| Lyft VPC         |         1.11 |      - |       1.22 |            1.23 |
| Romana           |          1.8 |      - |       1.18 |            1.19 |
| Weave            |          1.5 |      - |       1.23 | Kubernetes 1.23 |

### Specifying network option for cluster creation

You can specify the network provider via the `--networking` command line switch. However, this will only give a default configuration of the provider. Typically you would often modify the `spec.networking` section of the cluster spec to configure the provider further.

### Kubenet

The "kubenet" option has the control plane allocate a /24 CIDR to each Node, drawing from the Node network.
Routes for each node are then configured in the cloud provider network's routing tables.

One important limitation when using `kubenet` networking on AWS is that an AWS routing table cannot have more than
50 entries, which sets a limit of 50 nodes per cluster. AWS support will sometimes raise the limit to 100,
but their documentation notes that routing tables over 50 may take a performance hit.

Because kubernetes modifies the AWS routing table, this means that, realistically, Kubernetes needs to own the
routing table, and thus it requires its own subnet.  It is theoretically possible to share a routing table
with other infrastructure (but not a second cluster!), but this is not really recommended.  Certain
`cni` networking solutions claim to address these problems.

Users running `--topology private` will not be able to choose `kubenet` networking because `kubenet`
requires a single routing table. These advanced users are usually running in multiple availability zones
and as NAT gateways are single AZ, multiple route tables are needed to use each NAT gateway.

Kubenet is simple, however, it should not be used in
production clusters which expect a gradual increase in traffic and/or workload over time. Such clusters
will eventually "out-grow" the `kubenet` networking provider.

### CNI

[Container Network Interface](https://github.com/containernetworking/cni) provides a specification
and libraries for writing plugins to configure network interfaces in Linux containers.  Kubernetes
has built in support for CNI networking components.

Several CNI providers are currently built into kOps:

* [AWS VPC](networking/aws-vpc.md)
* [Calico](networking/calico.md)
* [Canal](networking/canal.md)
* [Cilium](networking/cilium.md)
* [Flannel](networking/flannel.md)
* [Kube-router](networking/kube-router.md)
* [Weave](networking/weave.md)

kOps makes it easy for cluster operators to choose one of these options. The manifests for the providers
are included with kOps, and you simply use `--networking <provider-name>`. Replace the provider name
with the name listed in the provider's documentation (from the list above) when you run
`kops cluster create`.  For instance, for a default Calico installation, execute the following:

```console
kops create cluster --networking calico
```

Later, when you run `kops get cluster -oyaml`, you will see the option you chose configured under `spec.networking`.

### Advanced

kOps makes a best-effort attempt to expose as many configuration options as possible for the upstream CNI options that it supports within the kOps cluster spec. However, as upstream CNI options are always changing, not all options may be available, or you may wish to use a CNI option which kOps doesn't support. There may also be edge-cases to operating a given CNI that were not considered by the kOps maintainers. Allowing kOps to manage the CNI installation is sufficient for the vast majority of production clusters; however, if this is not true in your case, then kOps provides an escape-hatch that allows you to take greater control over the CNI installation.

When using the flag `--networking cni` on `kops create cluster`  or `spec.networking: cni {}`, kOps will not install any CNI at all, but expect that you install it.

If you try to create a new cluster in this mode, the master nodes will come up in `not ready` state. You will then be able to deploy any CNI DaemonSet by following [the vanilla kubernetes install instructions](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#pod-network). Once the CNI DaemonSet has been deployed, the master nodes should enter `ready` state and the remaining nodes should join the cluster shortly thereafter.

#### Important Caveats

For some of the CNI implementations, kOps does more than just launch a DaemonSet with the relevant CNI pod. For example, when installing Calico, kOps installs client certificates for Calico to enable mTLS for connections to etcd. If you were to simply replace `spec.networking`'s Calico options with `spec.networking: cni {}`, you would cause an outage.

If you do decide to take manual responsibility for maintaining the CNI, you should familiarize yourself with the parts of the kOps codebase which install your CNI ([example](https://github.com/kubernetes/kops/tree/master/nodeup/pkg/model/networking)) to ensure that you are replicating any additional actions which kOps was applying for your CNI option. You should closely follow your upstream CNI's releases and kOps's releases, to ensure that you can apply any updates or fixes issued by your upstream CNI or by the kOps maintainers.

Additionally, you should bear in mind that the kOps maintainers run e2e testing over the variety of supported CNI options that a kOps update must pass in order to be released. If you take over maintaining the CNI for your cluster, you should test potential kOps, Kubernetes, and CNI updates in a test cluster before updating.

## Validating CNI Installation

You will notice that `kube-dns` and similar pods that depend on pod networks fail to start properly until you deploy your CNI provider.

Here are some steps items that will confirm a good CNI install:

- The CNS provider started without errors.
- `kube-dns` daemonset starts.
- Logging on a node will display messages on pod create and delete.

The sig-networking and sig-cluster-lifecycle channels on K8s slack are always good starting places
for Kubernetes specific CNI challenges.

## Switching between networking providers

Switching from `kubenet` providers to a CNI provider is considered safe. Just update the config and roll the cluster.

It is also possible to switch between CNI providers, but this usually is a disruptive change. kOps will also not clean up any resources left behind by the previous CNI, _including_ the CNI daemonset.

## Additional Reading

* [Kubernetes Cluster Networking Documentation](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
