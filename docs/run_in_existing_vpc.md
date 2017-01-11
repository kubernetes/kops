## Running in a shared VPC

When launching into a shared VPC, the VPC & the Internet Gateway will be reused. By default we create a new subnet per zone,
and a new route table, but you can also use a shared subnet (see [below](#running-in-a-shared-subnet)).

Use kops create cluster with the `--vpc` and `--network-cidr` arguments for your existing VPC:


```
export KOPS_STATE_STORE=s3://<somes3bucket>
export CLUSTER_NAME=<sharedvpc.mydomain.com>

kops create cluster --zones=us-east-1b --name=${CLUSTER_NAME} \
  --vpc=vpc-a80734c1 --network-cidr=10.100.0.0/16
```

Then `kops edit cluster ${CLUSTER_NAME}` should show you something like:

```
metadata:
  creationTimestamp: "2016-06-27T14:23:34Z"
  name: ${CLUSTER_NAME}
spec:
  cloudProvider: aws
  networkCIDR: 10.100.0.0/16
  networkID: vpc-a80734c1
  nonMasqueradeCIDR: 100.64.0.0/10
  zones:
  - cidr: 10.100.32.0/19
    name: eu-central-1a
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

## Running in a shared subnet

You can also use a shared subnet. Doing so is not recommended unless you are using external networking ([kope-routing](https://github.com/kopeio/kope-routing)).

Edit your cluster to add the ID of the subnet:

`kops edit cluster ${CLUSTER_NAME}`

```
metadata:
  creationTimestamp: "2016-06-27T14:23:34Z"
  name: ${CLUSTER_NAME}
spec:
  cloudProvider: aws
  networkCIDR: 10.100.0.0/16
  networkID: vpc-a80734c1
  nonMasqueradeCIDR: 100.64.0.0/10
  zones:
  - cidr: 10.100.32.0/19
    name: eu-central-1a
    id: subnet-1234567 # Replace this with the ID of your subnet
```

Make sure that the CIDR matches the CIDR of your subnet. Then update your cluster through the normal update procedure:

```
kops update cluster ${CLUSTER_NAME}
# Review changes
kops update cluster ${CLUSTER_NAME} --yes
```
