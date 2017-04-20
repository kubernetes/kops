---
title: Troubleshooting the IP Allocator
menu_order: 30
---


The command

    weave status

reports on the current status of the weave router and IP allocator:

```
...
       Service: ipam
        Status: awaiting consensus (quorum: 2, known: 0)
         Range: 10.32.0.0-10.47.255.255
 DefaultSubnet: 10.32.0.0/12
...
```

The first section covers the router; see the [troubleshooting
guide](/site/troubleshooting.md#weave-status) for full details.

The 'Service: ipam' section displays the consensus state as well as
the total allocation range and default subnet. Columns are as follows:

* 'Status' - allocator state
    * 'idle' - no allocation requests or claims have been made yet;
      consensus is deferred until then
    * 'awaiting consensus' - an attempt to achieve consensus is
      ongoing, triggered by an allocation or claim request;
      allocations will block.  This state persists until a quorum of
      peers are able to communicate amongst themselves successfully.
    * 'priming' - peer is an observer and is waiting to receive IPAM
      data from seeding or consensus elsewhere in the network
    * 'ready' - consensus achieved; allocations proceed normally
    * 'waiting for IP range grant from peers' - peer has exhausted its
      agreed portion of the range and is waiting to be granted some
      more
    * 'all IP ranges owned by unreachable peers' - peer has exhausted
      its agreed portion of the range but cannot reach anyone to ask
      for more
* 'Range' - total allocation range set by `--ipalloc-range`
* 'DefaultSubnet' - default subnet set by `--ipalloc-default-subnet`

Information regarding the division of the IP allocation range amongst
peers and their reachability can be obtained with

```
$ weave status ipam
00:00:00:00:00:01(one)      349526 IPs (33.3% of total)
00:00:00:00:00:02(two)      349525 IPs (33.3% of total)
00:00:00:00:00:03(three)    349525 IPs (33.3% of total) - unreachable!
```

Columns are as follows:

* Peer Name and Nickname
* Absolute quantity/percentage of allocation range managed by peer
* Indication of unreachability. This means that the peer is not
  visible (directly or indirectly) to the peer on which `weave status
  ipam` was run; whilst this could be a transient condition due to a
  partition, it may be because the peer has failed and needs to be
  removed administratively - see [Starting, Stopping and Removing
  Peers](/site/ipam/stop-remove-peers-ipam.md) for more details.
