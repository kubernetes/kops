# Network Topologies in Kops

Kops supports a number of pre defined network topologies. They are separated into commonly used scenarios, or topologies.

Each of the supported topologies are listed below, with an example on how to deploy them.

# AWS

Kops supports the following topologies on AWS

|      Topology     |   Value    | Description                                                                                                 |
| ----------------- |----------- | ----------------------------------------------------------------------------------------------------------- |
|   Public Cluster  |   public   | All masters/nodes will be launched in a **public subnet** in the VPC                                        |
|   Private Cluster |   private  | All masters/nodes will be launched in a **private subnet** in the VPC                                       |
|     Private Masters Public Nodes    |   privatemasters  | All masters will be launched into a **private subnet**, All nodes will be launched into a **public subnet** |

[More information](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Subnets.html) on Public and Private subnets in AWS

Taken from the AWS documentation :

##### Public Subnet
If a subnet's traffic is routed to an Internet gateway, the subnet is known as a public subnet.

##### Private Subnet
If a subnet doesn't have a route to the Internet gateway, the subnet is known as a private subnet.-



#### Defining a topology on create

To specify a topology use the `--topology` or `-t` flag as in :

```
kops create cluster ... --topology public|private|privatemasters
```

### Models

Each directory holds as the definition for associated resources for that topology.

These are controlled by kops `tags` - the tree walker will only parse a model if it's tag is found, so having these tags described is critical in making kops support topologies.

#### Tags

- _topology_private
- _topology_privatemasters
- _topology_public


#### Masters

Right now masters are tagged outside of topologies. The master configurations have already been cut into with topologies so we probably should look at getting masters into their respective topologies.

TODO Kris - Cut an issue for porting masters into into topologies. Also according to Justin some of the code here might be very old - so we can probably deprecate a large portion of it.

##### Nodes

Nodes have already been ported over into topology folders. Each topology can describe nodes however it needs.


