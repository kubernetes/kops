---
title: Name resolution via `/etc/hosts`
menu_order: 40
---


When starting Weave Net enabled containers, the proxy automatically
replaces the container's `/etc/hosts` file, and disables Docker's control
over it. The new file contains an entry for the container's hostname
and Weave Net IP address, as well as additional entries that have been
specified using the `--add-host` parameters. 

This ensures that:

- name resolution of the container's hostname, for example, via `hostname -i`,
returns the Weave Net IP address. This is required for many cluster-aware
applications to work.
- unqualified names get resolved via DNS, for example typically via weaveDNS
to Weave Net IP addresses. This is required so that in a typical setup
one can simply "ping `<container-name>`", i.e. without having to
specify a `.weave.local` suffix.

If you prefer to keep `/etc/hosts` under Docker's control (for
example, because you need the hostname to resolve to the Docker-assigned
IP instead of the Weave IP, or you require name resolution for
Docker-managed networks), the proxy must be launched using the
`--no-rewrite-hosts` flag.

    host1$ weave launch-router && weave launch-proxy --no-rewrite-hosts
    
**See Also**

 * [Using Automatic Discovery With the Weave Proxy](/site/weave-docker-api/automatic-discovery-proxy.md)    
    
