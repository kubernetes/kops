## Networking

### Common Tasks

* [Specifying a Network Plugin at Cluster Creation]
* [Modifying the Network Plugin of an Existing Cluster]
* [Troubleshooting]

### Specifying a Network Plugin at Cluster Creation

`kops create cluster` currently supports the following options, which can be
specified using the `--networking` flag:

* [`kubenet`]
* [`weave`]
* [`cni`]
* [`external`]
* [`classic`]

#### `kubenet`

`kops` uses `kubenet` by default. This sets up networking on AWS using VPC
networking, where the master allocates a /24 CIDR to each Pod, drawing from the
Pod network. Using `kubenet`, routes for each node are then configured in the
AWS VPC routing tables.

One important limitation when using `kubenet` is that an AWS routing table
cannot have more than 50 entries, which sets a limit of 50 nodes per cluster.
AWS support will sometimes raise the limit to 100, but their documentation
notes that routing tables over 50 may take a performance hit.

Because k8s modifies the AWS routing table, this means that realistically
kubernetes needs to own the routing table, and thus it requires its own subnet.
It is theoretically possible to share a routing table with other infrastructure
(but not a second cluster!), but this is not really recommended. Certain
[`cni`] networking solutions claim to address these problems.

##### Example Cluster Creation with `kubenet`

```bash
export ZONES=mylistofzones KOPS_STATE_STORE=s3://my-store
kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking kubenet \
  --yes \
  --name myclustername.mydns.io
```

#### `weave`

> TODO

#### `cni`

[Container Network Interface] provides a specification and libraries for
writing plugins to configure network interfaces in Linux containers.
Kubernetes has built in support for CNI networking components. The `cni` option
only enables CNI networking, and does not actually install any particular
plugin. This is meant to provide users a more advanced way to configure
networking. To see a full list of CNI plugins that are baked directly into
`kops`, see [Specifying a Network Plugin at Cluster Creation].

#### `external`

> TODO

#### `classic`

> TODO

## Modifying the Network Plugin of an Existing Cluster

> TODO

## Troubleshooting

### Validating CNI Installation

You will notice that `kube-dns` fails to start properly until you deploy your
CNI provider.  Pod networking and IP addresses are provided by the CNI
provider.

Here are some steps items that will confirm a good CNI install:

* `kubelet` is running with the with `--network-plugin=cni` option.
* The CNI  provider started without errors.
* `kube-dns` daemonset starts.
* Logging on a node will display messages on pod create and delete.

The `#sig-networking` and `#sig-cluster-lifecycle` channels on K8s slack are
always good starting places for Kubernetes-specific CNI challenges.

[Container Network Interface]: https://github.com/kubernetes/kops/pull/819
[Specifying a Network Plugin at Cluster Creation]: #specifying-a-network-plugin-at-cluster-creation
[Modifying the Network Plugin of an Existing Cluster]: #modifying-the-network-plugin-of-an-existing-cluster
[Troubleshooting]: #troubleshooting
[`kubenet`]: #kubenet
[`weave`]: #weave
[`cni`]: #cni
[`external`]: #external
[`classic`]: #classic
