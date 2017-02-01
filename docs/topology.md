# Network Topologies in Kops

Kops supports a number of pre defined network topologies. They are separated into commonly used scenarios, or topologies.

Each of the supported topologies are listed below, with an example on how to deploy them.

# AWS

Kops supports the following topologies on AWS

|      Topology     |   Value    | Description                                                                                                 |
| ----------------- |----------- | ----------------------------------------------------------------------------------------------------------- |
|   Public Cluster  |   public   | All masters/nodes will be launched in a **public subnet** in the VPC                                        |
|   Private Cluster |   private  | All masters/nodes will be launched in a **private subnet** in the VPC                                       |


[More information](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Subnets.html) on Public and Private subnets in AWS

Notes on subnets

##### Public Subnet
If a subnet's traffic is routed to an Internet gateway, the subnet is known as a public subnet.

##### Private Subnet
If a subnet doesn't have a route to the Internet gateway, the subnet is known as a private subnet.

Private topologies *will* have public access via the Kubernetes API and an (optional) SSH bastion instance.

# Defining a topology on create

To specify a topology use the `--topology` or `-t` flag as in :

```
kops create cluster ... --topology public|private
```

In the case of a private cluster you must also set a networking option other
than `kubenet`.  Currently the supported options are:

- kopeio-vxlan
- weave
- calico
- cni

More information about [networking options](networking.md) can be found in our documentation.
