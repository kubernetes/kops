---
title: WeaveDNS (service discovery) Design Notes
layout: default
---

# WeaveDNS (service discovery) Design Notes

The model is that each host has a service that is notified of
hostnames and weave addresses for containers on the host.  Like IPAM,
this service is embedded within the router.  It binds to
the host bridge to answer DNS queries from local containers; for
anything it can't answer, it uses the information in the host's
/etc/resolv.conf to query an 'fallback' server.

The service is comprised of a DNS server, which answers all DNS queries
from containers, and a in-memory database of hostnames and IPs.  The
database on each node contains a complete copy of the hostnames and IPs
for every containers in the cluster.

For hostname queries in the local domain (default weave.local), the DNS
server will consult the in-memory database.  For reverse queries, we
first consult the local database, and if not found we query the
upstream server.  For all other queries, we consult the upstream
server.

Updates to the in-memory database are broadcast to other DNS servers
within the cluster.  The in-memory database only contains entries from
connected DNS servers; if a DNS server becomes partitioned from the
cluster, entries belonging to that server are removed from each node in
the cluster.  When the partitioned DNS server reconnects, the entries
are re-broadcast around the cluster.

The DNS server also listens to the Docker event stream, and removes
entries for containers when they die.  Entries removed in this way are
tombstoned, and the tombstone lazily broadcast around the cluster.
After a short timeout the tombstones are independently removed from
each host.


## DNS server API

The DNS server accepts HTTP requests on the following URL (patterns)
and methods:

`PUT /name/<identifier>/<ip-address>`

Put a record for an IP, bound to a host-scoped identifier (e.g., a
container ID), in the DNS database.  The request body must contain
a `fqdn=foo.weave.local` key pair.

`DELETE /name/<identifier>/<ip-address>`

Remove a specific record for an IP and host-scoped identifier. The request
body can optionally contain a `fqdn=foo.weave.local` key pair.

`DELETE /name/<identifier>`

Remove all records for the host-scoped identifier.

`GET /name/<fqdn>`

List of all IPs (in JSON format) for givne FQDN.

## DNS updater

The updater component uses the Docker remote API to monitor containers
coming and going, and tells the DNS server to update its records via
its HTTP interface. It does not need to be attached to the weave
network.

The updater starts by subscribing to the events, and getting a list of
the current containers. Any containers given a domain ending with
".weave" are considered for inclusion in the name database.

When it sees a container start or stop, the updater checks the weave
network attachment of the container, and updates the DNS server.

> How does it check the network attachment from within a container?

> Will it need to delay slightly so that `attach` has a chance to run?
> Perhaps it could put containers on a watch list when it's noticed
> them.

