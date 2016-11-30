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

# Troubleshooting

### 1) Unable to forward agent to master via bastion

Verify your local agent is configured correctly

```
$ ssh-add -L
ssh-rsa <PUBLIC_RSA_HASH> /Users/kris/.ssh/id_rsa
```

If that command returns no results, add the agent to `ssh-agent`

```
ssh-add ~/.ssh/id_rsa
```

Check the agent is now added using `ssh-add -L`

SSH into the bastion, then into a master

```
ssh -A admin@<bastion_elb_a_record>
ssh admin@<master_ip>
```