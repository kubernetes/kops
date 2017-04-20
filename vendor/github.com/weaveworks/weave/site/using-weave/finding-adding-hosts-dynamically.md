---
title: Adding and Removing Hosts Dynamically
menu_order: 90
---

To add a host to an existing Weave network, launch Weave Net on the
host, and supply the address of at least one host. Weave Net
automatically discovers any other hosts in the network and establishes
connections with them if it can (in order to avoid unnecessary
multi-hop routing).

In some situations existing Weave Net hosts may be unreachable from
the new host due to firewalls, etc.  However, it is still possible to
add the new host, provided that inverse connections, for example, from
existing hosts to the new hosts, are available.

To accomplish this, launch Weave Net onto the new host without
supplying any additional addresses and then, from one of the existing
hosts run:

    host# weave connect $NEW_HOST

Any other existing hosts on the Weave network will attempt to
establish connections to the new host as well.

### Instructing Peers to Forget a Host

To instruct a peer to forget a particular host specified to it via
`weave launch` or `weave connect` run:

    host# weave forget $DECOMMISSIONED_HOST

This prevents the peer from reconnecting to that host once
connectivity to it is lost, and can be used to administratively remove
any decommissioned peers from the network.

### Bulk Replacing Hosts

Hosts can also be bulk-replaced. All existing hosts will be forgotten,
and the new hosts added:

    host# weave connect --replace $NEW_HOST1 $NEW_HOST2

### Restarting Docker and Weave Net

If Weave Net is restarted by Docker it automatically remembers any
previous connect and forget operations, however if you stop it
manually and launch it again, it will not remember any prior connects.
If you want to launch again and retain the results of those operations
use `--resume`:

    host# weave launch --resume

> **Note:** In this case, you cannot specify a list of addresses,
> since the previous peer list is used exclusively.

For complete control over the peer topology, disable automatic
discovery using the `--no-discovery` option with `weave launch`.

If discovery if disabled, Weave Net only connects to the addresses
specified at launch time and with `weave connect`.

To return a list of all hosts and their peer connections established
with `weave launch` and `weave connect` run:

    host# weave status targets

**See Also**

 * [Enabling Multi-Cloud, Multi-Hop Networking and Routing](/site/using-weave/multi-cloud-multi-hop.md)
 * [Stopping and Removing Peers](/site/ipam/stop-remove-peers-ipam.md)
