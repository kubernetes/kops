---
title: Automatic Allocation Across Multiple Subnets
menu_order: 10
---


IP subnets are used to define or restrict routing. By default, Weave Net
puts all containers into a subnet that spans the entire allocation
range, so that every Weave-attached container can communicate with every other
Weave-attached container.

If you want some [isolation](/site/using-weave/application-isolation.md), you
can choose to run containers on different subnets.  To request the
allocation of an address from a particular subnet, set the
`WEAVE_CIDR` environment variable to `net:<subnet>` when creating the
container, for example:

    host1$ docker run -e WEAVE_CIDR=net:10.2.7.0/24 -ti weaveworks/ubuntu

You can ask for multiple addresses in different subnets and add in
manually-assigned addresses (outside the automatic allocation range),
for instance:

    host1$ docker run -e WEAVE_CIDR="net:10.2.7.0/24 net:10.2.8.0/24 ip:10.3.9.1/24" -ti weaveworks/ubuntu

>**Note:** The ".0" and ".-1" addresses in a subnet are not used, as required by
[RFC 1122](https://tools.ietf.org/html/rfc1122#page-29)).

When working with multiple subnets in this way, it is usually
desirable to constrain the default subnet - for example, the one chosen by the
allocator when no subnet is supplied - so that it does not overlap
with others. You can specify this by using `--ipalloc-default-subnet`:

    host1$ weave launch --ipalloc-range 10.2.0.0/16 --ipalloc-default-subnet 10.2.3.0/24

`--ipalloc-range` should cover the entire range that you will ever use
for allocation, and `--ipalloc-default-subnet` is the subnet that will
be used when you don't explicitly specify one.

When specifying addresses, the default subnet can be denoted
symbolically using `net:default`.


### <a name="manual"></a>Mixing automatic and manual allocation

Containers can be started using a mixture of automatically-allocated
addresses and manually-chosen addresses in the same range. However, you may
find that the automatic allocator has already reserved a specific
address that you wanted.

To reserve a range for manual allocation in the same subnet as the
automatic allocator, you can specify an
`--ipalloc-range` that is smaller than `--ip-default-subnet`, For
example, if you launch weave with:

    host1$ weave launch --ipalloc-range 10.9.0.0/17 --ipalloc-default-subnet 10.9.0.0/16

then you can run all containers in the 10.9.0.0/16 subnet, with
automatic allocation using the lower half, leaving the upper half free
for manual allocation.


**See Also**

 * [Address Allocation with IP Address Management (IPAM)](/site/ipam.md)
 * [Isolating Applications on a Weave Network](/site/using-weave/application-isolation.md)
 * [Starting, Stopping and Removing Peers](/site/ipam/stop-remove-peers-ipam.md)
