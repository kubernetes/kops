---
title: Autoscaling
menu_order: 40
---

### Bootstrapping

An autoscaling configuration begins with a fixed cluster:

* Configured as per the [Uniform Fixed Cluster](/site/operational-guide/uniform-fixed-cluster.md)
  scenario.
* Hosted on reserved or protected instances to ensure long-term
  stability.
* Ideally sized at a minimum of three or five nodes (you can make your
  fixed cluster bigger to accommodate base load as required; a minimum
  is specified here in the interests of resilience only.)

Building on this foundation, arbitrary numbers of dynamic peers can be
added or removed concurrently as desired, without requiring any
changes to the configuration of the fixed cluster. As with the fixed
cluster, dynamically added nodes recover automatically from reboots
and partitions.

### Scaling Out

On the additional dynamic peer, at boot, via
[systemd](/site/installing-weave/systemd.md) or equivalent:

    weave launch --no-restart --ipalloc-init=observer $PEERS

Where, 

 * `$PEERS` means all peers in the _fixed cluster_, initial
and subsequently added, which have not been explicitly removed. It
should include fixed peers which are temporarily offline or stopped.

You do not have to keep track of and specify the addresses
of other dynamic peers in `$PEERS` - they will discover and connect to
each other via the fixed cluster.

>>**Note:** The use of `--ipalloc-init=observer` prevents dynamic peers from
coming to a consensus on their own - this is important to stop a
clique forming amongst a group of dynamically added peers if they
become partitioned from the fixed cluster after having learned about
each other via discovery.

### Scaling In

On the dynamic peer to be removed:

    weave reset

If for any reason you cannot arrange for `weave reset` to be run on
the peer before the underlying host is destroyed (for example when
using spot instances that can be destroyed without notice), you will
need an asynchronous process to [reclaim lost IP address
space](/site/operational-guide/tasks.md#detect-reclaim-ipam).
