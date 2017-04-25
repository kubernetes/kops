---
title: IP Addresses, Routes and Networks
menu_order: 50
---


Weave Net runs containers on a private network, which means that IP addresses are isolated from the rest of the
Internet, and that you don't have to worry about addresses clashing. 

You can of course also manually change the IP of any given container or subnet on a Weave network.  See, [How to Manually Specify IP Addresses and Subnets](/site/using-weave/manual-ip-address.md)

### Some Definitions

- _IP_ is the Internet Protocol, the fundamental basis of network
   communication between billions of connected devices.
- The _IP address_ is (for most purposes) the four numbers separated
  by dots, like `192.168.48.12`. Each number is one byte in size, so can
  be between 0 and 255.
- Each IP address lives on a _Network_, which is some set of those
  addresses that all know how talk to each other. The network address
  is some prefix of the IP address, like `192.168.48`. To show
  which part of the address is the network, we append a slash
  and then the number of bits in the network prefix, like
  `/24`.
- A _route_ is an instruction for how to deal with traffic destined
  for somewhere else - it specifies a Network, and a way to talk to
  that network.  Every device using IP has a table of routes, so for
  any destination address it looks up that table, finds the right
  route, and sends it in the direction indicated.

### IP Address Notation in Weave

In the IP address `10.4.2.6/8`, the network prefix is the first 8 bits
- `10`. Written out in full, that network is `10.0.0.0/8`.

The most common prefix lengths are 8, 16 and 24, but there is nothing
stopping you using a /9 network or a /26. For example, `6.250.3.1/9` is on the
`6.128.0.0/9` network.

Several websites offer calculators to decode this kind of address, see, for example: [IP Address Guide](http://www.ipaddressguide.com/cidr).

The following is an example route table for a container that is attached to a Weave
network:

    # ip route show
    default via 172.17.42.1 dev eth0 
    10.2.2.0/24 dev ethwe  proto kernel  scope link  src 10.2.2.1 
    172.17.0.0/16 dev eth0  proto kernel  scope link  src 172.17.0.170 

It has two interfaces: one that Docker gave it called `eth0`, and one
that weave gave it called `ethwe`. They are on networks
`172.17.0.0/16` and `10.2.2.0/24` respectively, and if you want to
talk to any other address on those networks then the routing table
tells it to send directly down that interface. If you want to talk to
anything else not matching those rules, the default rule says to send
it to `172.17.42.1` down the eth0 interface.

So, suppose this container wants to talk to another container at
address `10.2.2.9`; it will send down the ethwe interface and weave
Net will take care of routing the traffic. To talk an external server
at address `74.125.133.128`, it looks in its routing table, doesn't
find a match, so uses the default rule.

**See Also**

 * [Allocating IPs in a Specific Range](/site/using-weave/configuring-weave.md)
