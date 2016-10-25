# Network Topologies in Kops

Kops supports a number of pre defined network topologies. They are separated into commonly used scenarios, or topologies.

Each of the supported topologies are listed below, with an example on how to deploy them.

## AWS

Kops supports the following topologies on AWS

|      Topology     |   Value    | Description                                                                                                 |
| ----------------- |----------- | ----------------------------------------------------------------------------------------------------------- |
|   Public Cluster  |   public   | All masters/nodes will be launched in a **public subnet** in the VPC                                        |
|   Private Cluster |   private  | All masters/nodes will be launched in a **private subnet** in the VPC                                       |
|     Private Masters Public Nodes    |   privatemasters  | All masters will be launched into a **private subnet**, All nodes will be launched into a **public subnet** |


#### Defining a topology on create

To specify a topology use the `--topology` or `-t` flag as in :

```
kops create cluster ... --topology public|private|privatemasters
```


#### Defining a topology in the cluster configuration

The topology definition in the kops configuration is as follows

```
topology:
    type: public|private|privatemasters
```

Where kops will default to a public topology

```
topology:
    type: public
```