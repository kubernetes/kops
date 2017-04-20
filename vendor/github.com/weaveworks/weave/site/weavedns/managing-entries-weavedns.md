---
title: Managing Domain Entries
menu_order: 40
---


The following topics are discussed: 

* [Adding and removing extra DNS entries](#add-remove)
* [Resolving WeaveDNS entries from the Host](#resolve-weavedns-entries-from-host)
* [Hot-swapping Service Containers](#hot-swapping)
* [Retaining DNS Entries When Containers Stop](#retain-stopped)
* [Configuring a Custom TTL](#ttl)



### <a name="add-remove"></a>Adding and Removing Extra DNS Entries

If you want to give the container a name in DNS *other* than its
hostname, you can register it using the `dns-add` command. For example:

```
$ C=$(docker run -ti weaveworks/ubuntu)
$ weave dns-add $C -h pingme2.weave.local
```

You can also use `dns-add` to add the container's configured hostname
and domain simply by omitting `-h <fqdn>`, or specify additional IP
addresses to be registered against the container's hostname e.g.
`weave dns-add 10.2.1.27 $C`.

The inverse operation can be carried out using the `dns-remove`
command:

```
$ weave dns-remove $C
```

By omitting the container name it is possible to add/remove DNS
records that associate names in the weaveDNS domain with IP addresses
that do not belong to containers, e.g. non-weave addresses of external
services:
```
$ weave dns-add 192.128.16.45 -h db.weave.local
```

Note that such records get removed when stopping the weave peer on
which they were added.

### <a name="resolve-weavedns-entries-from-host"></a>Resolving WeaveDNS Entries From the Host

You can resolve entries from any host running weaveDNS with `weave
dns-lookup`:

    host1$ weave dns-lookup pingme
    10.40.0.1

### <a name="hot-swapping"></a>Hot-swapping service containers

If you would like to deploy a new version of a service, keep the old
one running because it has active connections but make all new
requests go to the new version, then you can simply start the new
server container and then [remove](#add-remove) the entry for the old
server container. Later, when all connections to the old server have
terminated, stop the container as normal.

### <a name="ttl"></a>Configuring a custom TTL

By default, weaveDNS specifies a TTL of 30 seconds in responses to DNS
requests.  However, you can force a different TTL value by launching
weave with the `--dns-ttl` argument:

```
$ weave launch --dns-ttl=10
```

This will shorten the lifespan of answers sent to clients, so you will
be effectively reducing the probability of them having stale
information, but you will also be increasing the number of request this
weaveDNS instance will receive.

**See Also**

 * [How Weave Finds Containers](/site/how-works-weavedns.md)
 * [Load Balancing and Fault Resilience with WeaveDNS](/site/weavedns/load-balance-fault-weavedns.md)
 * [Managing Domains](/site/weavedns/managing-domains-weavedns.md)
