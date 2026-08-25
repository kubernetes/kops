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

### Restricting externalIPs and loadBalancerIPs

Since v2.8.0 kube-router defaults `--strict-external-ip-validation` to `true`, which makes it drop every `externalIP` and `loadBalancerIP` that falls outside an explicitly configured range. Because an unset range means "accept nothing" rather than "accept everything", kOps leaves strict validation off unless you tell it which CIDRs to trust:

```yaml
spec:
  networking:
    kubeRouter:
      externalIPRanges:
      - 192.0.2.0/24
      loadBalancerIPRanges:
      - 198.51.100.0/24
```

Setting either field turns strict validation on, so anything outside the ranges you list stops being programmed into IPVS and stops being advertised over BGP. Leaving both empty keeps the current behavior, where every address is accepted. Both fields take a list of CIDRs, and neither may overlap `spec.networking.serviceClusterIPRange`, since kube-router rejects those at runtime anyway.

Two side effects are worth knowing about. `externalIPRanges` also feeds the network policy controller's firewall rules, which was the flag's original purpose upstream, so it does a little more than filter the service proxy. And `loadBalancerIPRanges` does **not** turn on kube-router's load balancer IPAM - that is gated separately on `--run-loadbalancer`, which kOps does not enable.
