---
title: Managing Services - Exporting, Importing, Binding and Routing
menu_order: 70
---

This section contains the following topics: 

 * [Exporting Services](#exporting)
 * [Importing Services](#importing)
 * [Binding Services](#binding)
 * [Routing Services](#routing)
 * [Dynamically Changing Service Locations](#change-location)



### <a name="exporting"></a>Exporting Services

Services running in containers on a Weave network can be made
accessible to the outside world (and, more generally, to other networks)
from any Weave Net host, irrespective of where the service containers are
located.

Returning to the [netcat example service](/site/using-weave.md), 
you can expose the netcat service running on `HOST1` and make it accessible to the outside world via `$HOST2`. 

First, expose the application network to `$HOST2`, as explained in [Integrating with the Host Network](/site/using-weave/host-network-integration.md):

    host2$ weave expose
    10.2.1.132

Then add a NAT rule that routes the traffic from the outside world to the destination container service.

    host2$ iptables -t nat -A PREROUTING -p tcp -i eth0 --dport 2211 \
           -j DNAT --to-destination $(weave dns-lookup a1):4422

In this example, it is assumed that the "outside world" is connecting to `$HOST2` via 'eth0'. The TCP traffic to port 2211 on the external IPs will be routed to the 'nc' service running on port 4422 in the container a1.

With the above in place, you can connect to the 'nc' service from anywhere using:

    echo 'Hello, world.' | nc $HOST2 2211

>**Note:** Due to the way routing is handled in the Linux kernel, this won't work when run *on* `$HOST2`.

Similar NAT rules to the above can be used to expose services not just to the outside world but also to other, internal, networks.

### <a name="importing"></a>Importing Services

Applications running in containers on a Weave network can be given access to services, which are only reachable from certain 
Weave hosts, regardless of where the actual application containers are located.

Expanding on the [netcat service example](/site/using-weave.md), you now decide to add a third, non-containerized, netcat service. This additional netcat service runs on `$HOST3`, and listens on port 2211, but it is not on the Weave network. 

An additional caveat is that `$HOST3` can only be reached from `$HOST1`, which is not accessible via `$HOST2`. Nonetheless, you still need to make the `$HOST3` service available to an application that is running in a container on `$HOST2`.

To satisfy this scenario, first [expose the application network to the host](/site/using-weave/host-network-integration.md) by running the following on `$HOST1`: 

    host1$ weave expose -h host1.weave.local
    10.2.1.3

Then add a NAT rule, which routes from the above IP to the destination service.

    host1$ iptables -t nat -A PREROUTING -p tcp -d 10.2.1.3 --dport 3322 \
           -j DNAT --to-destination $HOST3:2211

This allows any application container to reach the service by connecting to 10.2.1.3:3322. So if `$HOST3` is running a 
netcat service on port 2211:

    host3$ nc -lk -p 2211

You can now connect to it from the application container running on `$HOST2` using:

    root@a2:/# echo 'Hello, world.' | nc host1 3322

Note that you should be able to run this command from any application container.

### <a name="binding"></a>Binding Services

Importing a service provides a degree of indirection that allows late and dynamic binding, similar to what can be achieved with a proxy. 

Referring back to the [netcat services example that is running on three hosts](#importing), the application containers are completely unaware that the service they are accessing at `10.2.1.3:3322` actually resides on `$HOST3:2211`. 

You can point application containers to another service location by changing the above NAT rule, without altering the applications.

### <a name="routing"></a>Routing Services

You can combine the service export and service import features to establish connectivity between applications and services residing on disjointed networks, even if those networks are separated by firewalls and have overlapping IP ranges. 

Each network imports its services into Weave Net, while at the same time, exports from Weave Net any services that are required by its applications. In this scenario, there are no application containers (although, there could be). Weave Net is acting as an address translation and routing facility, and uses the Weave container network as an intermediary.

Expanding on the [netcat example](/site/using-weave.md), you can also import an additional netcat service running on `$HOST3` into Weave Net via `$HOST1`. 

Begin importing the service onto `$HOST2` by first exposing the application network:

    host2$ weave expose
    10.2.1.3

Then add a NAT rule which routes traffic from the `$HOST2` network (for example, anything that can connect to `$HOST2`) to the service endpoint on the Weave network:

    host2$ iptables -t nat -A PREROUTING -p tcp -i eth0 --dport 4433 \
           -j DNAT --to-destination 10.2.1.3:3322

Now any host on the same network as `$HOST2` is able to access the service:

    echo 'Hello, world.' | nc $HOST2 4433

### <a name="change-location"></a>Dynamically Changing Service Locations

Furthermore, as explained in Binding Services, service locations can be dynamically altered without having to change any of the applications that access them.  

For example, you can move the netcat service to `$HOST4:2211`  and it will retain its 10.2.1.3:3322 endpoint on the Weave network.


**See Also**

 * [Adding and Removing Hosts Dynamically](/site/using-weave/finding-adding-hosts-dynamically.md)
 * [Enabling Multi-Cloud, Multi-Hop Networking and Routing](/site/using-weave/multi-cloud-multi-hop.md)

