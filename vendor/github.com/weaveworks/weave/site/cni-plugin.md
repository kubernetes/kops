---
title: Integrating Kubernetes and Mesos via the CNI Plugin
menu_order: 65
---

> The recommended way of using Weave with Kubernetes is via the new
> [Kubernetes Addon](/site/kube-addon.md). The instructions below
> remain valid however, and are still the recommended method for
> integrating with Mesos.

CNI, the [_Container Network Interface_](https://github.com/containernetworking/cni),
is a proposed standard for configuring network interfaces for Linux
application containers.  CNI is supported by
[Kubernetes](http://kubernetes.io/), [Apache Mesos](mesos.apache.org)
and 
[others](https://github.com/containernetworking/cni#who-is-using-cni).

### Installing the Weave Net CNI plugin

If your machine has the directories normally used to host CNI plugins, 
then the Weave Net CNI plugin is installed when you run `weave setup`.

To create those directories, run (as root):

    mkdir -p /opt/cni/bin
    mkdir -p /etc/cni/net.d

Then run:

    weave setup

#### Launching Weave Net

To create a network that spans multiple hosts, the Weave peers must be connected to each other.  
This is accomplished by specifying the other hosts during `weave launch` or via
[`weave connect`](/site/using-weave/finding-adding-hosts-dynamically.md).

See [Creating Peer Connections Between Hosts](/site/using-weave.md#peer-connections) 
for a discussion on peer connections. 

    weave launch <peer hosts>

#### Using the CNI network configuration file

All CNI plugins are configured by a JSON file in the directory
`/etc/cni/net.d/`.  `weave setup` installs a minimal configuration
file named `10-weave.conf`, which you can alter to suit your needs.

See the [CNI Spec](https://github.com/appc/cni/blob/master/SPEC.md#network-configuration)
for details on the format and contents of this file.

By default, the Weave CNI plugin adds a default route out via an IP
address on the Weave bridge, so your containers can access resources
on the internet.  If you do not want this, add a section to the config
file that specifies no routes:

    "ipam": {
        "routes": [ ]
    }

The following other fields in the spec are supported:

- `ipam / type` - default is to use Weave's own IPAM
- `ipam / subnet` - default is to use Weave's IPAM default subnet
- `ipam / gateway` - default is to use the Weave bridge IP address (allocated by `weave expose`)

### Using the Weave Net CNI plugin

#### Configuring Kubernetes to use the CNI Plugin

After you've launched Weave and peered your hosts, you can configure
Kubernetes to use Weave, by adding the following options to the
`kubelet` command:

    --network-plugin=cni --network-plugin-dir=/etc/cni/net.d

See the [`kubelet` documentation](http://kubernetes.io/v1.1/docs/admin/kubelet.html)
for more details.

Now, whenever Kubernetes starts a pod, it will be attached to the Weave network.

#### Configuring Mesos to use the CNI plugin

To use the CNI plugin, the Mesos Agent must be started with reference
to the CNI configuration and binary directories:

    sudo mesos-slave
    --network_cni_config_dir=/etc/cni/net.d
    --network_cni_plugins_dir=/opt/cni/bin
    ...

To start a container that is connected to the Weave network via CNI,
use the name specified in the configuration file. This example starts
a alpine container running a `nc` server listening on port 1080 with
the `mesos-execute` command. From the Master, run:

    nohup sudo mesos-execute --command="ifconfig; nc -k -l 0.0.0.0 1080" --docker_image=alpine --master=localhost:5050 --name=example --networks=weave --resources=cpus:0.1 --shell </dev/null >test.log 2>&1 &

After this task has started, it is possible to obtain the ip address of
the container and ping it from any of other agents (which are also 
connected to the weave network)"

    nc -z -v <IP FROM LOGS> 1080
    
For more information, see the 
[Mesos documentation](http://mesos.apache.org/documentation/cni/).

### Caveats

- The Weave Net router container must be running for CNI to allocate addresses
- The CNI plugin does not add entries to weaveDNS.
