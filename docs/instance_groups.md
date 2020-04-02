# Instance Groups

kops has the concept of "instance groups", which are a group of similar machines.  On AWS, they map to
an AutoScalingGroup.

By default, a cluster has:

* An instance group called `nodes` spanning all the zones; these instances are your workers.
* One instance group for each master zone, called `master-<zone>` (e.g. `master-us-east-1c`).  These normally have
  minimum size and maximum size = 1, so they will run a single instance.  We do this so that the cloud will
  always relaunch masters, even if everything is terminated at once.  We have an instance group per zone
  because we need to force the cloud to run an instance in every zone, so we can mount the master volumes - we
  cannot do that across zones.

## Instance Groups Disclaimer

* When there is only one availability zone in a region (eu-central-1) and you would like to run multiple masters,
  you have to define multiple instance groups for each of those masters. (e.g. `master-eu-central-1-a` and
  `master-eu-central-1-b` and so on...)
* If instance groups are not defined correctly (particularly when there are an even number of master or multiple
  groups of masters into one availability zone in a single region), etcd servers will not start and master nodes will not check in. This is because etcd servers are configured per availability zone. DNS and Route53 would be the first places to check when these problems are happening.

## Listing instance groups

`kops get instancegroups`

```
NAME                    ROLE    MACHINETYPE     MIN     MAX     ZONES
master-us-east-1c       Master                  1       1       us-east-1c
nodes                   Node    t2.medium       2       2
```

You can also use the `kops get ig` alias.

## Change the instance type in an instance group

First you edit the instance group spec, using `kops edit ig nodes`.  Change the machine type to `t2.large`,
for example.  Now if you `kops get ig`, you will see the large instance size.  Note though that these changes
have not yet been applied (this may change soon though!).

To preview the change:

`kops update cluster <clustername>`

```
...
Will modify resources:
  *awstasks.LaunchConfiguration launchConfiguration/mycluster.mydomain.com
    InstanceType t2.medium -> t2.large
```

Presuming you're happy with the change, go ahead and apply it: `kops update cluster <clustername> --yes`

This change will apply to new instances only; if you'd like to roll it out immediately to all the instances
you have to perform a rolling update.

See a preview with: `kops rolling-update cluster`

Then restart the machines with: `kops rolling-update cluster --yes`

This will drain nodes, restart them with the new instance type, and validate them after startup.

## Resize an instance group

The procedure to resize an instance group works the same way:

* Edit the instance group, set minSize and maxSize to the desired size: `kops edit ig nodes`
* Preview changes: `kops update cluster <clustername>`
* Apply changes: `kops update cluster <clustername>  --yes`
* (you do not need a `rolling-update` when changing instancegroup sizes)

## Changing the root volume size or type

The default volume size for Masters is 64 GB, while the default volume size for a node is 128 GB.

The procedure to resize the root volume works the same way:

* Edit the instance group, set `rootVolumeSize` and/or `rootVolumeType` to the desired values: `kops edit ig nodes`
* `rootVolumeType` must be one of [supported volume types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html), e.g. `gp2` (default), `io1` (high performance) or `standard` (for testing).
* If `rootVolumeType` is set to `io1` then you can define the number of Iops by specifying `rootVolumeIops` (defaults to 100 if not defined)
* Preview changes: `kops update cluster <clustername>`
* Apply changes: `kops update cluster <clustername> --yes`
* Rolling update to update existing instances: `kops rolling-update cluster --yes`

For example, to set up a 200GB gp2 root volume, your InstanceGroup spec might look like:

```YAML
metadata:
  name: nodes
spec:
  machineType: t2.medium
  maxSize: 2
  minSize: 2
  role: Node
  rootVolumeSize: 200
  rootVolumeType: gp2
```

For example, to set up a 200GB io1 root volume with 200 provisioned Iops, your InstanceGroup spec might look like:

