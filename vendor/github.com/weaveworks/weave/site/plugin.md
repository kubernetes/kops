---
title: Integrating Docker via the Network Plugin
menu_order: 60
---

 * [Launching Weave Net and Running Containers Using the Plugin](#launching)
 * [Restarting the Plugin](#restarting)
 * [Configuring the Plugin](#configuring)

Docker versions 1.9 and later have a plugin mechanism for adding
different network providers. Weave Net installs itself as a network
plugin when you start it with `weave launch`. The Weave Docker
Networking Plugin is fast and easy to use, and, unlike other
networking plugins, does not require an external cluster store.

To create a network which can span multiple Docker hosts, Weave Net peers must be connected to each other, by specifying the other hosts during `weave launch` or via
[`weave connect`](/site/using-weave/finding-adding-hosts-dynamically.md).

See [Using Weave Net](/site/using-weave.md#peer-connections) for a discussion on peer connections. 

After you've launched Weave Net and peered your hosts,  you can start containers using the following, for example:

    $ docker run --net=weave -ti weaveworks/ubuntu

on any of the hosts, and they can all communicate with each other
using any protocol, even multicast.

In order to use Weave Net's [Service Discovery](/site/weavedns.md) you
must pass the additional arguments `--dns` and `-dns-search`, for
which a helper is provided in the Weave script:

    $ docker run --net=weave -h foo.weave.local $(weave dns-args) -tdi weaveworks/ubuntu
    $ docker run --net=weave -h bar.weave.local $(weave dns-args) -ti weaveworks/ubuntu
    # ping foo


### <a name="launching"></a>Launching Weave Net and Running Containers Using the Plugin

Just launch the Weave Net router onto each host and make a peer connection with the other hosts:

    host1$ weave launch host2
    host2$ weave launch host1

then run your containers using the Docker command-line:

    host1$ docker run --net=weave -ti weaveworks/ubuntu
    root@1458e848cd90:/# hostname -i
    10.32.0.2

    host2$ docker run --net=weave -ti weaveworks/ubuntu
    root@8cc4b5dc5722:/# ping 10.32.0.2

    PING 10.32.0.2 (10.32.0.2) 56(84) bytes of data.
    64 bytes from 10.32.0.2: icmp_seq=1 ttl=64 time=0.116 ms
    64 bytes from 10.32.0.2: icmp_seq=2 ttl=64 time=0.052 ms


### <a name="multi"></a>Creating multiple Docker Networks

Docker enables you to create multiple independent networks and attach
different sets of containers to each network. However, coordinating
this between hosts requires that you run Docker in ["swarm mode"](https://docs.docker.com/engine/swarm/swarm-mode/) or configure a
["key-value store"](https://docs.docker.com/engine/userguide/networking/get-started-overlay/#/set-up-a-key-value-store).

Docker swarm mode requires Docker version 1.13 or later to work with
plugins such as Weave Net.

To create a new network for services in swarm mode, run:

    $ docker network create --driver=weaveworks/net-plugin:2.0.0 mynetwork

then use it to create a service:

    $ docker service create --network=mynetwork --name myservice ...


To create a new network to attach containers in swarm mode, run:

    $ docker network create --driver=weave --attachable mynetwork

then use it to connect a container:

    $ docker run --net=mynetwork ...


If your Docker installation has a key-value store, create a network
based on Weave Net as follows:

    $ docker network create --driver=weave mynetwork

then use it to connect a container:

    $ docker run --net=mynetwork ...

or

    $ docker network connect mynetwork somecontainer

Containers attached to different Docker Networks are
[isolated through subnets](https://www.weave.works/docs/net/latest/using-weave/application-isolation/).


### <a name="restarting"></a>Restarting the Plugin

The plugin, like all Weave Net components, is started with a policy of `--restart=always`, so that it is always there after a restart or reboot. If you remove this container (for example, when using `weave reset`) before removing all endpoints created using `--net=weave`, Docker may hang for a long time when it subsequently tries to re-establish communications to the plugin.

Unfortunately, [Docker 1.9 may also try to communicate with the plugin before it has even started it](https://github.com/docker/libnetwork/issues/813).

If you are using `systemd` with Docker 1.9, it is advised that you modify the Docker unit to remove the timeout on startup. This gives Docker enough time to abandon its attempts. For example, in the file `/lib/systemd/system/docker.service`, add the following under `[Service]`:

    TimeoutStartSec=0


### <a name="configuring"></a>Configuring the Plugin

The plugin accepts a number of configuration parameters. To supply
these, instead of running `weave launch`, start the router and plugin
separately with:

    $ weave launch-router [other peers]
    $ weave launch-plugin [plugin parameters]

The plugin configuration parameters are:

 * `--log-level=debug|info|warning|error` --tells the plugin
   how much information to emit for debugging.
 * `--no-restart` -- remove the default policy of `--restart=always`, if
   you want to control start-up of the plugin yourself


**See Also**

 * [How the Weave Network Plugin Works](/site/plugin/plugin-how-it-works.md)
