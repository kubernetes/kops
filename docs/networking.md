## Kubernetes Networking Setup

Kubernetes Operations (kops) currently supports 4 networking modes:

* `kubenet` kubernetes native networking via a CNI plugin.  This is the default.
* `cni` Container Network Interface(CNI) style networking, often installed via a Daemonset.
* `classic` kubernetes native networking, done in-process.
* `external` networking is done via a Daemonset. This is used in some custom implementations.

### kops Default Networking

Kubernetes Operations (kops) uses `kubenet` networking by default. This sets up networking on AWS using VPC
networking, where the  master allocates a /24 CIDR to each Node, drawing from the Node network.  
Using `kubenet` mode routes for  each node are then configured in the AWS VPC routing tables.

One important limitation when using `kubenet` networking is that an AWS routing table cannot have more than
50 entries, which sets a limit of 50 nodes per cluster. AWS support will sometimes raise the limit to 100,
but their documentation notes that routing tables over 50 may take a performance hit.

Because k8s modifies the AWS routing table, this means that realistically kubernetes needs to own the
routing table, and thus it requires its own subnet.  It is theoretically possible to share a routing table
with other infrastructure (but not a second cluster!), but this is not really recommended.  Certain
`cni` networking solutions claim to address these problems.

### Supported CNI Networking

Several different providers are currently built into kops:

1. kopeio-vxlan
2. [weave](https://github.com/weaveworks/weave-kube)
3. [flannel](https://github.com/coreos/flannel)
4. [Calico](http://docs.projectcalico.org/v2.0/getting-started/kubernetes/installation/hosted/)
5. [Canal (Flannel + Calico)](https://github.com/projectcalico/canal)

The manifests for the providers are included with kops, and you simply use `--networking provider-name`.
Replace the provider name with the names listed above with you `kops cluster create`.  For instance
to install `kopeio-vxlan` execute the following:

```console
$ kops create cluster --networking kopeio-vxlan
```

### CNI Networking

[Container Network Interface](https://github.com/containernetworking/cni)  provides a specification
and libraries for writing plugins to configure network interfaces in Linux containers.  Kubernetes
has built in support for CNI networking components.  Various solutions exist that
support Kubernetes CNI networking, listed in alphabetical order:

- [Romana](https://github.com/romana/romana/tree/master/containerize#using-kops)

This is not an all comprehensive list. At the time of writing this documentation, weave has
been tested and used in the example below.  This project has no bias over the CNI provider
that you run, we care that we provide the correct setup to run CNI providers.

Both `kubenet` and `classic` networking options are completely baked into kops, while since
CNI networking providers are not part of the Kubernetes project, we do not maintain
their installation processes.  With that in mind, we do not support problems with
different CNI providers but support configuring Kubernetes to run CNI providers.

## Specifying network option for cluster creation

You are able to specify your networking type via command line switch or in your yaml file.
The `--networking` option accepts the three different values defined above: `kubenet`, `cni`,
`classic`, and `external`. If `--networking` is left undefined `kubenet` is installed.

### Weave Example for CNI

#### Installation Weave on a new Cluster

The following command sets up a cluster, in HA mode, that is ready for a CNI installation.

```console
$ export $ZONE=mylistofzones
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

```console
$ kubectl create -f https://git.io/weave-kube
```

The above daemonset installation requires K8s 1.4.x or above.

### Calico Example for CNI and Network Policy

#### Installing Calico on a new Cluster

The following command sets up a cluster, in HA mode, with Calico as the CNI and Network Policy provider.

```console
$ export $ZONES=mylistofzones
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

#### More information about Calico

For Calico specific documentation please visit the [Calico Docs](http://docs.projectcalico.org/v2.0/getting-started/kubernetes/).

#### Getting help with Calico

For help with Calico or to report any issues:

- [Calico Github](https://github.com/projectcalico/calico)
- [Calico Users Slack](https://calicousers.slack.com)

#### Calico Backend

Calico currently uses etcd as a backend for storing information about workloads and policies.  Calico does not interfere with normal etcd operations and does not require special handling when upgrading etcd.  For more information please visit the [etcd Docs](https://coreos.com/etcd/docs/latest/)

### Canal Example for CNI and Network Policy

Canal is a project that combines [Flannel](https://github.com/coreos/flannel) and [Calico](http://docs.projectcalico.org/v2.0/getting-started/kubernetes/installation/hosted/) for CNI Networking.  It uses Flannel for networking pod traffic between hosts via VXLAN and Calico for network policy enforcement and pod to pod traffic.

#### Installing Canal on a new Cluster

The following command sets up a cluster, in HA mode, with Canal as the CNI and networking policy provider

```console
$ export $ZONES=mylistofzones
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

### Validating CNI Installation

You will notice that `kube-dns` fails to start properly until you deploy your CNI provider.
Pod networking and IP addresses are provided by the CNI provider.

Here are some steps items that will confirm a good CNI install:

- `kubelet` is running with the with `--network-plugin=cni` option.
- The CNS provider started without errors.
- `kube-dns` daesonset starts.
- Logging on a node will display messages on pod create and delete.

The sig-networking and sig-cluster-lifecycle channels on K8s slack are always good starting places
for Kubernetes specific CNI challenges.

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