```YAML
metadata:
  name: nodes
spec:
  machineType: t2.medium
  maxSize: 2
  minSize: 2
  role: Node
  rootVolumeSize: 200
  rootVolumeType: io1
  rootVolumeIops: 200
```

## Adding additional storage to the instance groups

As of Kops 1.12.0 you can add additional storage _(note, presently confined to AWS)_ via the instancegroup specification.

```YAML
---
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: my-beloved-cluster
  name: compute
spec:
  cloudLabels:
    role: compute
  image: coreos.com/CoreOS-stable-1855.4.0-hvm
  machineType: m4.large
  ...
  volumes:
  - device: /dev/xvdd
    encrypted: true
    size: 20
    type: gp2
```

In AWS the above example shows how to add an additional 20gb EBS volume, which applies to each node within the instancegroup.

## Automatically formatting and mounting the additional storage

You can add additional storage via the above `volumes` collection though this only provisions the storage itself. Assuming you don't wish to handle the mechanics of formatting and mounting the device yourself _(perhaps via a hook)_ you can utilize the `volumeMounts` section of the instancegroup to handle this for you.

```YAML
---
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: my-beloved-cluster
  name: compute
spec:
  cloudLabels:
    role: compute
  image: coreos.com/CoreOS-stable-1855.4.0-hvm
  machineType: m4.large
  ...
  volumeMounts:
  - device: /dev/xvdd
    filesystem: ext4
    path: /var/lib/docker
  volumes:
  - device: /dev/xvdd
    encrypted: true
    size: 20
    type: gp2
```

The above will provision the additional storage, format and mount the device into the node. Note this feature is purposely distinct from `volumes` so that it may be reused in areas such as ephemeral storage. Using a `c5d.large` instance as an example, which comes with a 50gb SSD drive; we can use the `volumeMounts` to mount this into `/var/lib/docker` for us.

```YAML
---
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: my-beloved-cluster
  name: compute
spec:
  cloudLabels:
    role: compute
  image: coreos.com/CoreOS-stable-1855.4.0-hvm
  machineType: c5d.large
  ...
  volumeMounts:
  - device: /dev/nvme1n1
    filesystem: ext4
    path: /data
  # -- mount the instance storage --
  - device: /dev/nvme2n1
    filesystem: ext4
    path: /var/lib/docker
  volumes:
  - device: /dev/nvme1n1
    encrypted: true
    size: 20
    type: gp2
```

For AWS you can find more information on device naming conventions [here](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html)

```shell
$ df -h | grep nvme[12]
/dev/nvme1n1      20G   45M   20G   1% /data
/dev/nvme2n1      46G  633M   45G   2% /var/lib/docker
```

> Note: at present its up to the user ensure the correct device names.

## Creating a new instance group

Suppose you want to add a new group of nodes, perhaps with a different instance type.  You do this using `kops create ig <InstanceGroupName> --subnet <zone(s)>`. Currently the
`--subnet` flag is required, and it receives the zone(s) of the subnet(s) in which the instance group will be. The command opens an editor with a skeleton configuration, allowing you to edit it before creation.

So the procedure is:

* `kops create ig morenodes --subnet us-east-1a`

  or, in case you need it to be in more than one subnet, use a comma-separated list:

* `kops create ig morenodes --subnet us-east-1a,us-east-1b,us-east-1c`
* Preview: `kops update cluster <clustername>`
* Apply: `kops update cluster <clustername> --yes`
* (no instances need to be relaunched, so no rolling-update is needed)

## Creating a instance group of mixed instances types (AWS Only)

