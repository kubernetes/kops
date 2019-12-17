## Running in a shared VPC

When launching into a shared VPC, the VPC & the Internet Gateway will be reused. If you are not using an internet gateway
 or NAT gateway you can tell _kops_ to ignore egress. By default we create a new subnet per zone, and a new route table, 
 but you can also use a shared subnet (see [below](#shared-subnets)).

1. Use `kops create cluster` with the `--vpc` argument for your existing VPC:

  ```shell
  export KOPS_STATE_STORE=s3://<somes3bucket>
  export CLUSTER_NAME=<sharedvpc.mydomain.com>
  export VPC_ID=vpc-12345678 # replace with your VPC id
  export NETWORK_CIDR=10.100.0.0/16 # replace with the cidr for the VPC ${VPC_ID}

  kops create cluster --zones=us-east-1b --name=${CLUSTER_NAME} --vpc=${VPC_ID}
  ```

2. Then `kops edit cluster ${CLUSTER_NAME}` will show you something like:

  ```yaml
  metadata:
    creationTimestamp: "2016-06-27T14:23:34Z"
    name: ${CLUSTER_NAME}
  spec:
    cloudProvider: aws
    networkCIDR: ${NETWORK_CIDR}
    networkID: ${VPC_ID}
    nonMasqueradeCIDR: 100.64.0.0/10
    subnets:
    - cidr: 172.20.32.0/19
      name: us-east-1b
      type: Public
      zone: us-east-1b
  ```

  Verify that `networkCIDR` & `networkID` match your VPC CIDR & ID. 
  You likely need to set the CIDR on each of the Zones, because subnets in a VPC cannot overlap.

3. You can then run `kops update cluster` in preview mode (without `--yes`). 
  You don't need any arguments, because they're all in the cluster spec:

  ```shell
  kops update cluster ${CLUSTER_NAME}
  ```

  Review the changes to make sure they are OK - the Kubernetes settings might 
   not be ones you want on a shared VPC (in which case, open an issue!)

  **Note also the Kubernetes VPCs (currently) require `EnableDNSHostnames=true`. kops will detect the required change,
   but refuse to make it automatically because it is a shared VPC. Please review the implications and make the change
   to the VPC manually.**

4. Once you're happy, you can create the cluster using:

  ```shell
  kops update cluster ${CLUSTER_NAME} --yes
  ```

  This will add an additional Tag to your aws vpc resource. This tag
  will be removed automatically if you delete your kops cluster.

  ```
  "kubernetes.io/cluster/<cluster-name>" = "shared"
  ```

  **Prior to kops 1.8 this Tag Key was `KubernetesCluster` which is obsolete and should
  not be used anymore as it only supports one cluster.**


### VPC with multiple CIDRs

AWS now allows you to add more CIDRs to a VPC, the param `additionalNetworkCIDRs` allows you to specify any additional CIDRs added to the VPC.

```yaml
metadata:
  creationTimestamp: "2016-06-27T14:23:34Z"
  name: ${CLUSTER_NAME}
spec:
  cloudProvider: aws
  networkCIDR: 10.1.0.0/16
  additionalNetworkCIDRs:
  - 10.2.0.0/16
  networkID: vpc-00aa5577
  subnets:
  - cidr: 10.1.0.0/19
    name: us-east-1b
    type: Public
    zone: us-east-1b
    id: subnet-1234567
  - cidr: 10.2.0.0/19
    name: us-east-1b
    type: Public
    zone: us-east-1b
    id: subnet-1234568
```


## Advanced Options for Creating Clusters in Existing VPCs

### Shared Subnets

`kops` can create a cluster in shared subnets in both public and private network [topologies](topology.md). Doing so is not recommended unless you are using [external networking](networking.md#supported-cni-networking)

1. Use kops create cluster with the `--subnets` argument for your existing subnets:

  ```shell
  export KOPS_STATE_STORE=s3://<somes3bucket>
  export CLUSTER_NAME=<sharedvpc.mydomain.com>
  export VPC_ID=vpc-12345678 # replace with your VPC id
  export NETWORK_CIDR=10.100.0.0/16 # replace with the cidr for the VPC ${VPC_ID}
  export SUBNET_ID=subnet-12345678 # replace with your subnet id
  export SUBNET_CIDR=10.100.0.0/24 # replace with your subnet CIDR
  export SUBNET_IDS=$SUBNET_IDS # replace with your comma separated subnet ids

  kops create cluster --zones=us-east-1b --name=${CLUSTER_NAME} --subnets=${SUBNET_IDS}
  ```

  `--vpc` is optional when specifying `--subnets`. When creating a cluster with a 
  private topology and shared subnets, the utility subnets should be specified similarly with `--utility-subnets`.

2. Then `kops edit cluster ${CLUSTER_NAME}` will show you something like:

  ```
  metadata:
    creationTimestamp: "2016-06-27T14:23:34Z"
    name: ${CLUSTER_NAME}
  spec:
    cloudProvider: aws
    networkCIDR: ${NETWORK_CIDR}
    networkID: ${VPC_ID}
    nonMasqueradeCIDR: 100.64.0.0/10
    subnets:
    - cidr: ${SUBNET_CIDR}
      id: ${SUBNET_ID}
      name: us-east-1b
      type: Public
      zone: us-east-1b
  ```

3. Once you're happy, you can create the cluster using:

  ```
  kops update cluster ${CLUSTER_NAME} --yes
  ```

### Subnet Tags

  By default, kops will tag your existing subnets with the standard tags:

  Public/Utility Subnets:
  ```
  "kubernetes.io/cluster/<cluster-name>" = "shared"
  "kubernetes.io/role/elb"               = "1"
  "SubnetType"                           = "Utility"
  ```

  Private Subnets:
  ```
  "kubernetes.io/cluster/<cluster-name>" = "shared"
  "kubernetes.io/role/internal-elb"      = "1"
  "SubnetType"                           = "Private"
  ```
  
  These tags are important, for example, your services will be unable to create public or private Elastic Load Balancers (ELBs) if the respective `elb` or `internal-elb` tags are missing.
  
  If you would like to manage these tags externally then specify `--disable-subnet-tags` during your cluster creation. This will prevent kops from tagging existing subnets and allow some custom control, such as separate subnets for internal ELBs.
  
  Prior to kops 1.8 `KubernetesCluster` tag was used instead of `kubernetes.io/cluster/<cluster-name>`. This lead to several problems if there were more than one Kubernetes Cluster in a subnet. After you upgraded to kops 1.8 ensure the `KubernetesCluster` Tag is removed from subnets otherwise `kubernetes.io/cluster/<clustername>` won't have any effect!

### Shared NAT Egress

On AWS in private [topology](topology.md), `kops` creates one NAT Gateway (NGW) per AZ. If your shared VPC is already set up with an NGW in the subnet that `kops` deploys private resources to, it is possible to specify the ID and have `kops`/`kubernetes` use it.

If you don't want to use NAT Gateways but have setup [EC2 NAT Instances](https://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_NAT_Instance.html) in your VPC that you can share, it's possible to specify the IDs of said instances and have `kops`/`kubernetes` use them.

After creating a basic cluster spec, edit your cluster to specify NGW:

`kops edit cluster ${CLUSTER_NAME}`

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    egress: nat-987654321
    type: Private
    zone: us-east-1a
  - cidr: 10.20.96.0/21
    name: us-east-1b
    egress: i-987654321
    type: Private
    zone: us-east-1a
  - cidr: 10.20.32.0/21
    name: utility-us-east-1a
    type: Utility
    zone: us-east-1a
```

Please note:

* You must specify pre-create subnets for all the subnets, or for none of them.
* kops won't alter your existing subnets.  Therefore they must be correctly set up with route tables etc.  The
  Public or Utility subnets should have public IPs and an internet gateway configured as their default route
  in their route table.  Private subnets should not have public IPs, and will typically have a NAT gateway
  configured as their default route.
* kops won't create a route-table at all if we're not creating subnets.
* In the example above the first subnet is using a shared NAT Gateway while the
  second one is using a shared NAT Instance

### Externally Managed Egress

If you are using an unsupported egress configuration in your VPC, _kops_ can be told to ignore egress by using a configuration like:

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    egress: External
    type: Private
    zone: us-east-1a
  - cidr: 10.20.96.0/21
    name: us-east-1b
    egress: External
    type: Private
    zone: us-east-1a
  - cidr: 10.20.32.0/21
    name: utility-us-east-1a
    type: Utility
    zone: us-east-1a
    egress: External
```

This tells _kops_ that egress is being managed externally. This is preferable when using virtual private gateways 
(currently unsupported) or using other configurations to handle egress routing. 

### Proxy VPC Egress

See [HTTP Forward Proxy Support](http_proxy.md)
