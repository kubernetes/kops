---
title: Enabling Multi-Cloud, Multi-Hop Networking and Routing
menu_order: 100
---


### Enabling Multi-Cloud Networking

Before multi-cloud networking can be enabled, you must configure the network to allow 
connections through Weave Net's control and data ports on the Docker hosts. By default, the 
control port defaults to TCP 6783, and the data ports to 
UDP 6783/6784. 

To override Weave Net’s default ports, specify a port using 
the `WEAVE_PORT` setting. For example, if WEAVE_PORT is 
set to `9000`, then Weave uses TCP 9000 for its control 
port and UDP 9000/9001 for its data port. 

>**Important!** It is recommended that all peers be given 
the same setting.


### Multi-hop routing

A network of containers across more than two hosts can be 
established even when there is only partial connectivity 
between the hosts. 

Weave Net routes traffic between containers as long as 
there is at least one *path* of connected hosts 
between them.

For example, if a Docker host in a local data center can 
connect to hosts in GCE and EC2, but the latter two cannot 
connect to each other, containers in the latter two can 
still communicate and Weave Net in this instance will route the 
traffic via the local data center.

**See Also** 

 * [Finding and Adding Hosts Dynamically](/site/using-weave/finding-adding-hosts-dynamically.md)


