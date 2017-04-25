---
title: Using Automatic Discovery With the Weave Net Proxy
menu_order: 30
---

Containers launched via the proxy use [weaveDNS](/site/weavedns.md)
automatically if it is running when they are started -
see the [weaveDNS usage](/site/weavedns.md#usage) section for an in depth
explanation of the behaviour and how to control it.

Typically, the proxy passes on container names as-is to weaveDNS
for registration. However, there are situations in which the final container
name may be out of your control (for example, if you are using Docker orchestrators which
append control/namespacing identifiers to the original container names).

For those situations, the proxy provides the following flags:

 * `--hostname-from-label<labelkey>`
 * `--hostname-match <regexp>`
 * `--hostname-replacement <replacement>`
 
When launching a container, the hostname is initialized to the
value of the container label using key `<labelkey>`. If no `<labelkey>` was
provided, then the container name is used. 

Additionally, the hostname is matched against a regular expression `<regexp>` and based on that match,
`<replacement>` is used to obtain the final hostname, and then handed over to weaveDNS for registration.

For example, you can launch the proxy using all three flags, as follows:

    host1$ weave launch-router && weave launch-proxy --hostname-from-label hostname-label --hostname-match '^aws-[0-9]+-(.*)$' --hostname-replacement 'my-app-$1'
    host1$ eval $(weave env)

>**Note:** regexp substitution groups must be pre-pended with a dollar sign
(for example, `$1`). For further details on the regular expression syntax see
[Google's re2 documentation](https://github.com/google/re2/wiki/Syntax).

After launching the Weave Net proxy with these flags, running a container named `aws-12798186823-foo` without labels results in weaveDNS registering the hostname `my-app-foo` and not `aws-12798186823-foo`.

    host1$ docker run -ti --name=aws-12798186823-foo weaveworks/ubuntu ping my-app-foo
    PING my-app-foo.weave.local (10.32.0.2) 56(84) bytes of data.
    64 bytes from my-app-foo.weave.local (10.32.0.2): icmp_seq=1 ttl=64 time=0.027 ms
    64 bytes from my-app-foo.weave.local (10.32.0.2): icmp_seq=2 ttl=64 time=0.067 ms

Also, running a container named `foo` with the label
`hostname-label=aws-12798186823-foo` leads to the same hostname registration.

    host1$ docker run -ti --name=foo --label=hostname-label=aws-12798186823-foo weaveworks/ubuntu ping my-app-foo
    PING my-app-foo.weave.local (10.32.0.2) 56(84) bytes of data.
    64 bytes from my-app-foo.weave.local (10.32.0.2): icmp_seq=1 ttl=64 time=0.031 ms
    64 bytes from my-app-foo.weave.local (10.32.0.2): icmp_seq=2 ttl=64 time=0.042 ms

This is because, as explained above, if providing `--hostname-from-label`
to the proxy, the specified label takes precedence over the container's name.

**See Also**

 * [Name resolution via `/etc/hosts`](/site/weave-docker-api/name-resolution-proxy.md)
 * [How Weave Finds Containers](/site/weavedns/how-works-weavedns.md)
