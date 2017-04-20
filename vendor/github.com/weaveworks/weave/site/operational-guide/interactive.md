---
title: Interactive Deployment
menu_order: 20
---
Weave Net can be launched interactively on the command line, and as 
long as Docker is configured to start on boot, the network will survive 
host reboots without the use of a systemd. However, since launching 
Weave Net commands in interactive mode is not amenable to automation 
and configuration management, it is recommended that deploying Weave Net
in this mode be reserved for exploration and evaluation only. 

### Bootstrapping

On the initial peer:

    weave launch

### Adding a Peer

On the new peer:

    weave launch <extant peers>

Where, 

* `<extant peers>` indicates all peers in the network, initial and
subsequently added, which have not been explicitly removed. It should
include peers that are temporarily offline or stopped.

To ensure that the new peer has joined the existing network, 
execute the following:

    weave prime

Before adding any new peers, you _must_ wait for this to complete. 
If this command waits and does not exit, it means that there is some
issue (such as a network partition or failed peers) that is preventing
a quorum from being reached - you will need to [address
that](/site/troubleshooting.md) before moving on.

### Stopping a Peer

A peer can be stopped temporarily with the following command:

    weave stop

A temporarily stopped peer will remember IP address allocation information on the
next `weave launch` but will forget any discovered peers or
modifications to the initial peer list that were made with `weave
connect` or `weave forget`. Note that if the host reboots, Docker
automatically restarts the peer.

### Removing a Peer

On the peer to be removed:

    weave reset

Then optionally on each remaining peer:

    weave forget <removed peer>

This step is not mandatory, but it will eliminate log noise and
spurious network traffic by stopping reconnection attempts and
preventing further connection attempts after a restart.
