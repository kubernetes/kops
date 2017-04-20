---
title: Starting, Stopping and Removing Peers
menu_order: 20
---


You may wish to `weave stop` and re-launch to change some config or to
upgrade to a new version. Provided that the underlying protocol hasn't
changed, Weave Net picks up where it left off and learns from peers in
the network which address ranges it was previously using.

If, however, you run `weave reset` this removes the peer from the
network so if Weave Net is run again on that node it will start from
scratch.

For failed peers, the `weave rmpeer` command can be invoked to
permanently remove the ranges allocated to said peers.  This allows
other peers to allocate IPs in the ranges previously owned by the
removed peers, and as such should be used with extreme caution: if the
removed peers had transferred some range of IP addresses to other
peers but this is not known to the whole network, or if some of them
later rejoin the Weave network, the same IP address may be allocated
twice.

Assume you had started the three peers in the
[overview example](/site/ipam.md), and then host3
caught fire, you can go to one of the other hosts and run:

    host1$ weave rmpeer host3
    524288 IPs taken over from host3

Weave Net takes all the IP address ranges owned by host3 and transfers
them to be owned by host1. The name "host3" is resolved via the
'nickname' feature of Weave Net, which defaults to the local host
name. Alternatively, you can supply a peer name as shown in `weave status`.

### <a name="caution-rmpeer"></a>Caution###

You cannot call `weave rmpeer` on more than one host. The address
space, which was owned by the stale peer cannot be left dangling, and
as a result it gets reassigned. In this instance, the address is
reassigned to the peer on which `weave rmpeer` was run. Therefore, if
you run `weave forget` and then `weave rmpeer` on more than one host
at a time, it results in duplicate IPs on more than one host.

Once the peers detect the inconsistency, they log the error and drop
the connection that supplied the inconsistent data. The rest of the
peers will carry on with their view of the world, but the network will
not function correctly.

Some peers may be able to communicate their claim to the others before
they run `rmpeer` (i.e. it's a race), so what you can expect is a few
cliques of peers that are still talking to each other, but repeatedly
dropping attempted connections with peers in other cliques.

**See Also**

 * [Address Allocation with IP Address Management (IPAM)](/site/ipam.md)
 * [Automatic Allocation Across Multiple Subnets](/site/ipam/allocation-multi-ipam.md)
 * [Isolating Applications on a Weave Network](/site/using-weave/application-isolation.md)
