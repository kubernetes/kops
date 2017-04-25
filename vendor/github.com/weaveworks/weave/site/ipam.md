---
title: Allocating IP Addresses
menu_order: 70
---


Weave Net automatically assigns containers a unique IP address
across the network, and also releases that address when the container
exits. Unless you explicitly specify an address, this occurs for all
invocations of the
`attach`, `detach`, `expose`, and `hide` commands. Weave Net can also assign
addresses in multiple subnets.

The following automatic IP address management topics are discussed:

 * [Initializing Peers on a Weave Network](#initialization)
 * [`--ipalloc-init:consensus` and How Quorum is Achieved](#quorum)
 * [Priming a Peer](#priming-a-peer)
 * [Choosing an Allocation Range](#range)



### <a name="initialization"></a>Initializing Peers on a Weave Network

Three initialization strategies are available: seed, consensus and
observer. These options have different tradeoffs, so pick the one that
suits your deployment best.

#### <a name="seed"></a>Via Seed

Configuration via seed requires you to provide a list of _peer names_
(via the `--ipalloc-init seed=` parameter) amongst which the address
space will be shared initially. Normally weave derives a unique peer
name automatically at launch, but since you need to know them ahead of
time in this case you will need to name each peer explicitly via the
`--name` parameter.

> Peers in the weave network are identified by a 48-bit value
> formatted like an ethernet MAC address (e.g. 01:23:45:67:89:ab) -
> you can either specify the name fully, or substitute a single run of
> zero-octets using the `::` notation, similar to
> [IPv6 address representation](https://en.wikipedia.org/wiki/IPv6_address#Recommended_representation_as_text):
>
> * `00:00:00:00:00:01` can be written `::1`
> * `01:00:00:00:00:00` can be written `1::`
> * `01:00:00:00:00:01` can be written `1::1`

    host1$ weave launch --name ::1 --ipalloc-init seed=::1,::2,::3
    host2$ weave launch --name ::2 --ipalloc-init seed=::1,::2,::3
    host3$ weave launch --name ::3 --ipalloc-init seed=::1,::2,::3

In this configuration each peer knows in advance how the address space
has been divided up, and will be able to perform allocations from the
outset even under conditions of partition - no consensus is required.

#### <a name="consensus"></a>Via One-off Consensus

Alternatively, you can let Weave Net determine the seed automatically
via a consensus algorithm. Since you don't need to provide it with a
list of peer names anymore, you can let Weave Net derive those
automatically for you as well.

However, in order for Weave Net to form a single consensus
reliably you must now instead tell each peer how many peers there are
in total either by listing them as target peers or using the
`--ipalloc-init consensus=` parameter.

Just once, when the first automatic IP address allocation is requested
in the whole network, Weave Net needs a majority of peers to be present in
order to avoid formation of isolated groups, which can lead to
inconsistency, for example, the same IP address being allocated to two
different containers.

Therefore, you must either supply the list of all peers in the network at `weave launch` or add the
`--ipalloc-init consensus=` flag to specify how many peers there will be.

To illustrate, suppose you have three hosts, accessible to each other
as `$HOST1`, `$HOST2` and `$HOST3`. You can start Weave Net on those three
hosts using these three commands:

    host1$ weave launch $HOST2 $HOST3
    host2$ weave launch $HOST1 $HOST3
    host3$ weave launch $HOST1 $HOST2

Or, if it is not convenient to name all the other hosts at launch
time, you can pass the number of peers like this:

    host1$ weave launch --ipalloc-init consensus=3
    host2$ weave launch --ipalloc-init consensus=3 $HOST3
    host3$ weave launch --ipalloc-init consensus=3 $HOST2

The consensus mechanism used to determine a majority transitions
through three states: 'deferred', 'waiting' and 'achieved':

* 'deferred' - no allocation requests or claims have been made yet;
  consensus is deferred until then
* 'waiting' - an attempt to achieve consensus is ongoing, triggered by
  an allocation or claim request; allocations will block. This state
  persists until a quorum of peers are able to communicate amongst
  themselves successfully
* 'achieved' - consensus achieved; allocations proceed normally

#### <a name="observer"></a>Via Observation

Finally, some (but never all) peers can be launched as observers
by specifying the `--ipalloc-init observer` option:

    host4$ weave launch --ipalloc-init observer $HOST3

You do not need to specify an initial peer count or seed to such
peers. This can be useful to add peers to an existing fixed cluster
(for example in response to a scale-out event) without worrying about
adjusting initial peer counts accordingly.

#### <a name="quorum"></a> `--ipalloc-init consensus=` and How Quorum is Achieved

Normally it isn't a problem to over-estimate the value supplied to
`--ipalloc-init consensus=`, but if you supply a number that is too
small, then multiple independent groups may form.

Weave Net uses the estimate of the number of peers at initialization to
compute a majority or quorum number - specifically floor(n/2) + 1.

If the actual number of peers is less than half the number stated, then
they keep waiting for someone else to join in order to reach a quorum.

But if the actual number is more than twice the quorum
number, then you may end up with two sets of peers with each reaching a quorum and
initializing independent data structures. You'd have to be quite unlucky
for this to happen in practice, as they would have to go through the whole
agreement process without learning about each other, but it's
definitely possible.

The quorum number is only used once at start-up (specifically, the
first time someone tries to allocate or claim an IP address). Once
a set of peers is initialized, you can add more and they will join on
to the data structure used by the existing set.

The one issue to watch is if the earlier peers are restarted, you must restart
them using the current number of peers. If they use the smaller number
that was correct when they first started, then they could form an
independent set again.

To illustrate this last point, the following sequence of operations
is safe with respect to Weave Net's startup quorum:

    host1$ weave launch
    ...time passes...
    host2$ weave launch $HOST1
    ...time passes...
    host3$ weave launch $HOST1 $HOST2
    ...time passes...
    ...host1 is rebooted...
    host1$ weave launch $HOST2 $HOST3

### <a name="priming-a-peer"></a>Priming a Peer

Under certain circumstances (for example when adding new peers to an
existing network) it is desirable to ensure that a peer has
successfully joined and is ready to allocate IP addresses. An
administrative command is provided for this purpose:

    host1$ weave prime

This operation will block until the peer on which it is run has joined
successfully.

### <a name="range"></a>Choosing an Allocation Range

By default, Weave Net allocates IP addresses in the 10.32.0.0/12
range. This can be overridden with the `--ipalloc-range` option:

    host1$ weave launch --ipalloc-range 10.2.0.0/16

and must be the same on every host.

The range parameter is written in
[CIDR notation](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing) -
in this example "/16" means the first 16 bits of the address form the
network address and the allocator is to allocate container addresses
that all start 10.2. See [IP
addresses and routes](/site/how-it-works/ip-addresses.md) for more information.

Weave shares the IP address range across all peers, dynamically
according to their needs.  If a group of peers becomes isolated from
the rest (a partition), they can continue to work with the address
ranges they had before isolation, and can subsequently be re-connected
to the rest of the network without any conflicts arising.

### <a name="persistence"></a>Data persistence

Key IPAM data is saved to disk, so that it is immediately available
when the peer restarts:

* The division of the IP allocation range amongst peers
* Allocation of addresses to containers on the local peer

A [data volume
container](https://docs.docker.com/engine/userguide/containers/dockervolumes/#creating-and-mounting-a-data-volume-container)
named `weavedb` is used to store this data.

 **See Also**

 * [Automatic Allocation Across Multiple Subnets](/site/ipam/allocation-multi-ipam.md)
 * [Integrating Docker via the Network Plugin](/site/plugin.md)
