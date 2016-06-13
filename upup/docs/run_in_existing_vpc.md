## Running in a shared VPC

CloudUp is actually driven by a configuration file, stored in your state directory (`./state/config`) by default.

To build a cluster in an existing VPC, you'll need to configure the config file with the extra information
(the CLI flags just act as shortcuts to configuring the config file manually, editing the config file is "expert mode").

When launching into a shared VPC, the VPC & the Internet Gateway will be reused, but we create a new subnet per zone,
and a new route table.

Use cloudup in `--dryrun` mode to create a base configuration file:

```
cloudup --cloud=aws --zones=us-east-1b --name=<mycluster.mydomain.com> --node-size=t2.medium --master-size=t2.medium --node-count=2 --dryrun
```

Now edit your `./state/config' file.  It will probably look like this:

```
CloudProvider: aws
ClusterName: <mycluster.mydomain.com>
MasterMachineType: t2.medium
MasterZones:
- us-east-1b
NetworkCIDR: 172.22.0.0/16
NodeCount: 2
NodeMachineType: t2.medium
NodeZones:
- cidr: 172.22.0.0/19
  name: us-east-1b
```

You need to specify your VPC id, which is called NetworkID.  You likely also need to update NetworkCIDR to match whatever value your existing VPC is using,
and you likely need to set the CIDR on each of the NodeZones, because subnets in a VPC cannot overlap.  For example:

```
CloudProvider: aws
ClusterName: cluster2.awsdata.com
MasterMachineType: t2.medium
MasterZones:
- us-east-1b
NetworkID: vpc-10f95a77
NetworkCIDR: 172.22.0.0/16
NodeCount: 2
NodeMachineType: t2.medium
NodeZones:
- cidr: 172.22.224.0/19
  name: us-east-1b
```

You can then run cloudup in dryrun mode (you don't need any arguments, because they're all in the config file):

```
cloudup --dryrun
```

You should see that your VPC changes from `Shared <nil> -> true`, and you should review them to make sure
that the changes are OK - the Kubernetes settings might not be ones you want on a shared VPC (in which case,
open an issue!)

Once you're happy, you can create the cluster using:

```
cloudup
```


Finally, if your shared VPC has a KubernetesCluster tag (because it was created with cloudup), you should
probably remove that tag to indicate to indicate that the resources are not owned by that cluster, and so
deleting the cluster won't try to delete the VPC.  (Deleting the VPC won't succeed anyway, because it's in use,
but it's better to avoid the later confusion!)