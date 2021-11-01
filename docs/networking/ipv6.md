# IPv6

{{ kops_feature_table(kops_added_ff='1.23') }}

kOps has experimental support for configuring clusters with IPv6-only pods and dual-stack nodes.

IPv6 mode is specified by setting `nonMasqueradeCIDR: "::/0"` in the cluster spec.
The `--ipv6` flag of `kops create cluster` sets this field, among others.

## Cloud providers

kOps currently supports IPv6 on AWS only.

IPv6 requires the external Cloud Controller Manager.

## VPC and subnets

The VPC can be either shared or managed by kOps. If shared, it must have an IPv6 pool associated.

Subnet IPv6 CIDR allocations may be specified in the cluster spec using the special syntax `/LEN#N`,
where "LEN" is the prefix length and "N" is the hexadecimal sequence number of the CIDR within the VPC's IPv6 CIDR.
For example, if the VPC's CIDR is `2001:db8::/56` then the syntax `/64#a` would mean `2001:db8:0:a/64`.

## CNI

kOps currently supports IPv6 on Calico, Cilium, and bring-your-own CNI only.

CNIs must not masquerade IPv6 addresses.

### Calico

Running IPv6 with Calico requires a Debian 11-based AMI. As of the writing of this document, Ubuntu does not work due to an 
[issue with systemd's handling of AWS's incorrect DHCP responses](https://github.com/systemd/systemd/issues/20803).

## Future work

* kOps currently does not have a solution for NAT64/DNS64.
