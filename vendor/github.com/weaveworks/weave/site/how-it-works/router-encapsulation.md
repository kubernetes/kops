---
title: Weave Net Router Sleeve Encapsulation
menu_order: 10
---

When the Weave Net router forwards packets to peers in `sleeve` mode
(rather than via the [fast data path](/site/using-weave/fastdp.md)), the
encapsulation looks something like this:

    +-----------------------------------+
    | Name of sending peer              |
    +-----------------------------------+
    | Frame 1: Name of capturing peer   |
    +-----------------------------------+
    | Frame 1: Name of destination peer |
    +-----------------------------------+
    | Frame 1: Captured payload length  |
    +-----------------------------------+
    | Frame 1: Captured payload         |
    +-----------------------------------+
    | Frame 2: Name of capturing peer   |
    +-----------------------------------+
    | Frame 2: Name of destination peer |
    +-----------------------------------+
    | Frame 2: Captured payload length  |
    +-----------------------------------+
    | Frame 2: Captured payload         |
    +-----------------------------------+
    |                ...                |
    +-----------------------------------+
    | Frame N: Name of capturing peer   |
    +-----------------------------------+
    | Frame N: Name of destination peer |
    +-----------------------------------+
    | Frame N: Captured payload length  |
    +-----------------------------------+
    | Frame N: Captured payload         |
    +-----------------------------------+

The name of the sending peer enables the receiving peer to identify
the sender of the UDP packet. This is followed by the meta data and
a payload for one or more captured frames. The router will perform batching
if it captures several frames very quickly which all need forwarding to
the same peer. And in this instance, it will fit as many frames as possible into a single
UDP packet.

The meta data for each frame contains the names of the capturing and
the destination peers. Since the name of the capturing peer is
associated with the source MAC of the captured payload, it allows
receiving peers to build up their mappings of which client MAC
addresses are local to which peers. 

The destination peer name enables the receiving peer to identify whether this frame is destined for
itself or whether it should be forwarded on to some other peer, and 
accommodate multi-hop routing. This works even when the receiving
intermediate peer has no knowledge of the destination MAC: only the
original capturing peer needs to determine the destination peer from
the MAC. In this way Weave peers never need to exchange the MAC addresses
of clients and need not take any special action for ARP traffic and
MAC discovery.

**See Also**

 * [How Weave Net Interprets Network Topology](/site/how-it-works/network-topology.md)
