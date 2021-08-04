# dns-controller

dns-controller creates DNS records.

## Motivation

In the bring-up of a new cluster, `protokube` has already ensured that 
we have an `etcd` cluster and an `apiserver`.  It also sets up DNS 
records for the `etcd` nodes (this is a much simpler problem, because 
we have a 1:1 mapping from an `etcd` node to a DNS name.)

However, none of the nodes can reach the API server to register.  Nor 
can end-users reach the API.  In future we might expose the API server 
as a normal service via `Type=LoadBalancer` or via a normal Ingress, 
but for now we just expose it via DNS.

## How it works

### Pods with hostNetworking

dns-controller can add DNS records that point to nodes *with hostNetworking* enabled.

The dns-controller recognizes annotations on pod.

* `dns.alpha.kubernetes.io/external` will set up records for accessing 
  the resource using the node's public IP.
* `dns.alpha.kubernetes.io/internal` will set up records for accessing 
  the resource using the node's private IP.

### Services

#### NodePort

The controller also recognizes these annotations on `NodePort` services:

* `dns.alpha.kubernetes.io/external` creates a Route53 A record with 
  `public` IPs of all the nodes
* `dns.alpha.kubernetes.io/internal` creates a Route53 A record with 
  `private` IPs of all the nodes

#### Loadbalancer

If _either_ of the two annotations are set on a `LoadBalancer` service, it will create a `CNAME` for the load balancer hostname _or_ it will create an A record if the load balancer has an IP.

### Ingress 

dns-controller can optionally watch `Ingress` resources. To enable this, you need to add the following to the cluster spec:
```
spec:
  externalDns:
    watchIngress: true
```

dns-controller will then map the specified ingress hostname and the `LoadBalancer` assigned to the ingress.
