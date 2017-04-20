See the [requirements](https://github.com/zettio/weave/wiki/IP-allocation-requirements).

At its highest level, the idea is that we start with a certain IP
address space, known to all peers, and divide it up between the
peers. This allows peers to allocate and free individual IPs locally
until they run out.

We use a CRDT to represent shared knowledge about the space,
transmitted over the Weave Gossip mechanism, together with
point-to-point messages for one peer to request more space from
another.

The allocator running at each peer also has an http interface which
the container infrastructure (e.g. the Weave script, or a Docker
plugin) uses to request IP addresses.

![Schematic of IP allocation in operation](https://docs.google.com/drawings/d/1-EUIRKYxwfKTpBJ7v_LMcdvSpodIMSz4lT3wgEfWKl4/pub?w=701&h=310)

## Commands

The commands supported by the allocator via the http interface are:

- Allocate: request one IP address for a container in a subnet
- Lookup: fetch the previously-allocated IP address for a container in a subnet
- Free: return an IP address that is currently allocated
- Claim: request a specific IP address for a container (e.g. because
  it is already using that address)

Each http request either specifies a subnet, or if no subnet is
specified this is taken as a request to allocate in a pre-defined
default subnet.

The allocator also watches via the Docker event mechanism: if a
container dies then all IP addresses allocated to that container are
freed.

## Definitions

1. Allocations. We use the word 'allocation' to refer to a specific
   IP address being assigned, e.g. so it can be given to a container.

2. Range. Most of the time, instead of dealing with individual IP
   addresses, we operate on them in contiguous groups, for which we
   use the word "range".

3. Ring. We consider the address space as a ring, so ranges wrap
   around from the highest address to the lowest address.

4. Peer. A Peer is a node on the Weave network. It can own zero or
   more ranges.

### The Allocation Process

When a peer owns some range(s), it can allocate freely from within
those ranges to containers on the same machine. If it runs out of
space in a subnet (all owned ranges are full), it will ask another
peer for space:
  - it picks a peer to ask at random, weighted by the amount of space
    owned by each peer in the subnet
    - if the target peer decides to give up space, it unicasts a message
      back to the asker with the newly-updated ring.
    - if the target peer has no space, it unicasts a message back to the
      asker with its current copy of the ring, on the basis that the
      requestor must have acted on out-of-date information.
  - it will continue to ask for space until it receives some, or its
    copy of the ring tells it all peers are full in that subnet.

### Data persistence

Key IPAM data is saved to disk, in a [BoltDB](https://github.com/boltdb/bolt)
file, stored within a [data volume container](https://docs.docker.com/engine/userguide/containers/dockervolumes/#creating-and-mounting-a-data-volume-container)
named `weavedb`.

This file is used to persist data for various Weave Net components; for IPAM it contains:

* the division of the IP allocation range amongst peers, and
* allocation of addresses to containers on the local peer,

so that it is immediately available when the peer restarts.

### Claiming an address

If a Weave process is restarted,

- if `weavedb` is present, then it loads persisted IPAM data from there, as
described, in the previous section;
- else, it learns from other peers which ranges it used to own.

Weave Net then consults the allocator to know which individual IP addresses
are assigned to containers, and therefore avoid giving the same address out in
subsequent allocation requests.

When the Allocator is told to claim an address, there are four
possibilities:
  - the address is outside of the space managed by this Allocator, in
    which case we ignore the request.
  - the address is owned by this peer, in which case we record the
    address as being assigned to a particular container and return
    success.
  - the address is owned by a different peer, in which case we return
    failure.
  - we have not yet heard of any address ranges being owned by anyone,
    in which case we wait until we do hear.

This approach fails if the peer does not hear from another peer about
the ranges it used to own, e.g. if all peers in a network partition
are restarted at once.

### The Ring CRDT

We use a Convergent Replicated Data Type - a CRDT - so that peers can
make changes concurrently and communicate them without central
coordination. To achieve this, we arrange that peers only make changes
to the data structure in ranges that they own (except under
administrator command - see later).

The data structure is a set of tokens, each containing the name of an
owning peer. A peer can own many ranges. Each token is placed at the
start address of a range, and the set is kept ordered so each range
goes from one token to the next. Each range on the ring includes the
start address but does not include the end address (which is the start
of the next range).  Ranges wrap, so the 'next' token after the last
one is the first token.

![Tokens on the Ring](https://docs.google.com/drawings/d/1hp--q2vmxbBAnPjhza4Kqjr1ugrw2iS1M1GerhH-IKY/pub?w=960&h=288)

In more detail:
- Each token is a tuple {peer name, version}, placed
  at an IP address.
- Peer names are taken from Weave: they are unique and survive across restarts.
- The contents of a token can only be updated by the owning peer, and
  when this is done the version is incremented
- The ring data structure is always gossiped in its entirety
- The merge operation when a peer receives a ring via gossip is:
  - Tokens with unique addresses are just copied into the combined ring
  - For tokens at the same address, pick the one with the highest
    version number
- The data gossiped about the ring also includes the amount of free
  space in each range: this is not essential but it improves the
  selection of which peer to ask for space.
- When a peer is asked for space, there are four scenarios:
  1. It has an empty range; it can change the peer associated with
     the token at the beginning of the range and increment the version.
  2. It has a range which can be subdivided by a single token to form
     a free range.  It inserts said token, owned by the peer requesting
     space.
  3. It has a 'hole' in the middle of a range; an empty range can be
     created by inserting two tokens, one at the beginning of the hole
     owned by the peer requesting the space, and one at the end of the
     hole owned by the requestee.
  4. It has no space.

## Initialisation

The previous sections describe how peers coordinate changes to the
ring.  But how is the initial state of the ring established?  If a new
peer joins a long-standing cluster, it can learn about the ring state
from other peers.  But in a freshly started cluster, the initial ring
state must be established from a clean slate.  And it must be
consistent across all peers.  This is a distributed consensus problem,
and we solve it using the Paxos algorithm.

Although Paxos has a reputation for being hard to understand and to
implement, the implementation used for ring initialization is
relatively straightforward:

- We only need to establish consensus for a single value, rather than
  a succession of values (as in a replicated transaction log).  Thus
  we implement basic Paxos, rather than multi-Paxos, and the
  implementation closely follows the outline of basic Paxos described
  [on
  wikipedia](http://en.wikipedia.org/wiki/Paxos_%28computer_science%29#Basic_Paxos).

- All peers play all the roles described in basic Paxos: Proposer,
  acceptor and listener.

- Paxos is usually described in terms of unicast messages from one
  agent to another, although three of the four messages are fanned out
  to all agents in a certain role ("prepare" and "accept request" from
  a proposer to all acceptors; "accepted" from an acceptor to all
  learners).  So most implementations involve a communications layer
  to implement the required communication patterns.  But our
  implementation is built on top of the weave gossip layer, which
  gives us a broadcast medium.  Using broadcast for the one truly
  unicast message ("promise") may seem wasteful, but as we only
  establish consensus for a single value, it is not a major concern.

The value the peers obtain consensus on is the set of peers that will
be represented in the initial ring (with each peer getting an equal
share of the address space).  When a proposing peer gets to choose the
value, it includes all the peers it has heard from during the Paxos
phase, which is at least a quorum.

Once a consensus is reached, it is used to initialize the ring.  If a
peer hears about an initialized ring via gossip, that implies that a
consensus was reached, so it will stop participating in Paxos.

Paxos requires that we know the quorum size for the cluster.  As weave
clusters are a dynamic entity, there is no fixed quorum size.  But the
Paxos phase used to initialize the ring should normally be
short-lived, so we just need to know the initial cluster size.  A
heuristic is used to derive this from the list of addresses passed to
`weave launch`, but the user can set it explicitly.

## Peer shutdown

When a peer leaves (a `weave reset` command), it grants all its own
tokens to another peer, then broadcasts the updated ring.

After sending the message, the peer terminates - it does not wait for
any response.

Failures:
- message lost
  - the space will be unusable by other peers because it will still be
    seen as owned.

To cope with the situation where a peer has left or died without
managing to tell its peers, an administrator may go to any other peer
and command that it take over the dead peer's tokens (with `weave
rmpeer`).  This information will then be gossipped out to the network.

(This operation is not guaranteed to be safe - if the dead peer made
some transfers which have are known to other peers but have not
reached the peer that does the `rmpeer`, then we get an inconsistent
ring.  We could make it safe, up to partitions, by inquiring of every
live peer what its view of the dead peer is, and making sure we use
the latest info.)
