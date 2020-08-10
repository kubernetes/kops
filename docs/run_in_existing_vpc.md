## Running in a shared VPC

When launching into a shared VPC, kops will reuse the VPC and Internet Gateway. If you are not using an Internet Gateway
 or NAT Gateway you can tell kops to ignore egress. By default, kops creates a new subnet per zone and a new route table, 
 but you can instead use a shared subnet (see [below](#shared-subnets)).

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

  Verify that `networkCIDR` and `networkID` match your VPC CIDR and ID. 
  You probably need to set the CIDR on each of the Zones, as subnets in a VPC cannot overlap.

3. You can then run `kops update cluster` in preview mode (without `--yes`). 
  You don't need any arguments because they're all in the cluster spec:

  ```shell
  kops update cluster ${CLUSTER_NAME}
  ```

  Review the changes to make sure they are OK—the Kubernetes settings might 
   not be ones you want on a shared VPC (in which case, open an issue!)

  **Note also the Kubernetes VPCs (currently) require `EnableDNSHostnames=true`. kops will detect the required change,
   but refuse to make it automatically because it is a shared VPC. Please review the implications and make the change
   to the VPC manually.**

4. Once you're happy, you can create the cluster using:

  ```shell
  kops update cluster ${CLUSTER_NAME} --yes
  ```

  This will add an additional tag to your AWS VPC resource. This tag
  will be removed automatically if you delete your kops cluster.

  ```
  "kubernetes.io/cluster/<cluster-name>" = "shared"
  ```


### VPC with multiple CIDRs

AWS allows you to add more CIDRs to a VPC. The parameter `additionalNetworkCIDRs` allows you to specify any additional CIDRs added to the VPC.

```yaml
metadata:
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

`kops` can create a cluster in shared subnets in both public and private network [topologies](topology.md).

1. Use `kops create cluster` with the `--subnets` argument for your existing subnets:

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

  ```yaml
  metadata:
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

  ```shell
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

### Shared NAT Egress

On AWS in private [topology](topology.md), kops creates one NAT Gateway (NGW) per AZ. If your shared VPC is already set up with an NGW in the subnet that `kops` deploys private resources to, it is possible to specify the ID and have `kops`/`kubernetes` use it.

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

* You must specify pre-created subnets for either all of the subnets or none of them.
* kops won't alter your existing subnets. They must be correctly set up with route tables, etc.  The
  Public or Utility subnets should have public IPs and an Internet Gateway configured as their default route
  in their route table.  Private subnets should not have public IPs and will typically have a NAT Gateway
  configured as their default route.
* kops won't create a route-table at all if it's not creating subnets.
* In the example above the first subnet is using a shared NAT Gateway while the
  second one is using a shared NAT Instance

### Externally Managed Egress

If you are using an unsupported egress configuration in your VPC, kops can be told to ignore egress by using a configuration such as:

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

This tells kops that egress is managed externally. This is preferable when using virtual private gateways 
(currently unsupported) or using other configurations to handle egress routing. 

### Proxy VPC Egress

See [HTTP Forward Proxy Support](http_proxy.md)
