

-----------
# InstanceGroup v1alpha2 kops

>bdocs-tab:example 

```bdocs-tab:example_yaml

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: null
  labels:
    kops.k8s.io/cluster: test.aws.k8spro.com
  name: nodes
spec:
  image: kope.io/k8s-1.8-debian-jessie-amd64-hvm-ebs-2017-12-02
  machineType: t2.medium
  maxSize: 2
  minSize: 2
  minPrice: "0.2"
  cloudLabels:
    team: me
    project: ion
  nodeLabels:
    kops.k8s.io/instancegroup: nodes
  role: Node
  rootVolumeSize: 200
  rootVolumeOptimization: true
  subnets:
  - us-west-2a
  taints:
  - dedicated=gpu:NoSchedule
  - team=search:PreferNoSchedule


```


Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | InstanceGroup







InstanceGroup represents a group of instances (either nodes or masters) with the same configuration

<aside class="notice">
Appears In:

<ul> 
<li><a href="#instancegrouplist-v1alpha2-kops">InstanceGroupList kops/v1alpha2</a></li>
</ul> </aside>

Field        | Description
------------ | -----------
apiVersion <br /> *string*    | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
kind <br /> *string*    | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
metadata <br /> *[ObjectMeta](#objectmeta-v1-meta)*    | 
spec <br /> *[InstanceGroupSpec](#instancegroupspec-v1alpha2-kops)*    | 


### InstanceGroupSpec v1alpha2 kops

<aside class="notice">
Appears In:

<ul>
<li><a href="#instancegroup-v1alpha2-kops">InstanceGroup kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
additionalSecurityGroups <br /> *string array*    | AdditionalSecurityGroups attaches additional security groups (e.g. i-123456)
additionalUserData <br /> *[UserData](#userdata-v1alpha2-kops) array*    | AdditionalUserData is any additional user-data to be passed to the host
associatePublicIp <br /> *boolean*    | AssociatePublicIP is true if we want instances to have a public IP
cloudLabels <br /> *object*    | CloudLabels indicates the labels for instances in this group, at the AWS level
detailedInstanceMonitoring <br /> *boolean*    | DetailedInstanceMonitoring defines if detailed-monitoring is enabled (AWS only)
fileAssets <br /> *[FileAssetSpec](#fileassetspec-v1alpha2-kops) array*    | FileAssets is a collection of file assets for this instance group
hooks <br /> *[HookSpec](#hookspec-v1alpha2-kops) array*    | Hooks is a list of hooks for this instanceGroup, note: these can override the cluster wide ones if required
image <br /> *string*    | Image is the instance (ami etc) we should use
kubelet <br /> *[KubeletConfigSpec](#kubeletconfigspec-v1alpha2-kops)*    | Kubelet overrides kubelet config from the ClusterSpec
machineType <br /> *string*    | MachineType is the instance class
maxPrice <br /> *string*    | MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
maxSize <br /> *integer*    | MaxSize is the maximum size of the pool
minSize <br /> *integer*    | MinSize is the minimum size of the pool
nodeLabels <br /> *object*    | NodeLabels indicates the kubernetes labels for nodes in this group
role <br /> *string*    | Type determines the role of instances in this group: masters or nodes
rootVolumeIops <br /> *integer*    | If volume type is io1, then we need to specify the number of Iops.
rootVolumeOptimization <br /> *boolean*    | RootVolumeOptimization enables EBS optimization for an instance
rootVolumeSize <br /> *integer*    | RootVolumeSize is the size of the EBS root volume to use, in GB
rootVolumeType <br /> *string*    | RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
subnets <br /> *string array*    | Subnets is the names of the Subnets (as specified in the Cluster) where machines in this instance group should be placed
suspendProcesses <br /> *string array*    | SuspendProcesses disables the listed Scaling Policies
taints <br /> *string array*    | Taints indicates the kubernetes taints for nodes in this group
tenancy <br /> *string*    | Describes the tenancy of the instance group. Can be either default or dedicated. Currently only applies to AWS.
zones <br /> *string array*    | Zones is the names of the Zones where machines in this instance group should be placed This is needed for regional subnets (e.g. GCE), to restrict placement to particular zones

### InstanceGroupList v1alpha2 kops



Field        | Description
------------ | -----------
apiVersion <br /> *string*    | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
items <br /> *[InstanceGroup](#instancegroup-v1alpha2-kops) array*    | 
kind <br /> *string*    | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
metadata <br /> *[ListMeta](#listmeta-v1-meta)*    | 





