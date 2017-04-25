---
title: Uniform Dynamic Clusters
menu_order: 50
---

A uniform dynamic cluster has the following characteristics:

* Recovers automatically after reboots and partitions.
* Identical configuration for all of its peers.
* Initial cluster peers can make progress from the outset even if
  bootstrapping occurs under conditions of partition.
* Once an initial cluster has been bootstrapped, arbitrary numbers of
  new peers can be added in parallel without coordination. This makes
  it ideally suited for use with convergent provisioning tools that
  operate across multiple hosts in an asynchronous fashion.

## Bootstrapping

On each initial peer N, at boot, via
[systemd](/site/installing-weave/systemd.md):

    hostN$ weave launch --no-restart --name ::N --ipalloc-init seed=$SEED $PEERS

Where, 

* `$SEED` is a comma separated list of the _names_ of the initial
peers, for example, `::1,::2,::3` and, 
* `$PEERS` is obtained from `/etc/sysconfig/weave` as described in the linked systemd
documentation, and includes the complete set of initial peer
_addresses_.

For example, if you have three initial peers you would specify the
following:

    host1$ weave launch --no-restart --name ::1 --ipalloc-init seed=$SEED $PEERS
    host2$ weave launch --no-restart --name ::2 --ipalloc-init seed=$SEED $PEERS
    host3$ weave launch --no-restart --name ::3 --ipalloc-init seed=$SEED $PEERS

Where

    SEED="::1,::2,::3"
    PEERS="host1 host2 host3"

## Adding a Peer

On each new peer, at boot, via
[systemd](/site/installing-weave/systemd.md):

    hostN$ weave launch --no-restart --name ::N --ipalloc-init seed=$SEED $PEERS

Where,

* `--no-restart` disables the Docker restart policy, since this will be
  handled by systemd.
* `--name` specifies a unique name for this new peer.
* `--ipalloc-init seed` specifies the names of only those peers that were
  involved in the initial cluster bootstrap - even if they have been
  subsequently removed from the cluster. You can view this as a kind
  of 'cluster identity',where  peers may only interoperate in the same
  cluster if they share the same seed.
* `$PEERS` is obtained from `/etc/sysconfig/weave` as described in the
  linked systemd documentation. For convenience, this may contain the
  address of the peer that is being launched, so that you don't have
  to compute separate lists of 'other' peers tailored to each peer -
  just supply the same complete list of peer addresses to every peer.

Note that unlike [Interactive](/site/operational-guide/interactive.md)
and [Uniform Fixed Cluster](/site/operational-guide/uniform-fixed-cluster.md) deployments
there is no `weave prime` step. You can add as many new peers in
parallel as you like, even under conditions of partition, and they
will all (eventually) join safely. This is ideal for use in
conjunction with asynchronous provisioning systems such as puppet or
chef. 

For maximum robustness, you should distribute an updated
`/etc/sysconfig/weave` file including the new peer to all existing
peers.

### Removing a Peer

On the peer to be removed:

    weave reset

You may remove a seed peer, as long as there is at least one other
seed peer left in the network.

Then, distribute an updated `/etc/sysconfig/weave` to the remaining
peers, omitting the removed peer from `$PEERS`.

On each remaining peer:

    weave forget <removed peer>

This step is not mandatory, but it will eliminate log noise and
spurious network traffic by stopping reconnection attempts.
