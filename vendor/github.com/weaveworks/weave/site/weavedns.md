---
title: Discovering Containers with WeaveDNS
menu_order: 80
---



WeaveDNS is a DNS server that answers name queries on a Weave network
and provides a simple way for containers to find each other. Just give
the containers hostnames and then tell other containers to connect to
those names.  Unlike Docker 'links', this requires no code changes and
it works across hosts.

WeaveDNS is deployed as an embedded service within the Weave router.
The service is automatically started when the router is launched:

```
host1$ weave launch
host1$ eval $(weave env)
```

WeaveDNS related configuration arguments can be passed to `launch`.

Application containers use weaveDNS automatically if it is
running when they are started. They use it for name
resolution, and will register themselves if they either have a
hostname in the weaveDNS domain (`weave.local` by default) or are given an explicit container name:

```
host1$ docker run -dti --name=pingme weaveworks/ubuntu
host1$ docker run -ti  --hostname=ubuntu.weave.local weaveworks/ubuntu
root@ubuntu:/# ping pingme
...
```

> **Note** If both hostname and container name are specified at
the same time, the hostname takes precedence. In this circumstance, if
the hostname is not in the weaveDNS domain, the container is *not*
registered, but it will still use weaveDNS for resolution.

To disable an application container's use of weaveDNS, add the
`--without-dns` option to `weave launch-proxy`. To
disable weaveDNS itself, launch weave with the `--no-dns` option.

**See Also**

 * [Integrating Docker via the API Proxy](/site/weave-docker-api.md)
 * [Load Balancing and Fault Resilience with WeaveDNS](/site/weavedns/load-balance-fault-weavedns.md)
