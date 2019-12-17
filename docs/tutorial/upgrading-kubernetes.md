# Upgrading kubernetes

Upgrading kubernetes is very easy with kops, as long as you are using a compatible version of kops.
The kops `1.8.x` series (for example) supports the kubernetes 1.6, 1.7 and 1.8 series,
as per the kubernetes deprecation policy.  Older versions of kubernetes will likely still work, but these
are on a best-effort basis and will have little if any testing.  kops `1.8` will not support the kubernetes
`1.9` series, and for full support of kubernetes `1.9` it is best to wait for the kops `1.9` series release.
We aim to release the next major version of kops within a few weeks of the equivalent major release of kubernetes,
so kops `1.9.0` will be released within a few weeks of kubernetes `1.9.0`.  We try to ensure that a 1.9 pre-release
(alpha or beta) is available at the kubernetes release, for early adopters.

Upgrading kubernetes is similar to changing the image on an InstanceGroup, except that the kubernetes version is
controlled at the cluster level.  So instead of `kops edit ig <name>`, we `kops edit cluster`, and change the
`kubernetesVersion` field.  `kops edit cluster` will open your editor with the cluster, similar to:

```
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  creationTimestamp: 2017-10-04T03:52:25Z
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
  kubernetesVersion: 1.7.2
  masterInternalName: api.internal.simple.k8s.local
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
    masters: public
    nodes: public
```

Edit `kubernetesVersion`, changing it to `1.7.7` for example.


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
master-us-central1-a-8fcc   Ready     26m       v1.7.7    35.194.56.129   Container-Optimized OS from Google   4.4.35+
nodes-9cml                  Ready     16m       v1.7.7    35.193.12.73    Ubuntu 16.04.3 LTS                   4.10.0-35-generic
nodes-km98                  Ready     10m       v1.7.7    35.194.25.144   Ubuntu 16.04.3 LTS                   4.10.0-35-generic
nodes-wbb2                  Ready     2m        v1.7.7    35.188.177.16   Ubuntu 16.04.3 LTS                   4.10.0-35-generic
```

<!-- TODO: Do we drain, validate and then restart -->
<!-- TODO: Fix timings in rolling update -->
