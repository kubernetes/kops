---
title: Using The Weave Docker API Proxy
menu_order: 10
---


When containers are created via the Weave Net proxy, their entrypoint is 
modified to wait for the Weave network interface to become
available. 

When they are started via the Weave Net proxy, containers are 
[automatically assigned IP addresses](/site/ipam.md) and connected to the
Weave network.  

### Creating and Starting Containers with the Weave Net Proxy

To create and start a container via the Weave Net proxy run:

    host1$ docker run -ti weaveworks/ubuntu

or, equivalently run:

    host1$ docker create -ti --name=foo weaveworks/ubuntu
    host1$ docker start foo

Specific IP addresses and networks can be supplied in the `WEAVE_CIDR`
environment variable, for example:

    host1$ docker run -e WEAVE_CIDR=10.2.1.1/24 -ti weaveworks/ubuntu

Multiple IP addresses and networks can be supplied in the `WEAVE_CIDR`
variable by space-separating them, as in
`WEAVE_CIDR="10.2.1.1/24 10.2.2.1/24"`.


### Returning Weave Network Settings Instead of Docker Network Settings

The Docker NetworkSettings (including IP address, MacAddress, and
IPPrefixLen), are still returned when `docker inspect` is run. If you want
`docker inspect` to return the Weave NetworkSettings instead, then the
proxy must be launched using the `--rewrite-inspect` flag. 

This command substitutes the Weave network settings when the container has a
Weave Net IP. If a container has more than one Weave Net IP, then the inspect call
only includes one of them.

    host1$ weave launch-router && weave launch-proxy --rewrite-inspect

### Multicast Traffic and Launching the Weave Proxy

By default, multicast traffic is routed over the Weave network.
To turn this off, for example, because you want to configure your own multicast
route, add the `--no-multicast-route` flag to `weave launch-proxy`.

### Other Weave Proxy options

 * `--without-dns` -- stop telling containers to use [WeaveDNS](/site/weavedns.md)
 * `--log-level=debug|info|warning|error` -- controls how much
   information to emit for debugging
 * `--no-restart` -- remove the default policy of `--restart=always`, if
   you want to control start-up of the proxy yourself

**See Also**

 * [Setting Up The Weave Docker API Proxy](/site/weave-docker-api.md)
 * [Securing Docker Communications With TLS](/site/weave-docker-api/securing-proxy.md)
