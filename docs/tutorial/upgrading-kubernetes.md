# Upgrading kubernetes

## **NOTE for Kubernetes >1.31**

Kops' upgrade procedure has historically risked violating the [Kubelet version skew policy](https://kubernetes.io/releases/version-skew-policy/#kubelet). After `kops update cluster --yes` completes and before every kube-apiserver is replaced with `kops rolling-update cluster --yes`, newly launched nodes running newer kubelet versions could be connecting to older `kube-apiserver` nodes.

**Violating this policy when upgrading to Kubernetes 1.31 can cause newer kubelets to crash.** [This kubernetes issue](https://github.com/kubernetes/kubernetes/issues/127316) provides details though it was not addressed because the change does not actually violate the version skew policy, it merely breaks tooling that was already violating the policy.

To upgrade a cluster to Kubernetes 1.31 or newer, use the new `kops reconcile cluster` command introduced in Kops 1.31. This replaces both `kops update cluster --yes` and `kops rolling-update cluster --yes`.

`kops reconcile cluster` will interleave the cloud provider updates of `kops update cluster --yes` with the node rotations of `kops rolling-update cluster --yes`.

It is comparable to the following sequence:
1. `kops update cluster --instance-group-roles=control-plane,apiserver --yes`
2. `kops rolling-update cluster --instance-group-roles=control-plane,apiserver --yes`
3. `kops update cluster --yes`
4. `kops rolling-update cluster --yes`
5. `kops update cluster --prune --yes`

**Terraform** users will need to use a targeted terraform apply with the normal `kops rolling-update cluster --yes`:

```sh
$ kops update cluster --target terraform ...

# Get the terraform resource IDs of the instance groups with a spec.role of `ControlPlane`, `Master`, or `APIServer`
# The exact output may vary.
$ terraform state list | grep -E 'aws_autoscaling_group|google_compute_instance_group_manager|hcloud_server|digitalocean_droplet|scaleway_instance_server'
aws_autoscaling_group.controlplane-us-east-1a-example-com
aws_autoscaling_group.controlplane-us-east-1b-example-com
aws_autoscaling_group.controlplane-us-east-1c-example-com
aws_autoscaling_group.nodes-example-com
aws_autoscaling_group.bastion-example-com

# Apply the changes to all control plane instance groups
$ terraform apply -target 'aws_autoscaling_group.controlplane-us-east-1a-example-com' -target 'aws_autoscaling_group.controlplane-us-east-1b-example-com' -target 'aws_autoscaling_group.controlplane-us-east-1c-example-com'

# Roll the apiserver nodes
$ kops rolling-update cluster --yes --instance-group-roles control-plane,apiserver

# Apply everything else
$ terraform apply

# Roll the remaining nodes
$ kops rolling-update cluster --yes
```

## Upgrades before Kops 1.31

Upgrading kubernetes is very easy with kOps, as long as you are using a compatible version of kOps.
The kOps `1.30.x` series (for example) supports the kubernetes 1.28, 1.29, and 1.30 series,
as per the kubernetes deprecation policy. Older versions of kubernetes will likely still work, but these
are on a best-effort basis and will have little if any testing. kOps `1.30` will not support the kubernetes
`1.31` series, and for full support of kubernetes `1.31` it is best to wait for the kOps `1.31` series release.
We aim to release the next major version of kOps within a few weeks of the equivalent major release of kubernetes.
We try to ensure that a pre-release (alpha or beta) is available at the kubernetes release date, for early adopters.

Upgrading kubernetes is similar to changing the image on an InstanceGroup, the kubernetes version is
controlled at the cluster level.  So instead of `kops edit ig <name>`, we `kops edit cluster`, and change the
`kubernetesVersion` field.  `kops edit cluster` will open your editor with the cluster, similar to:

```yaml
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  name: simple.k8s.local
spec:
  api:
    loadBalancer:
      type: Public
  authorization:
    alwaysAllow: {}
  channel: stable
  cloudProvider: gce
  configBase: gs://kubernetes-clusters/simple.k8s.local
  etcdClusters:
  - etcdMembers:
    - instanceGroup: master-us-central1-a
      name: a
    name: main
  - etcdMembers:
    - instanceGroup: master-us-central1-a
      name: a
    name: events
  iam:
    legacy: false
  kubernetesApiAccess:
  - 0.0.0.0/0
  kubernetesVersion: 1.17.2
  masterPublicName: api.simple.k8s.local
  networking:
    kubenet: {}
  nonMasqueradeCIDR: 100.64.0.0/10
  project: gce-project
  sshAccess:
  - 0.0.0.0/0
  subnets:
  - name: us-central1
    region: us-central1
    type: Public
  topology:
    dns:
      type: Public
```

Edit `kubernetesVersion`, changing it to `1.17.7` for example.


Apply the changes to the cloud infrastructure using `kops update cluster` and `kops update cluster --yes`:

```
Will create resources:
  InstanceTemplate/master-us-central1-a-simple-k8s-local
  	Network             	name:default id:default
  	Tags                	[simple-k8s-local-k8s-io-role-master]
  	Preemptible         	false
  	BootDiskImage       	cos-cloud/cos-stable-57-9202-64-0
  	BootDiskSizeGB      	64
  	BootDiskType        	pd-standard
  	CanIPForward        	true
  	Scopes              	[compute-rw, monitoring, logging-write, storage-ro, https://www.googleapis.com/auth/ndev.clouddns.readwrite]
  	Metadata            	{cluster-name: <resource>, startup-script: <resource>}
  	MachineType         	n1-standard-1

  InstanceTemplate/nodes-simple-k8s-local
  	Network             	name:default id:default
  	Tags                	[simple-k8s-local-k8s-io-role-node]
  	Preemptible         	false
  	BootDiskImage       	debian-cloud/debian-9-stretch-v20170918
  	BootDiskSizeGB      	128
  	BootDiskType        	pd-standard
  	CanIPForward        	true
  	Scopes              	[compute-rw, monitoring, logging-write, storage-ro]
  	Metadata            	{startup-script: <resource>, cluster-name: <resource>}
  	MachineType         	n1-standard-2

Will modify resources:
  InstanceGroupManager/us-central1-a-master-us-central1-a-simple-k8s-local
  	InstanceTemplate    	 id:master-us-central1-a-simple-k8s-local-1507089163 -> name:master-us-central1-a-simple-k8s-local

  InstanceGroupManager/us-central1-a-nodes-simple-k8s-local
  	InstanceTemplate    	 id:nodes-simple-k8s-local-1507089694 -> name:nodes-simple-k8s-local
```


`kops rolling-update cluster` will show that all nodes need to be restarted.

```
NAME			STATUS		NEEDUPDATE	READY	MIN	MAX	NODES
master-us-central1-a	NeedsUpdate	1		0	1	1	1
nodes			NeedsUpdate	3		0	3	3	3
```

Restart the instances with `kops rolling-update cluster --yes`.

```
> kubectl get nodes -owide
NAME                        STATUS    AGE       VERSION   EXTERNAL-IP     OS-IMAGE                             KERNEL-VERSION
master-us-central1-a-8fcc   Ready     26m       v1.17.7   35.194.56.129   Container-Optimized OS from Google   4.4.35+
nodes-9cml                  Ready     16m       v1.17.7   35.193.12.73    Ubuntu 16.04.3 LTS                   4.10.0-35-generic
nodes-km98                  Ready     10m       v1.17.7   35.194.25.144   Ubuntu 16.04.3 LTS                   4.10.0-35-generic
nodes-wbb2                  Ready     2m        v1.17.7   35.188.177.16   Ubuntu 16.04.3 LTS                   4.10.0-35-generic
```

<!-- TODO: Do we drain, validate and then restart -->
<!-- TODO: Fix timings in rolling update -->
