/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="role",type="string",JSONPath=".spec.role",description="Role",priority=0
// +kubebuilder:printcolumn:name="machineType",type="string",JSONPath=".spec.machineType",description="Machine Type",priority=0
// +kubebuilder:printcolumn:name="min",type="integer",JSONPath=".spec.minSize",description="Min",priority=0
// +kubebuilder:printcolumn:name="max",type="integer",JSONPath=".spec.maxSize",description="Max",priority=0
// +kubebuilder:printcolumn:name="zones",type="string",JSONPath=".spec.zones",description="Zones",priority=0
// +kubebuilder:resource:shortName=ig
// InstanceGroup represents a group of instances (either nodes or masters) with the same configuration
type InstanceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceGroupSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceGroupList is a list of instance groups
type InstanceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole string describes the roles of the nodes in this InstanceGroup (master or nodes)
type InstanceGroupRole string

const (
	// InstanceGroupRoleMaster is a master role
	InstanceGroupRoleMaster InstanceGroupRole = "Master"
	// InstanceGroupRoleNode is a node role
	InstanceGroupRoleNode InstanceGroupRole = "Node"
	// InstanceGroupRoleBastion is a bastion role
	InstanceGroupRoleBastion InstanceGroupRole = "Bastion"
)

// AllInstanceGroupRoles is a list of all available roles
var AllInstanceGroupRoles = []InstanceGroupRole{
	InstanceGroupRoleBastion,
	InstanceGroupRoleMaster,
	InstanceGroupRoleNode,
}

const (
	// BtfsFilesystem indicates a btfs filesystem
	BtfsFilesystem = "btfs"
	// Ext4Filesystem indicates a ext3 filesystem
	Ext4Filesystem = "ext4"
	// XFSFilesystem indicates a xfs filesystem
	XFSFilesystem = "xfs"
)

var (
	// SupportedFilesystems is a list of supported filesystems to format as
	SupportedFilesystems = []string{BtfsFilesystem, Ext4Filesystem, XFSFilesystem}
)

// InstanceGroupSpec is the specification for an instanceGroup
type InstanceGroupSpec struct {
	// Type determines the role of instances in this group: masters or nodes
	Role InstanceGroupRole `json:"role,omitempty"`
	// Image is the instance (ami etc) we should use
	Image string `json:"image,omitempty"`
	// MinSize is the minimum size of the pool
	MinSize *int32 `json:"minSize,omitempty"`
	// MaxSize is the maximum size of the pool
	MaxSize *int32 `json:"maxSize,omitempty"`
	// MachineType is the instance class
	MachineType string `json:"machineType,omitempty"`
	// RootVolumeSize is the size of the EBS root volume to use, in GB
	RootVolumeSize *int32 `json:"rootVolumeSize,omitempty"`
	// RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
	RootVolumeType *string `json:"rootVolumeType,omitempty"`
	// If volume type is io1, then we need to specify the number of Iops.
	RootVolumeIops *int32 `json:"rootVolumeIops,omitempty"`
	// RootVolumeOptimization enables EBS optimization for an instance
	RootVolumeOptimization *bool `json:"rootVolumeOptimization,omitempty"`
	// RootVolumeDeleteOnTermination configures root volume retention policy upon instance termination.
	// The root volume is deleted by default. Cluster deletion does not remove retained root volumes.
	// NOTE: This setting applies only to the Launch Configuration and does not affect Launch Templates.
	RootVolumeDeleteOnTermination *bool `json:"rootVolumeDeleteOnTermination,omitempty"`
	// Volumes is a collection of additional volumes to create for instances within this InstanceGroup
	Volumes []*VolumeSpec `json:"volumes,omitempty"`
	// VolumeMounts a collection of volume mounts
	VolumeMounts []*VolumeMountSpec `json:"volumeMounts,omitempty"`
	// Subnets is the names of the Subnets (as specified in the Cluster) where machines in this instance group should be placed
	Subnets []string `json:"subnets,omitempty"`
	// Zones is the names of the Zones where machines in this instance group should be placed
	// This is needed for regional subnets (e.g. GCE), to restrict placement to particular zones
	Zones []string `json:"zones,omitempty"`
	// Hooks is a list of hooks for this instanceGroup, note: these can override the cluster wide ones if required
	Hooks []HookSpec `json:"hooks,omitempty"`
	// MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
	MaxPrice *string `json:"maxPrice,omitempty"`
	// SpotDurationInMinutes indicates this is a spot-block group, with the specified value as the spot reservation time
	SpotDurationInMinutes *int64 `json:"spotDurationInMinutes,omitempty"`
	// AssociatePublicIP is true if we want instances to have a public IP
	AssociatePublicIP *bool `json:"associatePublicIp,omitempty"`
	// AdditionalSecurityGroups attaches additional security groups (e.g. i-123456)
	AdditionalSecurityGroups []string `json:"additionalSecurityGroups,omitempty"`
	// CloudLabels indicates the labels for instances in this group, at the AWS level
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`
	// NodeLabels indicates the kubernetes labels for nodes in this group
	NodeLabels map[string]string `json:"nodeLabels,omitempty"`
	// FileAssets is a collection of file assets for this instance group
	FileAssets []FileAssetSpec `json:"fileAssets,omitempty"`
	// Describes the tenancy of the instance group. Can be either default or dedicated.
	// Currently only applies to AWS.
	Tenancy string `json:"tenancy,omitempty"`
	// Kubelet overrides kubelet config from the ClusterSpec
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`
	// Taints indicates the kubernetes taints for nodes in this group
	Taints []string `json:"taints,omitempty"`
	// MixedInstancesPolicy defined a optional backing of an AWS ASG by a EC2 Fleet (AWS Only)
	MixedInstancesPolicy *MixedInstancesPolicySpec `json:"mixedInstancesPolicy,omitempty"`
	// AdditionalUserData is any additional user-data to be passed to the host
	AdditionalUserData []UserData `json:"additionalUserData,omitempty"`
	// SuspendProcesses disables the listed Scaling Policies
	SuspendProcesses []string `json:"suspendProcesses,omitempty"`
	// ExternalLoadBalancers define loadbalancers that should be attached to the instancegroup
	ExternalLoadBalancers []LoadBalancer `json:"externalLoadBalancers,omitempty"`
	// DetailedInstanceMonitoring defines if detailed-monitoring is enabled (AWS only)
	DetailedInstanceMonitoring *bool `json:"detailedInstanceMonitoring,omitempty"`
	// IAMProfileSpec defines the identity of the cloud group IAM profile (AWS only).
	IAM *IAMProfileSpec `json:"iam,omitempty"`
	// SecurityGroupOverride overrides the default security group created by Kops for this IG (AWS only).
	SecurityGroupOverride *string `json:"securityGroupOverride,omitempty"`
	// InstanceProtection makes new instances in an autoscaling group protected from scale in
	InstanceProtection *bool `json:"instanceProtection,omitempty"`
	// SysctlParameters will configure kernel parameters using sysctl(8). When
	// specified, each parameter must follow the form variable=value, the way
	// it would appear in sysctl.conf.
	SysctlParameters []string `json:"sysctlParameters,omitempty"`
	// RollingUpdate defines the rolling-update behavior
	RollingUpdate *RollingUpdate `json:"rollingUpdate,omitempty"`
	// InstanceInterruptionBehavior defines if a spot instance should be terminated, hibernated,
	// or stopped after interruption
	InstanceInterruptionBehavior *string `json:"instanceInterruptionBehavior,omitempty"`
}

