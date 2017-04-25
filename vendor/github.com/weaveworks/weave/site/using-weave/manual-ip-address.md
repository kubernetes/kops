---
title: Manually Specifying the IP Address of a Container
menu_order: 20
---

Containers are automatically allocated an IP address that is unique across the Weave network. You can see which address was allocated by running, [`weave ps`](/site/troubleshooting.md#weave-status):

    host1$ weave ps a1
    a7aee7233393 7a:44:d3:11:10:70 10.32.0.2/12

Weave Net detects when a container has exited and releases its allocated addresses so they can be re-used by the network.

See the [Automatic IP Address Management](/site/ipam.md) and also an explanation of [the basics of IP addressing](/site/how-it-works/ip-addresses.md) for further details.

Instead of allowing Weave Net to allocate IP addresses automatically (using IPAM), there may be instances where you need to control a particular container or a cluster by setting an IP address for it.  

You can specify an IP address and a network explicitly, using Classless Inter-Domain Routing or [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing).

For example, we can launch a couple of containers on `$HOST1` and
`$HOST2`, respectively, with specified IP addresses, as follows...

On `$HOST1`:

    host1$ docker run -e WEAVE_CIDR=10.2.1.1/24 -ti weaveworks/ubuntu
    root@7ca0f6ecf59f:/#

And `$HOST2`:

    host2$ docker run -e WEAVE_CIDR=10.2.1.2/24 -ti weaveworks/ubuntu
    root@04c4831fafd3:/#

Then test that the container on `$HOST2` can be reached from the container on `$HOST1`:

    root@7ca0f6ecf59f:/# ping -c 1 -q 10.2.1.2
    PING 10.2.1.2 (10.2.1.2): 48 data bytes
    --- 10.2.1.2 ping statistics ---
    1 packets transmitted, 1 packets received, 0% packet loss
    round-trip min/avg/max/stddev = 1.048/1.048/1.048/0.000 ms

And in the other direction...

    root@04c4831fafd3:/# ping -c 1 -q 10.2.1.1
    PING 10.2.1.1 (10.2.1.1): 48 data bytes
    --- 10.2.1.1 ping statistics ---
    1 packets transmitted, 1 packets received, 0% packet loss
    round-trip min/avg/max/stddev = 1.034/1.034/1.034/0.000 ms

The IP addresses and netmasks can be set to anything, but ensure they don’t conflict with any of the IP ranges in use on the hosts or with IP addresses used by any external services to which the hosts or containers may need to connect. 

Individual IP addresses given to containers must, of course, be unique. If you pick an address that the automatic allocator has already assigned a warning appears.

**See Also**

 * [Managing Services - Exporting, Importing, Binding and Routing](/site/using-weave/service-management.md)
 * [Allocating IPs in a Specific Range](/site/using-weave/configuring-weave.md) 
 * [Automatic IP Address Management](/site/ipam.md)   

