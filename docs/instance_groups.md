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
* If `rootVolumeType` is set to `io1` then you can define the number of Iops by specifing `rootVolumeIops` (defaults to 100 if not defined)
* Preview changes: `kops update cluster <clustername>`
* Apply changes: `kops update cluster <clustername> --yes`
* Rolling update to update existing instances: `kops rolling-update cluster --yes`

For example, to set up a 200GB gp2 root volume, your InstanceGroup spec might look like:

```
metadata:
  creationTimestamp: "2016-07-11T04:14:00Z"
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

```
metadata:
  creationTimestamp: "2016-07-11T04:14:00Z"
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

## Creating a new instance group

Suppose you want to add a new group of nodes, perhaps with a different instance type.  You do this using `kops create ig <InstanceGroupName> --subnet <zone(s)>`. Currently the
`--subnet` flag is required, and it recieves the zone(s) of the subnet(s) in which the instance group will be. The command opens an editor with a skeleton configuration, allowing you to edit it before creation.

So the procedure is:

* `kops create ig morenodes --subnet us-east-1a`

  or, in case you need it to be in more than one subnet, use a comma-separated list:

* `kops create ig morenodes --subnet us-east-1a,us-east-1b,us-east-1c`
* Preview: `kops update cluster <clustername>`
* Apply: `kops update cluster <clustername> --yes`
* (no instances need to be relaunched, so no rolling-update is needed)


## Moving from one instance group spanning multiple AZs to one instance group per AZ

It may be beneficial to have one IG per AZ rather than one IG spanning multiple AZs. One common example is, when you have a persistent volume claim bound to an AWS EBS Volume this volume is bound to the AZ it has been created in so any resource (e.g. a StatefulSet) depending on that volume is bound to that same AZ. In this case you have to ensure that there is at least one node running in that same AZ, which is not guaruanteed by one IG. This however can be guarantueed by one IG per AZ.

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

```
metadata:
  creationTimestamp: "2016-07-10T15:47:14Z"
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

```
metadata:
  creationTimestamp: "2016-07-10T15:47:14Z"
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

```
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

```
spec:
  rootVolumeOptimization: true
```

## Additional user-data for cloud-init

Kops utilizes cloud-init to initialize and setup a host at boot time. However in certain cases you may already be leaveraging certain features of cloud-init in your infrastructure and would like to continue doing so. More information on cloud-init can be found [here](http://cloudinit.readthedocs.io/en/latest/)


Aditional user-user data can be passed to the host provisioning by setting the `AdditionalUserData` field. A list of valid user-data content-types can be found [here](http://cloudinit.readthedocs.io/en/latest/topics/format.html#mime-multi-part-archive)

Example:
```
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

```
# Example for nodes
apiVersion: kops/v1alpha2
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

```
# Example for nodes
apiVersion: kops/v1alpha2
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

## Enabling Detailed-Monitoring on AWS instances

Detailed-Monitoring will cause the monitoring data to be available every 1 minute instead of every 5 minutes. [Enabling Detailed Monitoring](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch-new.html). In production environments you may want to consider to enable detailed monitoring for quicker troubleshooting.

**Note: that enabling detailed monitoring is a subject for [charge](https://aws.amazon.com/cloudwatch)**

```
# Example for nodes
apiVersion: kops/v1alpha2
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