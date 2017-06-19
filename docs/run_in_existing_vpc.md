## Running in a shared VPC

When launching into a shared VPC, the VPC & the Internet Gateway will be reused. By default we create a new subnet per zone,
and a new route table, but you can also use a shared subnet (see [below](#shared-subnets)).

Use kops create cluster with the `--vpc` argument for your existing VPC:


```
export KOPS_STATE_STORE=s3://<somes3bucket>
export CLUSTER_NAME=<sharedvpc.mydomain.com>
export VPC_ID=vpc-12345678 # replace with your VPC id
export NETWORK_CIDR=10.100.0.0/16 # replace with the cidr for the VPC ${VPC_ID}

kops create cluster --zones=us-east-1b --name=${CLUSTER_NAME} --vpc=${VPC_ID}
```

Then `kops edit cluster ${CLUSTER_NAME}` will show you something like:

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
  - cidr: 172.20.32.0/19
    name: us-east-1b
    type: Public
    zone: us-east-1b
```


Verify that networkCIDR & networkID match your VPC CIDR & ID.  You likely need to set the CIDR on each of the Zones,
because subnets in a VPC cannot overlap.

You can then run `kops update cluster` in preview mode (without --yes).  You don't need any arguments,
because they're all in the cluster spec:

```
kops update cluster ${CLUSTER_NAME}
```

Review the changes to make sure they are OK -  the Kubernetes settings might not be ones you want on a shared VPC (in which case,
open an issue!)

There is currently a [bug](https://github.com/kubernetes/kops/issues/476) where kops will tell you that it will modify the VPC and InternetGateway names:

```
Will modify resources:
  VPC    vpc/k8s.somefoo.com
    Name baz -> k8s.somefoo.com

  InternetGateway    internetGateway/k8s.somefoo.com
    Name baz -> k8s.somefoo.com
```

This will not actually happen and you can safely ignore the message.

Note also the Kubernetes VPCs (currently) require `EnableDNSHostnames=true`.  kops will detect the required change,
 but refuse to make it automatically because it is a shared VPC.  Please review the implications and make the change
 to the VPC manually.

Once you're happy, you can create the cluster using:

```
kops update cluster ${CLUSTER_NAME} --yes
```


Finally, if your shared VPC has a KubernetesCluster tag (because it was created with kops), you should
probably remove that tag to indicate that the resources are not owned by that cluster, and so
deleting the cluster won't try to delete the VPC.  (Deleting the VPC won't succeed anyway, because it's in use,
but it's better to avoid the later confusion!)

## Advanced Options for Creating Clusters in Existing VPCs

### Shared Subnets

`kops` can create a cluster in shared subnets in both public and private network [topologies](topology.md). Doing so is not recommended unless you are using [external networking](networking.md#supported-cni-networking)

After creating a basic cluster spec, edit your cluster to add the ID of the subnet:

`kops edit cluster ${CLUSTER_NAME}`

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
  - cidr: 172.20.32.0/19 # You can delete the CIDR here; it will be queried
    name: us-east-1b
    type: Public
    zone: us-east-1b
    id: subnet-1234567 # Replace this with the ID of your subnet
```

If you specify the CIDR, it must match the CIDR for the subnet; otherwise it will be populated by querying the subnet.
It is probably easier to specify the `id` and remove the `cidr`!  Remember also that the zone must match the subnet Zone.

Then update your cluster through the normal update procedure:

```
kops update cluster ${CLUSTER_NAME}
# Review changes
kops update cluster ${CLUSTER_NAME} --yes
```

If you run in AWS private topology with shared subnets, and you would like Kubernetes to provision resources in these shared subnets, you must create tags on them with Key=value `KubernetesCluster=<clustername>`. This is important, for example, if your `utility` subnets are shared, you will not be able to launch any services that create Elastic Load Balancers (ELBs).

### Shared NAT Gateways

On AWS in private [topology](topology.md), `kops` creates one NAT Gateway (NGW) per AZ. If your shared VPC is already set up with an NGW in the subnet that `kops` deploys private resources to, it is possible to specify the ID and have `kops`/`kubernetes` use it.

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

### Proxy VPC Egress

See [HTTP Forward Proxy Support](http_proxy.md)
