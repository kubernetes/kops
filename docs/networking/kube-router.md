# Kube-router

[Kube-router](https://github.com/cloudnativelabs/kube-router) is project that provides one cohesive solution that provides CNI networking for pods, an IPVS based network service proxy and iptables based network policy enforcement.

Kube-router also provides a service proxy, so kube-proxy will not be deployed in to the cluster.

## Installing kube-router on a new Cluster

The following command sets up a cluster with Kube-router.

```sh
export ZONES=mylistofzones
kops create cluster \
  --zones $ZONES \
  --networking kube-router \
  --yes \
  --name myclustername.mydns.io
```

## Configuration

No additional configurations are required to be done by user. Kube-router automatically disables source-destination check on all AWS EC2 instances. For the traffic within a subnet there is no overlay or tunneling used. For cross-subnet pod traffic ip-ip tunneling is used implicitly and no configuration is required.

### Enforcing network policies with nftables

By default, kube-router enforces `NetworkPolicy` objects with iptables and ipsets. Setting `useNFTablesForNetpol` makes its network policy controller use nftables instead:

```yaml
spec:
  networking:
    kuberouter:
      useNFTablesForNetpol: true
```

This only affects the network policy controller; the service proxy and the router controller keep their existing implementations. Upstream considers the nftables backend [experimental](https://github.com/cloudnativelabs/kube-router/blob/master/docs/nftables.md), with no benchmarks against the iptables implementation, so it is not recommended for production clusters yet.