---
title: Dynamically Attaching and Detaching Applications
menu_order: 40
---


When containers may not know the network to which they will be attached, Weave Net enables you to dynamically attach and detach containers to and from a given network, even when a container is already running. 

To illustrate...

    host1$ C=$(docker run -e WEAVE_CIDR=none -dti weaveworks/ubuntu)
    host1$ weave attach $C
    10.2.1.3

where,

 *  `C=$(docker run -e WEAVE_CIDR=none -dti weaveworks/ubuntu)` starts a container and assigns its ID to a variable
 *  `weave attach` – the Weave Net command to attach to the specified container
 *  `10.2.1.3` - the allocated IP address output by `weave attach`, in this case in the default subnet

>Note If you are using the Weave Docker API proxy, it will have modified `DOCKER_HOST` to point to the proxy and therefore you will have to pass `-e WEAVE_CIDR=none` to start a container that _doesn't_ get automatically attached to the weave network for the purposes of this example.

If `weave attach` sees the container has a hostname with a
domain-name, it will add those into WeaveDNS (unless you turn this off
with the `--without-dns` argument).

    host1$ docker run -dti --name=c1 --hostname=c1.weave.local weaveworks/ubuntu
    host1$ weave attach c1
    10.32.0.1
    host1$ weave dns-lookup c1
    10.32.0.1

If you would like `/etc/hosts` to contain the Weave Net address (the
same way [the proxy does](/site/weave-docker-api/name-resolution-proxy.md)),
specify `--rewrite-hosts` when running `weave attach`:

    host1$ weave attach --rewrite-hosts c1

### Dynamically Detaching Containers

A container can be detached from a subnet, by using the `weave detach` command:

    host1$ weave detach $C
    10.2.1.3

You can also detach a container from one network and then attach it to a different one:

    host1$ weave detach net:default $C
    10.2.1.3
    host1$ weave attach net:10.2.2.0/24 $C
    10.2.2.3

or, attach a container to multiple application networks, effectively sharing the same container between applications:

    host1$ weave attach net:default
    10.2.1.3
    host1$ weave attach net:10.2.2.0/24
    10.2.2.3

Finally, multiple addresses can be attached or detached using a single command:

    host1$ weave attach net:default net:10.2.2.0/24 net:10.2.3.0/24 $C
    10.2.1.3 10.2.2.3 10.2.3.1
    host1$ weave detach net:default net:10.2.2.0/24 net:10.2.3.0/24 $C
    10.2.1.3 10.2.2.3 10.2.3.1

>**Important!** Any addresses that were dynamically attached will not be re-attached if the container restarts.

**See Also**

 * [Adding and Removing Hosts Dynamically](/site/using-weave/finding-adding-hosts-dynamically.md)
