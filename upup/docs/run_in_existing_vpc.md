## Running in a shared VPC

When launching into a shared VPC, the VPC & the Internet Gateway will be reused, but we create a new subnet per zone,
and a new route table.

Use cloudup with the `--vpc` and `--network-cidr` arguments for your existing VPC, with --dryrun so we can see the
config before we apply it.


```
export KOPS_STATE_STORE=s3://<somes3bucket>
export CLUSTER_NAME=<sharedvpc.mydomain.com>

cloudup --zones=us-east-1b --name=${CLUSTER_NAME} \
  --vpc=vpc-a80734c1 --network-cidr=10.100.0.0/16 --dryrun
```

Then `kops edit cluster  --name=${CLUSTER_NAME}` should show you something like:

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


You can then run cloudup again in dryrun mode (you don't need any arguments, because they're all in the config file):

```
cloudup --dryrun --name=${CLUSTER_NAME}
```

Review the changes to make sure they are OK -  the Kubernetes settings might not be ones you want on a shared VPC (in which case,
open an issue!)

Note also the Kubernetes VPCs (currently) require `EnableDNSHostnames=true`.  Cloudup will detect the required change,
 but refuse to make it automatically because it is a shared VPC.  Please review the implications and make the change
 to the VPC manually.

Once you're happy, you can create the cluster using:

```
cloudup --name=${CLUSTER_NAME}
```


Finally, if your shared VPC has a KubernetesCluster tag (because it was created with cloudup), you should
probably remove that tag to indicate to indicate that the resources are not owned by that cluster, and so
deleting the cluster won't try to delete the VPC.  (Deleting the VPC won't succeed anyway, because it's in use,
but it's better to avoid the later confusion!)
