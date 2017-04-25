---
title: Integrating Kubernetes via the Addon
menu_order: 63
---

The following topics are discussed:

* [Installation](#install)
 * [Upgrading Kubernetes to version 1.6](#kube-1.6-upgrade)
 * [Upgrading the Daemon Sets](#daemon-sets)
* [Network Policy Controller](#npc)
 * [Troubleshooting Blocked Connections](#blocked-connections)
 * [Changing Configuration Options](#configuration-options)


## <a name="install"></a> Installation

Weave Net can be installed onto your CNI-enabled Kubernetes cluster
with a single command:

* Kubernetes versions `1.6` and above:

```
$ kubectl apply -f https://git.io/weave-kube-1.6
```

* Kubernetes versions up to `1.5`:

```
$ kubectl apply -f https://git.io/weave-kube
```

After a few seconds, a Weave Net pod should be running on each
Node and any further pods you create will be automatically attached to the Weave
network.

**Note:** This command requires Kubernetes 1.4 or later.

> CNI, the [_Container Network Interface_](https://github.com/containernetworking/cni),
> is a proposed standard for configuring network interfaces for Linux
> containers.
>
> If you do not already have a CNI-enabled cluster, you can bootstrap
> one easily with
> [kubeadm](http://kubernetes.io/docs/getting-started-guides/kubeadm/).
>
> Alternatively, you can [configure CNI yourself](http://kubernetes.io/docs/admin/network-plugins/#cni)

**Note:** If using the [Weave CNI
Plugin](/site/cni-plugin.md) from a prior full install of Weave Net with your
cluster, you must first uninstall it before applying the Weave-kube addon.
Shut down Kubernetes, and _on all nodes_ perform the following:

 * `weave reset`
 * Remove any separate provisions you may have made to run Weave at
   boot-time, e.g. `systemd` units
 * `rm /opt/cni/bin/weave-*`

Then relaunch Kubernetes and install the addon as described
above.

The URLs [https://git.io/weave-kube](https://git.io/weave-kube) and [https://git.io/weave-kube-1.6](https://git.io/weave-kube-1.6) point
to the YAML file for the [latest release](https://github.com/weaveworks/weave/releases/tag/latest_release) of the Weave Net addon.
Historic versions are archived on our [GitHub release
page](https://github.com/weaveworks/weave/releases).

## <a name="kube-1.6-upgrade"></a> Upgrading Kubernetes to version 1.6

In version 1.6, Kubernetes has increased security, so we need to
create a special service account to run Weave Net. This is done in
the file `weave-daemonset-k8s-1.6.yaml` attached to the [Weave Net
release](https://github.com/weaveworks/weave/releases/latest).

Also, the
[toleration](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/taint-toleration-dedicated.md)
required to let Weave Net run on master nodes has moved from an
annotation to a field on the DaemonSet spec object.

If you have edited the Weave Net DaemonSet from a previous release,
you will need to re-make your changes against the new version.

### <a name="daemon-sets"></a> Upgrading the Daemon Sets

Kubernetes does not currently support rolling upgrades of daemon sets,
and so you will need to perform the procedure manually:

* Apply the updated addon manifest `kubectl apply -f https://git.io/weave-kube`
* Kill each Weave Net pod with `kubectl delete` and then wait for it to reboot before moving on to the next pod.

**Note:** If you delete all Weave Net pods at the same time they will
  lose track of IP address range ownership, possibly leading to
  duplicate IP addresses if you then start a new copy of Weave Net.

## <a name="npc"></a>Network Policy Controller

The addon also supports the [Kubernetes policy
API](http://kubernetes.io/docs/user-guide/networkpolicies/) so that
you can securely isolate pods from each other based on namespaces and
labels. For more information on configuring network policies in
Kubernetes see the
[walkthrough](http://kubernetes.io/docs/getting-started-guides/network-policy/walkthrough/)
and the [NetworkPolicy API object
definition](http://kubernetes.io/docs/api-reference/extensions/v1beta1/definitions/#_v1beta1_networkpolicy).

**Note:** as of version 1.9 of Weave Net, the Network Policy
  Controller allows all multicast traffic. Since a single multicast
  address may be used by multiple pods, we cannot implement rules to
  isolate them individually.  You can turn this behaviour off (block
  all multicast traffic) by adding `--allow-mcast` as an argument to
  `weave-npc` in the YAML configuration.

### <a name="blocked-connections"></a> Troubleshooting Blocked Connections

If you suspect that legitimate traffic is being blocked by the Weave Network Policy Controller, the first thing to do is check the `weave-npc` container's logs.

To do this, first you have to find the name of the Weave Net pod running on the relevant host:

```
$ kubectl get pods -n kube-system -o wide | grep weave-net
weave-net-08y45                  2/2       Running   0          1m        10.128.0.2   host1
weave-net-2zuhy                  2/2       Running   0          1m        10.128.0.4   host3
weave-net-oai50                  2/2       Running   0          1m        10.128.0.3   host2
```

Select the relevant container, for example, if you want to look at host2 then pick `weave-net-oai50` and run:

```
$ kubectl logs <weave-pod-name-as-above> -n kube-system weave-npc
```

When the Weave Network Policy Controller blocks a connection, it logs the following details about it:

* protocol used, 
* source IP and port, 
* destination IP and port, 

as per the below example:

```
TCP connection from 10.32.0.7:56648 to 10.32.0.11:80 blocked by Weave NPC.
UDP connection from 10.32.0.7:56648 to 10.32.0.11:80 blocked by Weave NPC.
```

### <a name="configuration-options"></a> Changing Configuration Options

The default configuration settings can be changed by saving and editing the
addon YAML before running `kubectl apply`. Additional arguments may be
supplied to the Weave router process by adding them to the `command:`
array in the YAML file.

Some parameters are changed by environment variables; these can be
inserted into the YAML file like this:

```
      containers:
        - name: weave
          env:
            - name: IPALLOC_RANGE
              value: 10.0.0.0/16
```

The list of variables you can set is:

* `CHECKPOINT_DISABLE` - if set to 1, disable checking for new Weave Net
  versions (default is blank, i.e. check is enabled)
* `IPALLOC_RANGE` - the range of IP addresses used by Weave Net
  and the subnet they are placed in (CIDR format; default `10.32.0.0/12`)
* `EXPECT_NPC` - set to 0 to disable Network Policy Controller (default is on)
* `KUBE_PEERS` - list of addresses of peers in the Kubernetes cluster
  (default is to fetch the list from the api-server)
* `IPALLOC_INIT` - set the initialization mode of the [IP Address
  Manager](/site/operational-guide/concepts.md#ip-address-manager)
  (defaults to consensus amongst the `KUBE_PEERS`)
* `WEAVE_EXPOSE_IP` - set the IP address used as a gateway from the
  Weave network to the host network - this is useful if you are
  configuring the addon as a static pod.
* `WEAVE_MTU` - Weave Net defaults to 1376 bytes, but you can set a
  smaller size if your underlying network has a tighter limit, or set
  a larger size for better performance if your network supports jumbo
  frames - see [here](/site/using-weave/fastdp.md#mtu) for more
  details.
