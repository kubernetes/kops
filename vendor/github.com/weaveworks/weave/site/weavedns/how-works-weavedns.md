---
title: How Weave Finds Containers
menu_order: 10
---


The weaveDNS service running on every host acts as the nameserver for
containers on that host. It learns about hostnames for local containers
from the proxy and from the `weave attach` command.  

If a hostname is in the `.weave.local` domain, then weaveDNS records the association of that
name with the container's Weave Net IP address(es) in its in-memory
database, and then broadcasts the association to other Weave Net peers in the
cluster.

When weaveDNS is queried for a name in the `.weave.local` domain, it
looks up the hostname in its memory database and responds with the IPs
of all containers for that hostname across the entire cluster.

When weaveDNS is queried for a name in a domain other than
`.weave.local`, it queries the host's configured nameserver, which is
the standard behaviour for Docker containers.

### Specifying a Different Docker Bridge Device

So that containers can connect to a stable and always routable IP
address, weaveDNS listens on port 53 to the Docker bridge device, which
is assumed to be `docker0`.  Some configurations may use a different
Docker bridge device. To supply a different bridge device, use the
environment variable `DOCKER_BRIDGE`, e.g.,

```
$ sudo DOCKER_BRIDGE=someother weave launch
```

In the event that weaveDNS is launched in this way, it's important that
other calls to `weave` also specify the bridge device:

```
$ sudo DOCKER_BRIDGE=someother weave attach ...
```

**See Also**

 * [Using WeaveDNS](/site/weavedns.md)
 * [Load Balancing with weaveDNS](/site/weavedns/load-balance-fault-weavedns.md)
 * [Managing Domain Entries](/site/weavedns/managing-domains-weavedns.md)
