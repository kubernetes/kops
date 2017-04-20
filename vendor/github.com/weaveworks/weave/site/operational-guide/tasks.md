---
title: Administrative Tasks
menu_order: 60
---

The following administrative tasks are discussed: 

* [Configuring Weave Net to Start Automatically on Boot](#start-on-boot)
* [Detecting and Reclaiming Lost IP Address Space](#detect-reclaim-ipam)
* [Manually Reclaiming Lost Address Space](#manually-reclaim-address-space)
* [Upgrading a Cluster](#cluster-upgrade)
* [Resetting Persisted Data](#reset)


## <a name="start-on-boot"></a>Configuring Weave Net to Start Automatically on Boot

`weave launch` runs all of Weave Net's containers with a Docker restart
policy set to `always`.  If you have launched Weave Net manually at least
once and your system is configured to start Docker on boot, then Weave Net
will start automatically on system restarts.

If you are aiming for a non-interactive installation, use
[systemd](/site/installing-weave/systemd.md) or a similar init system to
launch Weave using the `--no-restart` flag after Docker has been started.

## <a name="detect-reclaim-ipam"></a>Detecting and Reclaiming Lost IP Address Space

The recommended method of removing a peer is to run `weave reset` on that
peer before the underlying host is decommissioned or repurposed. This
ensures that the portion of the IPAM allocation range assigned to the
peer is released for reuse. 

Under certain circumstances this operation may not be successful, 
or possible:

* If the peer is partitioned from the rest of the network
  when `weave reset` is executed on it
* If the underlying host is no longer available to execute `weave
  reset` due to a hardware failure or other unplanned termination (for
  example when using autoscaling with spot-instances that can be
  destroyed without notice)

In some cases you may already be aware of the problem, as you were
unable to execute `weave reset` successfully or because you know
through other channels that the host has died. In these cases you can
proceed straight to [Manually Reclaiming Lost Address Space](#reclaim-address-space).

However in some scenarios it may not be obvious that space has been
lost, in which case you can check for it periodically with the
following command on any peer:

    weave status ipam

This command displays the peer names and nicknames, absolute quantity/percentage of allocation 
range managed by peer and also identifies the names of unreachable peers. If you are satisfied
that the peer is truly gone, rather than temporarily unreachable due to a
partition, you can reclaim their space manually.

### <a name="manually-reclaim-address-space"></a>Manually Reclaiming Address Space

When a peer dies unexpectedly the remaining peers will consider its
address space to be unavailable even after it has remained unreachable
for prolonged periods. There is no universally applicable time limit
after which one of the remaining peers could decide unilaterally that
it is safe to appropriate the space for itself, and so an
administrative action is required to reclaim it.

The [`weave rmpeer`](/site/ipam/stop-remove-peers-ipam.md)
command is provided to perform this task, and must
be executed on _one_ of the remaining peers. That peer will then take
ownership of the freed address space.

## <a name="cluster-upgrade"></a>Upgrading a Cluster

Protocol versioning and feature negotiation are employed in Weave Net
to enable incremental rolling upgrades. Each major maintains
the ability to speak to the preceding major release at a minimum, and
connected peers only utilize features which both support. 

The general upgrade procedure is as follows:

On each peer:

* Stop the old Weave Net with `weave stop` (or `systemctl stop weave` if
  you're using a systemd unit file)
* Download the new Weave Net script and replace the existing one
* Start the new Weave with `weave launch <existing peer list>` (or
  `systemctl start weave` if you're using a systemd unit file)

To minimize downtime while the new script is pulling the new container images:

* Download the new Weave Net script to a temporary location, for example,
  `/path/to/new/weave`
* Pull the new images with `/path/to/new/weave setup`
* Stop the old Weave Net with `weave stop` (or `systemctl stop weave` if
  you're using a systemd unit file)
* Replace the existing script with the new one
* Start the new Weave Net with `weave launch <existing peer list>` (or
  `systemctl start weave` if you're using a systemd unit file)

>>**Note:** Always check the Release Notes for specific versions in case
there are any special caveats or deviations from the standard
procedure.

## <a name="reset"></a>Resetting Persisted Data

Weave Net persists information in a data volume container named
`weavedb`. If you wish to start from a completely clean slate (for
example to withdraw a peer from one network and join it to another)
you can issue the following command:

    weave reset
    

**See Also**

 * [Allocating IP Addresses](/site/ipam.md)
 * [Troubleshooting the IP Allocator](/site/ipam/troubleshooting-ipam.md)

