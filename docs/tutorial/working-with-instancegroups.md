# Working with InstanceGroups

The kops InstanceGroup is a declarative model of a group of nodes.  By modifying the object, you
can change the instance type you're using, the number of nodes you have, the OS image you're running - essentially
all the per-node configuration is in the InstanceGroup.

We'll assume you have a working cluster - if not, you probably want to read [how to get started on GCE](../getting_started/gce.md).

## Changing the number of nodes

If you `kops get ig` you should see that you have InstanceGroups for your nodes and for your master:

```
> kops get ig
NAME			ROLE	MACHINETYPE	MIN	MAX	SUBNETS
master-us-central1-a	Master	n1-standard-1	1	1	us-central1
nodes			Node	n1-standard-2	2	2	us-central1
```

Let's change the number of nodes to 3.  We'll edit the InstanceGroup configuration using `kops edit` (which
should be very familiar to you if you've used `kubectl edit`).  `kops edit ig nodes` will open
the InstanceGroup in your editor, looking a bit like this:

```
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-10-03T15:17:31Z
  labels:
    kops.k8s.io/cluster: simple.k8s.local
  name: nodes
spec:
  image: cos-cloud/cos-stable-57-9202-64-0
  machineType: n1-standard-2
  maxSize: 2
  minSize: 2
  role: Node
  subnets:
  - us-central1
  zones:
  - us-central1-a
```

<!-- TODO enable cluster autoscaler or GCE autoscaler -->

Edit `minSize` and `maxSize`, changing both from 2 to 3, save and exit your editor.  If you wanted to change
the image or the machineType, you could do that here as well.  There are actually a lot more fields,
but most of them have their default values, so won't show up unless they are set.  The general approach is the same though.

<!-- TODO link to API reference docs -->

On saving you'll note that nothing happens.  Although you've changed the model, you need to tell kops to
apply your changes to the cloud.

<!-- TODO can we have a dirty flag somehow -->

We use the same `kops update cluster` command that we used when initially creating the cluster; when
run without `--yes` it should show you a preview of the changes, and now there should be only one change:

```
> kops update cluster
Will modify resources:
  InstanceGroupManager/us-central1-a-nodes-simple-k8s-local
  	TargetSize          	 2 -> 3
```

This is saying that we will alter the `TargetSize` property of the `InstanceGroupManager` object named
`us-central1-a-nodes-simple-k8s-local`, changing it from 2 to 3.

That's what we want, so we `kops update cluster --yes`.

<!-- TODO: Make Changes may require instances to restart: kops rolling-update cluster appear selectively -->

kops will resize the GCE managed instance group from 2 to 3, which will create a new GCE instance,
which will then boot and join the cluster.  Within a minute or so you should see the new node join:

```
> kubectl get nodes
NAME                        STATUS    AGE       VERSION
master-us-central1-a-thjq   Ready     10h       v1.7.2
nodes-g2v2                  Ready     10h       v1.7.2
nodes-tmk8                  Ready     10h       v1.7.2
nodes-z2cz                  Ready     1s       v1.7.2
```

`nodes-z2cz` just joined our cluster!


## Changing the image

That was a fairly simple change, because we didn't have to reboot the nodes.  Most changes though do
require rolling your instances - this is actually a deliberate design decision, in that we are aiming
for immutable nodes.  An example is changing your image.  We're using `cos-stable`, which is Google's
Container OS.  Let's try Debian Stretch instead.

If you run `gcloud compute images list` to list the images available to you in GCE, you should see
a debian-9 image:

```
> gcloud compute images list
...
debian-9-stretch-v20170918                        debian-cloud             debian-9                              READY
...
```

<!-- TODO: Auto select debian-cloud/debian-9 => debian-cloud/debian-9-stretch-v20170918 -->

So now we'll do the same `kops edit ig nodes`, except this time change the image to `debian-cloud/debian-9-stretch-v20170918`:

Now `kops update cluster` will show that you're going to create a new [GCE Instance Template](https://cloud.google.com/compute/docs/reference/latest/instanceTemplates),
and that the Managed Instance Group is going to use it:

```
Will create resources:
  InstanceTemplate/nodes-simple-k8s-local
  	Network             	name:default id:default
  	Tags                	[simple-k8s-local-k8s-io-role-node]
  	Preemptible         	false
  	BootDiskImage       	debian-cloud/debian-9-stretch-v20170918
  	BootDiskSizeGB      	128
  	BootDiskType        	pd-standard
  	CanIPForward        	true
  	Scopes              	[compute-rw, monitoring, logging-write, storage-ro]
  	Metadata            	{cluster-name: <resource>, startup-script: <resource>}
  	MachineType         	n1-standard-2

Will modify resources:
  InstanceGroupManager/us-central1-a-nodes-simple-k8s-local
  	InstanceTemplate    	 id:nodes-simple-k8s-local-1507043948 -> name:nodes-simple-k8s-local
```

Note that the `BootDiskImage` is indeed set to the debian 9 image you requested.

`kops update cluster --yes` will now apply the change, but if you were to run `kubectl get nodes` you would see
that the instances had not yet been reconfigured.  There's a hint at the bottom:

```
Changes may require instances to restart: kops rolling-update cluster`
```

These changes require your instances to restart (we'll remove the COS images and replace them with Debian images).  kops
can perform a rolling update to minimize disruption, but even so you might not want to perform the update right away;
you might want to make more changes or you might want to wait for off-peak hours.  You might just want to wait for
the instances to terminate naturally - new instances will come up with the new configuration - though if you're not
using preemptible/spot instances you might be waiting for a long time.

## Performing a rolling-update of your cluster

When you're ready to force your instances to restart, use `kops rollling-update cluster`:

```
> kops rolling-update cluster
Using cluster from kubectl context: simple.k8s.local

NAME			STATUS		NEEDUPDATE	READY	MIN	MAX	NODES
master-us-central1-a	Ready		0		1	1	1	1
nodes			NeedsUpdate	3		0	3	3	3

Must specify --yes to rolling-update.
```

You can see that your nodes need to be restarted, and your masters do not.  A `kops rolling-update cluster --yes` will perform the update.
It will only restart instances that need restarting (unless you `--force` a rolling-update).

When you're ready, do `kops rolling-update cluster --yes`.  It'll take a few minutes per node, because for each node
we cordon the node, drain the pods, shut it down and wait for the new node to join the cluster and for the cluster
to be healthy again.  But this procedure minimizes disruption to your cluster - a rolling-update cluster is never
going to be something you do during your superbowl commercial, but ideally it should be minimally disruptive.

<!-- TODO: Clean up rolling-update cluster stdout -->
<!-- TODO: Pause after showing preview, to give a change to Ctrl-C -->


After the rolling-update is complete, you can see that the nodes are now running a new image:

```
> kubectl get nodes -owide
NAME                        STATUS    AGE       VERSION   EXTERNAL-IP     OS-IMAGE                             KERNEL-VERSION
master-us-central1-a-8fcc   Ready     48m       v1.7.2    35.188.177.16   Container-Optimized OS from Google   4.4.35+
nodes-9cml                  Ready     17m       v1.7.2    35.194.25.144   Container-Optimized OS from Google   4.4.35+
nodes-km98                  Ready     11m       v1.7.2    35.202.95.161   Container-Optimized OS from Google   4.4.35+
nodes-wbb2                  Ready     5m        v1.7.2    35.194.56.129   Container-Optimized OS from Google   4.4.35+
```


Next steps: learn how to perform cluster-wide operations, like [upgrading kubernetes](upgrading-kubernetes.md).
