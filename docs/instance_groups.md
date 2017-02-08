# Instance Groups

kops has the concept of "instance groups", which are a group of similar machines.  On AWS, they map to
an AutoScalingGroup.

By default, a cluster has:

* An instance group called `nodes` spanning all the zones; these instances are your workers.
* One instance group for each master zone, called `master-<zone>` (e.g. `master-us-east-1c`).  These normally have
  minimum size and maximum size = 1, so they will run a single instance.  We do this so that the cloud will
  always relaunch masters, even if everything is terminated at once.  We have an instance group per zone
  because we need to force the cloud to run an instance in every zone, so we can mount the master volumes - we
  can't do that across zones.

## Listing instance groups

`kops get instancegroups`
> ```
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
> ```
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

NOTE: rolling-update does not yet perform a real rolling update - it just shuts down machines in sequence with a delay;
 there will be downtime [Issue #37](https://github.com/kubernetes/kops/issues/37)
We have implemented a new feature that does drain and validate nodes.  This feature is experimental, and you can use the new feature by setting `export KOPS_FEATURE_FLAGS="+DrainAndValidateRollingUpdate"`.

## Resize an instance group

The procedure to resize an instance group works the same way:

* Edit the instance group, set minSize and maxSize to the desired size: `kops edit ig nodes`
* Preview changes: `kops update cluster <clustername>`
* Apply changes: `kops update cluster <clustername>  --yes`
* (you do not need a `rolling-update` when changing instancegroup sizes)


## Changing the root volume size or type

The procedure to resize the root volume works the same way:

* Edit the instance group, set `rootVolumeSize` and/or `rootVolumeType` to the desired values: `kops edit ig nodes`
* Preview changes: `kops update cluster <clustername>`
* Apply changes: `kops update cluster <clustername> --yes`
* Rolling update to update existing instances: `kops rolling-update cluster --yes`

For example, to set up a 100GB gp2 root volume, your InstanceGroup spec might look like:

```
metadata:
  creationTimestamp: "2016-07-11T04:14:00Z"
  name: nodes
spec:
  machineType: t2.medium
  maxSize: 2
  minSize: 2
  role: Node
  rootVolumeSize: 100
  rootVolumeType: gp2
```

## Creating a new instance group

Suppose you want to add a new group of nodes, perhaps with a different instance type.  You do this using
`kops create ig <InstanceGroupName>`.  Currently it opens an editor with a skeleton configuration, allowing
you to edit it before creation.

So the procedure is:

* `kops create ig morenodes`, edit and save
* Preview: `kops update cluster <clustername>`
* Apply: `kops update cluster <clustername> --yes`
* (no instances need to be relaunched, so no rolling-update is needed)

## Converting an instance group to use spot instances

Follow the normal procedure for reconfiguring an InstanceGroup, but set the maxPrice property to your bid.
For example, "0.10" represents a spot-price bid of $0.10 (10 cents) per hour.

Warning: the t2 family is not currently supported with spot pricing.  You'll need to choose a different
instance type.

An example spec looks like this:

```
metadata:
  creationTimestamp: "2016-07-10T15:47:14Z"
  name: nodes
spec:
  machineType: m3.medium
  maxPrice: "0.1"
  maxSize: 3
  minSize: 3
  role: Node
```

($0.10 per hour is a huge over-bid for an m3.medium - this is only an example!)

So the procedure is:

* Edit: `kops edit ig nodes`
* Preview: `kops update cluster <clustername>`
* Apply: `kops update cluster <clustername> --yes`
* Rolling-update, only if you want to apply changes immediately: `kops rolling-update cluster`


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
future of rolling-update will probably do this automatically)


## Deleting an instance group

If you decide you don't need an InstanceGroup any more, you delete it using: `kops delete ig <name>`

Example: `kops delete ig morenodes`

No rolling-update is needed (and note this is not currently graceful, so there may be interruptions to
workloads where the pods are running on those nodes).
