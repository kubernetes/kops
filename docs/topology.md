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

## Changing Topology of the API server
To change the ELB that fronts the API server from Internet facing to Internal only there are a few steps to accomplish

The AWS ELB does not support changing from internet facing to Internal.  However what we can do is have kops recreate the ELB for us.

### Steps to change the ELB from Internet-Facing to Internal
- Edit the cluster: `kops edit cluster $NAME`
- Change the api load balancer type from: Public to Internal... should look like this when done:
```
 spec:
    api:
      loadBalancer:
        type: Internal
```
 - Quit the edit
 - Run the update command to check the config: `kops update cluster $NAME`
 - BEFORE DOING the same command with the `--yes` option go into the AWS console and DELETE the api ELB!!!!!!
 - Now run: `kops update cluster $NAME --yes`
 - Finally execute a rolling update so that the instances register with the new internal ELB,  execute: `kops rolling-update cluster --cloudonly --force` command.  We have to use the  `--cloudonly` option because we deleted the api ELB so there is no way to talk to the cluster through the k8s api.  The force option is there because kops / terraform doesn't know that we need to update the instances with the ELB so we have to force it.
 Once the rolling update has completed you have an internal only ELB that has the master k8s nodes registered with it.

