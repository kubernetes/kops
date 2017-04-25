---
title: Isolating Applications on a Weave Network
menu_order: 30
---

A single Weave network can host multiple, isolated applications where each application's containers are able 
to communicate with each other, but not with the containers of other applications.

To isolate applications, you can make use of `isolation-through-subnets` technique.
This common strategy is an example of how with Weave Net many of your `on metal` 
techniques can still be used to deploy applications to a container network.
 
To begin isolating an application (or parts of an application),  
configure Weave Net's IP allocator to manage multiple subnets. 

Using [the netcat example](/site/using-weave.md), configure multiple subsets:

    host1$ weave launch --ipalloc-range 10.2.0.0/16 --ipalloc-default-subnet 10.2.1.0/24
    host1$ eval $(weave env)
    host2$ weave launch --ipalloc-range 10.2.0.0/16 --ipalloc-default-subnet 10.2.1.0/24 $HOST1
    host2$ eval $(weave env)

This delegates the entire 10.2.0.0/16 subnet to Weave Net, and instructs
it to allocate from 10.2.1.0/24 within that, if a specific subnet is not
specified. 

Next, launch the two netcat containers onto the default subnet:

    host1$ docker run --name a1 -ti weaveworks/ubuntu
    host2$ docker run --name a2 -ti weaveworks/ubuntu

And then to test the isolation, launch a few more containers onto a different subnet:

    host1$ docker run -e WEAVE_CIDR=net:10.2.2.0/24 --name b1 -ti weaveworks/ubuntu
    host2$ docker run -e WEAVE_CIDR=net:10.2.2.0/24 --name b2 -ti weaveworks/ubuntu

Ping each container to confirm that they can talk to each other, but not to the containers of our first subnet:

    root@b1:/# ping -c 1 -q b2
    PING b2.weave.local (10.2.2.128) 56(84) bytes of data.
    --- b2.weave.local ping statistics ---
    1 packets transmitted, 1 received, 0% packet loss, time 0ms
    rtt min/avg/max/mdev = 1.338/1.338/1.338/0.000 ms

    root@b1:/# ping -c 1 -q a1
    PING a1.weave.local (10.2.1.2) 56(84) bytes of data.
    --- a1.weave.local ping statistics ---
    1 packets transmitted, 0 received, 100% packet loss, time 0ms

    root@b1:/# ping -c 1 -q a2
    PING a2.weave.local (10.2.1.130) 56(84) bytes of data.
    --- a2.weave.local ping statistics ---
    1 packets transmitted, 0 received, 100% packet loss, time 0ms

If required, a container can also be attached to multiple subnets when it is started using:

    host1$ docker run -e WEAVE_CIDR="net:default net:10.2.2.0/24" -ti weaveworks/ubuntu

`net:default` is used to request the allocation of an address from the default subnet in addition to one from an explicitly specified range.

>**Important:** Containers must be prevented from capturing and injecting raw network packets - this can be accomplished by starting them with the `--cap-drop net_raw` option.

>Note: By default docker permits communication between containers on the same host, via their docker-assigned IP addresses. For complete
isolation between application containers, that feature needs to be disabled by [setting `--icc=false`](https://docs.docker.com/engine/userguide/networking/default_network/container-communication/#communication-between-containers) in the docker daemon configuration. 

**See Also** 

 * [Automatic Allocation Across Multiple Subnets](/site/ipam/allocation-multi-ipam.md)
 * [Managing Services - Exporting, Importing, Binding and Routing](/site/using-weave/service-management.md)
