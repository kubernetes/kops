# Overlay Method Selection

Weave automatically selects the best working overlay implementation
when connecting to a new peer. It does this by initiating both
forwarders in parallel, eventually settling on the most preferred
established forwarder at which point the other forwarder is shut down,
assuming it had not failed already. In the event that the remaining
forwarder fails at a later time (e.g. because of a heartbeat timeout)
the control plane TCP connection will recycle allowing the selection
algorithm to run again. By default fast datapath is preferred over
Sleeve.

# Local Bridging

Unlike the weave bridge netdev used by Sleeve, the OVS datapath has no
inherent bridging capability. Consequently fastdp implements an
ethernet bridge in addition to the vxlan overlay, maintaining its own
forwarding database by learning MAC addresses and dispatching
broadcast, unicast & multicast traffic accordingly.

# Short Peer IDs

The router relies on having access to the source and destination peer
when deciding how to forward packets. When the Sleeve overlay is in
use this information is conveyed directly within the encapsulation by
including the names of the peers in question, a solution not
accommodated directly by the vxlan wire format. Fortunately vxlan does
have a twenty four bit segment ID field in the header we can use to
encode this data - the challenge is to identify peers uniquely with a
twelve bit identifier instead of the seventeen bytes used by Sleeve.
In practice this is achieved by peers adopting a random 'short ID' and
resolving ownership collisions of same via the existing gossip
mechanism.

# Heartbeats via vxlan

Heartbeats are implemented by sending an encapsulated ethernet frame
with source and destination MAC addresses of `00:00:00:00:00:00` and a
payload consisting of the connection UID and total ethernet frame
length. The vxlan vport miss handler on the receiving side detects the
all-zero addresses and acknowledges the heartbeat via the TCP control
channel after validating the connection UID and frame length against
their expected values.

# PMTU Discovery

The length of the heartbeat ethernet frame is set to the MTU of the
datapath interface. In the event the vxlan packet is dropped or
truncated, the heartbeat will not be acknowledged; this lack of
acknowledgement will cause the peers to fall back to the Sleeve
overlay, which has a more sophisticated dynamic mechanism for coping
with low path MTUs.

To avoid triggering this fallback in typical deployments, the datapath
interface is statically configured with an MTU of 1376 bytes allowing
it to work with most underlay network provider MTUs, including GCE at
1460 bytes (the eighty four byte difference accommodates the encrypted vxlan overhead).
This value can be overridden by setting `WEAVE_MTU` at launch if
necessary.

# Virtual Ports

There are three kinds of virtual port associated with the weave
datapath:

* `internal` - one of these is created automatically, named after the
  datapath. It corresponds to the network interface of the same name
  that appears on the host when the datapath is created; it is the
  ingress/egress port for `weave expose`
* `netdev` - one for each application container, corresponding to the
  host end of the veth pair
* `vxlan` - one for each UDP port on which we are listening for vxlan
  encapsulated packets. Typically there is only one of these, but
  there may be more in a network in which peers do not all use the same
  port configuration

# Misses and Flow Creation

As mentioned above, the datapath has no inherent behaviour - any
ingressing packet which does not match a flow is passed to a userspace
miss handler. The miss handler is then responsible for a) instructing
the datapath to take actions for the packet that caused the miss (for
example by copying it to one or more ports) and optionally b)
installing flow rules which will allow the datapath to act on similar
packets in future without invoking the miss handler.

At a high level the fast datapath implementation can be viewed as a
set of miss handlers that determine what actions the OVS datapath
should take based on router topology information, together with some
additional machinery that manages the expiry of resulting flows when
that state changes.

When a miss handler is invoked, it has two pieces of context: a byte
array containing the packet that triggered the miss, and a set of
'flow keys' that have been extracted from the packet by the kernel
module. The following flow keys are of interest to weave:

* `InPort` - the identifier of the ingress virtual port
* `Ethernet` - the source and destination MAC addresses of the
  ethernet frame
* `Tunnel` - tunnel identifier (see section on short peer IDs above),
  and source/dest IPv4 addresses. Only present for ingress via a vxlan
  vport

Crucially, the router must use this information alone to determine
which actions to take - this allows the specification of a flow which
matches these keys and instructs the kernel to take action
automatically in future without further userspace involvement.

Two actions are of interest:

* `Output` - output the packet to the specified vport
* `SetTunnel` - update the effective vxlan tunnel parameters. Only
  required prior to output from a vxlan vport

It is possible to have multiple actions in a flow, so the router can
for example create a single rule that copies broadcast traffic to all
'local' vports (save the ingress vport obviously) as well as to vxlan
vports for onward peers as dictated by the routing topology.

Every flow specified by the router has the following characteristics:

* An `InPort` key matching the ingress vport
* A `Tunnel` key if the ingress vport is of type vxlan
* An `Ethernet` key matching the source and destination MAC
* A list of `Output` (for internal and netdev vports) and
  `SetTunnel`+`Output` (for vxlan vports) actions

Under certain circumstances the miss handler will instruct the
datapath to execute a set of actions for a packet without creating a
corresponding flow:

* When broadcasting packets to discover the path to unknown
  destination MACs. This condition is transient in the overwhelming
  majority of cases and so the benefit of creating a flow is
  outweighed by the need to invalidate it once the MAC is learned
* When the packet also needs to be forwarded to one or more peers via
  Sleeve. In this case the router needs to handle all subsequent
  matching packets as there is no OVS flow action to do it without
  userspace involvement

# Flow Invalidation

Once a flow has been created for a particular combination of keys, the
miss handler will never again be invoked for matching packets. It is
therefore extremely important that we detect events which invalidate
existing flow actions:

* Addition of netdev vports to the datapath
* Route invalidation (topology change)
* Short peer ID collision

The response in each case is the same - delete all flows from the
datapath, allowing them to be recreated taking into account the
updated state.

In addition to these event based invalidations there is an expiry
process that executes every five minutes. This process enumerates all
flows in the datapath, removing any which have not been used since the
last check; this cleans up:

* Flows referring to netdev vports which have been removed
* Flows created by forwarders which have been stopped
* Flows related to MAC addresses which have not communicated recently

Flows keyed on tunnel IPv4 address do not need to be cleared when a
peer appears to change IP address due to NAT; this will cause a miss
resulting in a new flow, and the old flow will expire naturally via
the timer mechanism.

