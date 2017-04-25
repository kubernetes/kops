---
title: How Weave Net Interprets Network Topology
menu_order: 20
---

This section contains the following topics: 

 * [Communicating Topology Among Peers](#topology)
 * [How Messages are Formed](#messages)
 * [Removing Peers](#removing-peers)
 * [What Happens When The Topology is Out of Date?](#out-of-date-topology)


### <a name="topology"></a>Communicating Topology Among Peers

Topology messages capture which peers are connected to other peers. 
Weave peers communicate their knowledge of the topology
(and changes to it) to others, so that all peers learn about the
entire topology. 

Communication between peers occurs over TCP links using: 
a) a spanning-tree based broadcast mechanism, and b) a
neighbor gossip mechanism.

Topology messages are sent by a peer in the following instances:

- when a connection has been added; if the remote peer appears to be
  new to the network, the entire topology is sent to it, and an
  incremental update, containing information on just the two peers at
  the ends of the connection, is broadcast,
- when a connection has been marked as 'established', indicating that
  the remote peer can receive UDP traffic from the peer; an update
  containing just information about the local peer is broadcast,
- when a connection has been torn down; an update containing just
  information about the local peer is broadcast,
- periodically, on a timer, the entire topology is "gossiped" to a
  subset of neighbors, based on a topology-sensitive random
  distribution. This is done in case some of the aforementioned
  broadcasts do not reach all peers, due to rapid changes in the
  topology causing broadcast routing tables to become outdated.

The receiver of a topology update merges that update with its own
topology model, adding peers hitherto unknown to it, and updating
peers for which the update contains a more recent version than known
to it. If there were any such new/updated peers, and the topology
update was received over gossip (rather than broadcast), then an
improved update containing them is gossiped.

If the update mentions a peer that the receiver does not know, then
the entire update is ignored.

#### <a name="messages"></a>How Messages Are Formed

Every gossip message is structured as follows:

    +-----------------------------------+
    | 1-byte message type - Gossip      |
    +-----------------------------------+
    | 4-byte Gossip channel - Topology  |
    +-----------------------------------+
    | Peer Name of source               |
    +-----------------------------------+
    | Gossip payload (topology update)  |
    +-----------------------------------+

The topology update payload is laid out like this:

    +-----------------------------------+
    | Peer 1: Name                      |
    +-----------------------------------+
    | Peer 1: NickName                  |
    +-----------------------------------+
    | Peer 1: UID                       |
    +-----------------------------------+
    | Peer 1: Version number            |
    +-----------------------------------+
    | Peer 1: List of connections       |
    +-----------------------------------+
    |                ...                |
    +-----------------------------------+
    | Peer N: Name                      |
    +-----------------------------------+
    | Peer N: NickName                  |
    +-----------------------------------+
    | Peer N: UID                       |
    +-----------------------------------+
    | Peer N: Version number            |
    +-----------------------------------+
    | Peer N: List of connections       |
    +-----------------------------------+

Each List of connections is encapsulated as a byte buffer, within
which the structure is:

    +-----------------------------------+
    | Connection 1: Remote Peer Name    |
    +-----------------------------------+
    | Connection 1: Remote IP address   |
    +-----------------------------------+
    | Connection 1: Outbound            |
    +-----------------------------------+
    | Connection 1: Established         |
    +-----------------------------------+
    | Connection 2: Remote Peer Name    |
    +-----------------------------------+
    | Connection 2: Remote IP address   |
    +-----------------------------------+
    | Connection 2: Outbound            |
    +-----------------------------------+
    | Connection 2: Established         |
    +-----------------------------------+
    |                ...                |
    +-----------------------------------+
    | Connection N: Remote Peer Name    |
    +-----------------------------------+
    | Connection N: Remote IP address   |
    +-----------------------------------+
    | Connection N: Outbound            |
    +-----------------------------------+
    | Connection N: Established         |
    +-----------------------------------+

#### <a name="removing-peers"></a>Removing Peers

If a peer, after receiving a topology update, sees that another peer
no longer has any connections within the network, it drops all
knowledge of that second peer.


#### <a name="out-of-date-topology"></a>What Happens When The Topology is Out of Date?

The propagation of topology changes to all peers is not instantaneous.
Therefore, it is very possible for a node elsewhere in the network to have an
out-of-date view.

If the destination peer for a packet is still reachable, then
out-of-date topology can result in it taking a less efficient route.

If the out-of-date topology makes it look as if the destination peer
is not reachable, then the packet is dropped.  For most protocols
(for example, TCP), the transmission will be retried a short time later, by
which time the topology should have updated.


**See Also**

 * [Weave Net Router Encapsulation](/site/how-it-works/router-encapsulation.md)
 
 
 
