---
title: Troubleshooting and Present Limitations
menu_order: 50
---




### <a name="troubleshooting"></a>Troubleshooting

The command:

    weave status

reports on the current status of various Weave Net components, including
DNS:

```
...
       Service: dns
        Domain: weave.local.
      Upstream: 8.8.8.8, 8.8.4.4
           TTL: 1
       Entries: 9
...
```

The first section covers the router; see the [troubleshooting
guide](/site/troubleshooting.md#weave-status) for more details.

The 'Service: dns' section is pertinent to weaveDNS, and includes:

* The local domain suffix which is being served
* The list of upstream servers used for resolving names not in the local domain
* The response ttl
* The total number of entries

You may also use `weave status dns` to obtain a [complete
dump](/site/troubleshooting.md#weave-status-dns) of all DNS registrations.

Information on the processing of queries, and the general operation of
weaveDNS, can be obtained from the container logs with

    docker logs weave

### <a name="limitations"></a>Present Limitations

 * The server will not know about restarted containers, but if you
   re-attach a restarted container to the weave network, it will be
   re-registered with weaveDNS.
 * The server may give unreachable IPs as answers, since it doesn't
   try to filter by reachability. If you use subnets, align your
   hostnames with the subnets.