AWS permits the creation of mixed instance EC2 Autoscaling Groups using a [mixed instance policy](https://aws.amazon.com/blogs/aws/new-ec2-auto-scaling-groups-with-multiple-instance-types-purchase-options/), allowing the users to build a target capacity and make up of on-demand and spot instances while offloading the allocation strategy to AWS.

Support for mixed instance groups was added in Kops 1.12.0

```YAML
---
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: your.cluster.name
  name: compute
spec:
  cloudLabels:
    role: compute
  image: coreos.com/CoreOS-stable-1911.4.0-hvm
  machineType: m4.large
  maxSize: 50
  minSize: 10
  # You can manually set the maxPrice you're willing to pay - it will default to the onDemand price.
  maxPrice: "1.0"
  # add the mixed instance policy here
  mixedInstancesPolicy:
    instances:
    - m4.xlarge
    - m5.large
    - m5.xlarge
    - t2.medium
    onDemandAboveBase: 5
    spotInstancePools: 3
```

The mixed instance policy permits setting the following configurable below, but for more details please check against the AWS documentation.

```Go
// MixedInstancesPolicySpec defines the specification for an autoscaling backed by a ec2 fleet
type MixedInstancesPolicySpec struct {
  // Instances is a list of instance types which we are willing to run in the EC2 fleet
  Instances []string `json:"instances,omitempty"`
  // OnDemandAllocationStrategy indicates how to allocate instance types to fulfill On-Demand capacity
  OnDemandAllocationStrategy *string `json:"onDemandAllocationStrategy,omitempty"`
  // OnDemandBase is the minimum amount of the Auto Scaling group's capacity that must be
  // fulfilled by On-Demand Instances. This base portion is provisioned first as your group scales.
  OnDemandBase *int64 `json:"onDemandBase,omitempty"`
  // OnDemandAboveBase controls the percentages of On-Demand Instances and Spot Instances for your
  // additional capacity beyond OnDemandBase. The range is 0–100. The default value is 100. If you
  // leave this parameter set to 100, the percentages are 100% for On-Demand Instances and 0% for
  // Spot Instances.
  OnDemandAboveBase *int64 `json:"onDemandAboveBase,omitempty"`
  // SpotAllocationStrategy diversifies your Spot capacity across multiple instance types to
  // find the best pricing. Higher Spot availability may result from a larger number of
  // instance types to choose from https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-fleet.html#spot-fleet-allocation-strategy
  SpotAllocationStrategy *string `json:"spotAllocationStrategy,omitempty"`
  // SpotInstancePools is the number of Spot pools to use to allocate your Spot capacity (defaults to 2)
  // pools are determined from the different instance types in the Overrides array of LaunchTemplate
  SpotInstancePools *int64 `json:"spotInstancePools,omitempty"`
}
```

Note: as of writing this the kube cluster autoscaler does not support mixed instance groups, in the sense it will still scale groups up and down based on capacity but some of the simulations it does might be wrong as it's not aware of the instance type coming into the group.

Note: when upgrading from a launchconfiguration to launchtemplate with mixed instance policy the launchconfiguration is left undeleted as has to be manually removed.

## Moving from one instance group spanning multiple AZs to one instance group per AZ

It may be beneficial to have one IG per AZ rather than one IG spanning multiple AZs. One common example is, when you have a persistent volume claim bound to an AWS EBS Volume this volume is bound to the AZ it has been created in so any resource (e.g. a StatefulSet) depending on that volume is bound to that same AZ. In this case you have to ensure that there is at least one node running in that same AZ, which is not guaranteed by one IG. This however can be guaranteed by one IG per AZ.

So the procedure is:

* `kops edit ig nodes`
* Remove two of the subnets, e.g. `eu-central-1b` and `eu-central-1c`
  * Alternatively you can also delete the existing IG and create a new one with a more suitable name
* `kops create ig nodes-eu-central-1b --subnet eu-central-1b`
* `kops create ig nodes-eu-central-1c --subnet eu-central-1c`
* Preview: `kops update cluster <clustername>`
* Apply: `kops update cluster <clustername> --yes`
* Rolling update to update existing instances: `kops rolling-update cluster --yes`

## Converting an instance group to use spot instances

Follow the normal procedure for reconfiguring an InstanceGroup, but set the maxPrice property to your bid.
For example, "0.10" represents a spot-price bid of $0.10 (10 cents) per hour.

An example spec looks like this:

```YAML
metadata:
  name: nodes
spec:
  machineType: t2.medium
  maxPrice: "0.01"
  maxSize: 3
  minSize: 3
  role: Node
```

So the procedure is:

* Edit: `kops edit ig nodes`
* Preview: `kops update cluster <clustername>`
* Apply: `kops update cluster <clustername> --yes`
* Rolling-update, only if you want to apply changes immediately: `kops rolling-update cluster`

## Adding Taints or Labels to an Instance Group

If you're running Kubernetes 1.6.0 or later, you can also control taints in the InstanceGroup.
The taints property takes a list of strings. The following example would add two taints to an IG,
using the same `edit` -> `update` -> `rolling-update` process as above.

Additionally, `nodeLabels` can be added to an IG in order to take advantage of Pod Affinity. Every node in the IG will be assigned the desired labels. For more information see the [labels](./labels.md) documentation.

```YAML
metadata:
  name: nodes
spec:
  machineType: m3.medium
  maxSize: 3
  minSize: 3
  role: Node
  taints:
  - dedicated=gpu:NoSchedule
  - team=search:PreferNoSchedule
  nodeLabels:
    spot: "false"
```

## Resizing the master

(This procedure should be pretty familiar by now!)

Your master instance group will probably be called `master-us-west-1c` or something similar.

`kops edit ig master-us-west-1c`

Add or set the machineType:

```YAML
spec:
  machineType: m3.large
```

* Preview changes: `kops update cluster <clustername>`

* Apply changes: `kops update cluster <clustername> --yes`

* Rolling-update, only if you want to apply changes immediately: `kops rolling-update cluster`

If you want to minimize downtime, scale the master ASG up to size 2, then wait for that new master to
be Ready in `kubectl get nodes`, then delete the old master instance, and scale the ASG back down to size 1.  (A
future version of rolling-update will probably do this automatically)

## Deleting an instance group

If you decide you don't need an InstanceGroup any more, you delete it using: `kops delete ig <name>`

Example: `kops delete ig morenodes`

No `kops update cluster` nor `kops rolling-update` is needed, so **be careful** when deleting an instance group, your nodes will be deleted automatically (and note this is not currently graceful, so there may be interruptions to workloads where the pods are running on those nodes).

## EBS Volume Optimization

EBS-Optimized instances can be created by setting the following field:

```YAML
spec:
  rootVolumeOptimization: true
```

## Additional user-data for cloud-init

Kops utilizes cloud-init to initialize and setup a host at boot time. However in certain cases you may already be leveraging certain features of cloud-init in your infrastructure and would like to continue doing so. More information on cloud-init can be found [here](http://cloudinit.readthedocs.io/en/latest/)

Additional user-data can be passed to the host provisioning by setting the `additionalUserData` field. A list of valid user-data content-types can be found [here](http://cloudinit.readthedocs.io/en/latest/topics/format.html#mime-multi-part-archive)

Example:

```YAML
spec:
  additionalUserData:
  - name: myscript.sh
    type: text/x-shellscript
    content: |
      #!/bin/sh
      echo "Hello World.  The time is now $(date -R)!" | tee /root/output.txt
  - name: local_repo.txt
    type: text/cloud-config
    content: |
      #cloud-config
      apt:
        primary:
          - arches: [default]
            uri: http://local-mirror.mydomain
            search:
              - http://local-mirror.mydomain
              - http://archive.ubuntu.com
```

## Add Tags on AWS autoscalling groups and instances

If you need to add tags on auto scaling groups or instances (propagate ASG tags), you can add it in the instance group specs with *cloudLabels*. Cloud Labels defined at the cluster spec level will also be inherited.

```YAML
# Example for nodes
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: nodes
spec:
  cloudLabels:
    billing: infra
    environment: dev
  associatePublicIp: false
  machineType: m4.xlarge
  maxSize: 20
  minSize: 2
  role: Node
```

## Suspending Scaling Processes on AWS Autoscaling groups

Autoscaling groups automatically include multiple [scaling processes](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-suspend-resume-processes.html#process-types)
that keep our ASGs healthy.  In some cases, you may want to disable certain scaling activities.

An example of this is if you are running multiple AZs in an ASG while using a Kubernetes Autoscaler.
The autoscaler will remove specific instances that are not being used.  In some cases, the `AZRebalance` process
will rescale the ASG without warning.

```YAML
# Example for nodes
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: nodes
spec:
  machineType: m4.xlarge
  maxSize: 20
  minSize: 2
  role: Node
  suspendProcesses:
  - AZRebalance
```

## Protect new instances from scale in

Autoscaling groups may scale up or down automatically to balance types of instances, regions, etc.
[Instance protection](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-instance-termination.html#instance-protection) prevents the ASG from being scaled in.

```YAML
# Example for nodes
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: nodes
spec:
  machineType: m4.xlarge
  maxSize: 20
  minSize: 2
  role: Node
  instanceProtection: true
```

## Attaching existing Load Balancers to Instance Groups

Instance groups can be linked to up to 10 load balancers. When attached, any instance launched will
automatically register itself to the load balancer. For example, if you can create an instance group
dedicated to running an ingress controller exposed on a
[NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport), you can
manually create a load balancer and link it to the instance group. Traffic to the load balancer will now
automatically go to one of the nodes.

You can specify either `loadBalancerName` to link the instance group to an AWS Classic ELB or you can
specify `targetGroupArn` to link the instance group to a target group, which are used by Application
load balancers and Network load balancers.

```YAML
# Example ingress nodes
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: ingress
spec:
  machineType: m4.large
  maxSize: 2
  minSize: 2
  role: Node
  externalLoadBalancers:
  - targetGroupArn: arn:aws:elasticloadbalancing:eu-west-1:123456789012:targetgroup/my-ingress-target-group/0123456789abcdef
  - loadBalancerName: my-elb-classic-load-balancer
```

## Enabling Detailed-Monitoring on AWS instances

Detailed-Monitoring will cause the monitoring data to be available every 1 minute instead of every 5 minutes. [Enabling Detailed Monitoring](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch-new.html). In production environments you may want to consider to enable detailed monitoring for quicker troubleshooting.

**Note: that enabling detailed monitoring is a subject for [charge](https://aws.amazon.com/cloudwatch)**

```YAML
# Example for nodes
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: nodes
spec:
  detailedInstanceMonitoring: true
  machineType: t2.medium
  maxSize: 2
  minSize: 2
  role: Node
```

## Booting from a volume in OpenStack

If you want to boot from a volume when you are running in openstack you can set annotations on the instance groups.

```YAML
# Example for nodes
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: nodes
  annotations:
    openstack.kops.io/osVolumeBoot: enabled
    openstack.kops.io/osVolumeSize: "15" # In gigabytes
spec:
  detailedInstanceMonitoring: true
  machineType: t2.medium
  maxSize: 2
  minSize: 2
  role: Node
```

If `openstack.kops.io/osVolumeSize` is not set it will default to the minimum disk specified by the image.

## Setting Custom Kernel Runtime Parameters

To add custom kernel runtime parameters to your instance group, specify the
`sysctlParameters` field as an array of strings. Each string must take the form
of `variable=value` the way it would appear in sysctl.conf (see also
`sysctl(8)` manpage).

Unlike a simple file asset, specifying kernel runtime parameters in this manner
would correctly invoke `sysctl --system` automatically for you to apply said
parameters.

For example:

```YAML
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: nodes
spec:
  sysctlParameters:
    - fs.pipe-user-pages-soft=524288
    - net.ipv4.tcp_keepalive_time=200
```

which would end up in a drop-in file on nodes of the instance group in question.