const (
	// SpotAllocationStrategyLowestPrices indicates a lowest-price strategy
	SpotAllocationStrategyLowestPrices = "lowest-price"
	// SpotAllocationStrategyDiversified indicates a diversified strategy
	SpotAllocationStrategyDiversified = "diversified"
	// SpotAllocationStrategyCapacityOptimized indicates a capacity optimized strategy
	SpotAllocationStrategyCapacityOptimized = "capacity-optimized"
)

// SpotAllocationStrategies is a collection of supported strategies
var SpotAllocationStrategies = []string{SpotAllocationStrategyLowestPrices, SpotAllocationStrategyDiversified, SpotAllocationStrategyCapacityOptimized}

// MixedInstancesPolicySpec defines the specification for an autoscaling group backed by a ec2 fleet
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
	// instance types to choose from.
	SpotAllocationStrategy *string `json:"spotAllocationStrategy,omitempty"`
	// SpotInstancePools is the number of Spot pools to use to allocate your Spot capacity (defaults to 2)
	// pools are determined from the different instance types in the Overrides array of LaunchTemplate
	SpotInstancePools *int64 `json:"spotInstancePools,omitempty"`
}

// UserData defines a user-data section
type UserData struct {
	// Name is the name of the user-data
	Name string `json:"name,omitempty"`
	// Type is the type of user-data
	Type string `json:"type,omitempty"`
	// Content is the user-data content
	Content string `json:"content,omitempty"`
}

// VolumeSpec defined the spec for an additional volume attached to the instance group
type VolumeSpec struct {
	// DeleteOnTermination configures volume retention policy upon instance termination.
	// The volume is deleted by default. Cluster deletion does not remove retained volumes.
	// NOTE: This setting applies only to the Launch Configuration and does not affect Launch Templates.
	DeleteOnTermination *bool `json:"deleteOnTermination,omitempty"`
	// Device is an optional device name of the block device
	Device string `json:"device,omitempty"`
	// Encrypted indicates you want to encrypt the volume
	Encrypted *bool `json:"encrypted,omitempty"`
	// Iops is the provision iops for this iops (think io1 in aws)
	Iops *int64 `json:"iops,omitempty"`
	// Size is the size of the volume in GB
	Size int64 `json:"size,omitempty"`
	// Type is the type of volume to create and is cloud specific
	Type string `json:"type,omitempty"`
}

// VolumeMountSpec defines the specification for mounting a device
type VolumeMountSpec struct {
	// Device is the device name to provision and mount
	Device string `json:"device,omitempty"`
	// Filesystem is the filesystem to mount
	Filesystem string `json:"filesystem,omitempty"`
	// FormatOptions is a collection of options passed when formatting the device
	FormatOptions []string `json:"formatOptions,omitempty"`
	// MountOptions is a collection of mount options
	MountOptions []string `json:"mountOptions,omitempty"`
	// Path is the location to mount the device
	Path string `json:"path,omitempty"`
}

// IAMProfileSpec is the AWS IAM Profile to attach to instances in this instance
// group. Specify the ARN for the IAM instance profile (AWS only).
type IAMProfileSpec struct {
	// Profile of the cloud group IAM profile. In aws this is the arn
	// for the iam instance profile
	Profile *string `json:"profile,omitempty"`
}

// LoadBalancer defines a load balancer
type LoadBalancer struct {
	// LoadBalancerName to associate with this instance group (AWS ELB)
	LoadBalancerName *string `json:"loadBalancerName,omitempty"`
	// TargetGroupARN to associate with this instance group (AWS ALB/NLB)
	TargetGroupARN *string `json:"targetGroupArn,omitempty"`
}
