# dns-controller

dns-controller creates DNS records.

In the bring-up of a new cluster, `protokube` has already ensured that 
we have an `etcd` cluster and an `apiserver`.  It also sets up DNS 
records for the `etcd` nodes (this is a much simpler problem, because 
we have a 1:1 mapping from an `etcd` node to a DNS name.)

However, none of the nodes can reach the API server to register.  Nor 
can end-users reach the API.  In future we might expose the API server 
as a normal service via `Type=LoadBalancer` or via a normal Ingress, 
but for now we just expose it via DNS.

The dns-controller recognizes annotations on nodes.

* `dns.alpha.kubernetes.io/external` will set up records for accessing 
  the resource externally
* `dns.alpha.kubernetes.io/internal` will set up records for accessing 
  the resource internally

When added on `Service` controllers:

* `dns.alpha.kubernetes.io/external` creates a Route53 A record with 
  `public` IPs of all the nodes
* `dns.alpha.kubernetes.io/internal` creates a Route53 A record with 
  `private` IPs of all the nodes

The syntax is a comma separated list of fully qualified domain names.
