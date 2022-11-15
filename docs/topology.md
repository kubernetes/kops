# Network Topologies in kOps

kOps supports a number of pre-defined network topologies. They are separated into commonly used scenarios, or topologies.

Each of the supported topologies are listed below, with an example on how to deploy them.

# Supported Topologies

kOps supports the following topologies:

| Topology        | Value   | Description                                                               |
|-----------------|---------|---------------------------------------------------------------------------|
| Public Cluster  | public  | All nodes will be launched in a subnet accessible from the internet.      |
| Private Cluster | private | All nodes will be launched in a subnet with no ingress from the internet. |

# Types of Subnets

## Public Subnet

A subnet of type `Public` accepts incoming traffic from the internet.

## Private Subnet

A subnet of type `Private` does not route traffic from the internet.

If a cluster is IPv6, then `Private` subnets are IPv6-only.

If the subnet is capable of IPv4, it typically has a CIDR range from private IP address space.
Egress to the internet is typically routed through a Network Address Translation (NAT) device,
such as an AWS NAT Gateway.

If the subnet is capable of IPv6, egress to the internet is typically routed through a
connection-tracking firewall, such as an AWS Egress-only Internet Gateway. Egress to the
NAT64 range `64:ff9b::/96` is typically routed to a NAT64 device, such as an AWS NAT Gateway.

## DualStack Subnet

A subnet of type `DualStack` is like `Private`, but supports both IPv4 and IPv6.

On AWS, this subnet type is used for nodes, such as control plane nodes and bastions,
which need to be instance targets of a load balancer.

## Utility Subnet

A subnet of type `Utility` is like `Public`, but is not used to provision nodes.

Utility subnets are used to provision load balancers that accept ingress from the internet.
They are also used to provision NAT devices.

# Defining a topology on create

To specify a topology use the `--topology` or `-t` flag as in :

```
kops create cluster ... --topology public|private
```

You may also set a [networking option](networking.md), with the exception that the
`kubenet` option does not support private topology.

Newly created clusters with private topology *will* have public access to the Kubernetes API and an (optional) SSH bastion instance
through load balancers. This can be changed as described below.

## Changing the Topology of the API Server

To change the load balancer that fronts the API server from internet-facing to internal-only there are a few steps to accomplish:

AWS load balancers do not support changing from internet-facing to internal. However, we can manually delete it and have kOps recreate the ELB for us.

### Steps to Change the Load Balancer from Internet-Facing to Internal
 
- Edit the cluster: `kops edit cluster $NAME`
- Change the api load balancer type from: Public to Internal. It should look like this when done:
```yaml
 spec:
    api:
      loadBalancer:
        type: Internal
```
 - Save and exit the edit
 - Run the update command to check the config: `kops update cluster $NAME`
 - BEFORE DOING the same command with the `--yes` option, go into the AWS console and DELETE the api load balancer
 - Run: `kops update cluster $NAME --yes`
 - Run a rolling update so that the control plane nodes register with the new internal load balancer.
 Run `kops rolling-update cluster --cloudonly --force --instance-group-roles master --yes` command.  
 We have to use the  `--cloudonly` option because we deleted the API load balancer, leaving no way to talk to the Kubernetes API.
 The `--force` option is there because otherwise kOps doesn't know that we need to update the control plane nodes.
 Once the rolling update has completed you have an internal only load balancer that has the control plane nodes registered with it.

