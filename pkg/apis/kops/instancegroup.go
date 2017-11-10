/*
Copyright 2016 The Kubernetes Authors.

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

package kops

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelClusterName is a cloud tag label
const LabelClusterName = "kops.k8s.io/cluster"

// NodeLabelInstanceGroup is a node label set to the name of the instance group
const NodeLabelInstanceGroup = "kops.k8s.io/instancegroup"

// Deprecated - use the new labels & taints node-role.kubernetes.io/master and node-role.kubernetes.io/node
const TaintNoScheduleMaster15 = "dedicated=master:NoSchedule"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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
	// Items is a collection of instancegroups
	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole describes the roles of the nodes in this InstanceGroup (master or nodes)
type InstanceGroupRole string

const (
	// InstanceGroupRoleMaster indicates a master node
	InstanceGroupRoleMaster InstanceGroupRole = "Master"
	// InstanceGroupRoleNode indicates a compute node
	InstanceGroupRoleNode InstanceGroupRole = "Node"
	// InstanceGroupRoleBastion indicates a bastion node
	InstanceGroupRoleBastion InstanceGroupRole = "Bastion"
)

// AllInstanceGroupRoles is a collection of roles a node may be
var AllInstanceGroupRoles = []InstanceGroupRole{
	InstanceGroupRoleNode,
	InstanceGroupRoleMaster,
	InstanceGroupRoleBastion,
}

// RolloutStrategy defines the strategy to use when performing rollouts on this instance group
type RolloutStrategy string

// DuplicatStrategy contains options for the duplication rollout
type DuplicatStrategy struct {
	// DuplicateInstanceGroup indicate we should copy the entire group
	DuplicatInstanceGroup bool `json:"duplicateInstanceGroup,omitempty"`
}

const (
	// DefaultRollout indicates the default one by one with a time interval and or drain
	DefaultRollout = "default"
	// DuplicateRollout indicates a duplication of instancegroup
	DuplicateRollout = "duplicate"
	// ScaleUpRollout indicates a scalled ASG rollout
	ScaleUpRollout = "scale-up"
)

// UpdateStrategy provides details about the rollout stratergy for a instancegroup
type UpdateStrategy struct {
	// Batch is a batch size to operate within
	Batch int `json:"batch,omitempty"`
	// Drain indicates if this group should be drained
	Drain bool `json:"drain,omitempty"`
	// DrainTimeout is the amount of time we wait for drain pods
	DrainTimeout *metav1.Duration `json:"drainTimeout,omitempty"`
	// Interval is the time to given between iterations
	Interval *metav1.Duration `json:"interval,omitempty"`
	// PostDrainDelay is the duration we wait after draining each node
	PostDrainDelay *metav1.Duration `json:"postDrainDelay,omitempty"`
	// Rollout defines the strategy to use when performing a rollout on this instance group
	Rollout RolloutStrategy `json:"rollout,omitempty"`
}

// InstanceGroupSpec is the specification for a instanceGroup
type InstanceGroupSpec struct {
	// Strategy is the strategy to use for this instances group
	Strategy *UpdateStrategy `json:"strategy,omitempty"`
	// Type determines the role of instances in this group: masters or nodes
	Role InstanceGroupRole `json:"role,omitempty"`
	// Image is the instance instance (ami etc) we should use
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
	// Subnets is the names of the Subnets (as specified in the Cluster) where machines in this instance group should be placed
	Subnets []string `json:"subnets,omitempty"`
	// Zones is the names of the Zones where machines in this instance group should be placed This is needed for regional subnets (e.g. GCE), to restrict placement to particular zones
	Zones []string `json:"zones,omitempty"`
	// Hooks is a list of hooks for this instanceGroup, note: these can override the cluster wide ones if required
	Hooks []HookSpec `json:"hooks,omitempty"`
	// MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
	MaxPrice *string `json:"maxPrice,omitempty"`
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
	// Tenancy sescribes the tenancy of the instance group. Can be either default or dedicated. Currently only applies to AWS.
	Tenancy string `json:"tenancy,omitempty"`
	// Kubelet overrides kubelet config from the ClusterSpec
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`
	// Taints indicates the kubernetes taints for nodes in this group
	Taints []string `json:"taints,omitempty"`
	// AdditionalUserData is any additional user-data to be passed to the host
	AdditionalUserData []UserData `json:"additionalUserData,omitempty"`
	// SuspendProcesses disables the listed Scaling Policies
	SuspendProcesses []string `json:"suspendProcesses,omitempty"`
	// DetailedInstanceMonitoring defines if detailed-monitoring is enabled (AWS only)
	DetailedInstanceMonitoring *bool `json:"detailedInstanceMonitoring,omitempty"`
	// IAMProfileSpec defines the identity of the cloud group iam profile (AWS only).
	IAM *IAMProfileSpec `json:"iam,omitempty"`
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

// IAMProfileSpec is the AWS IAM Profile to attach to instances in this instance
// group. Specify the ARN for the IAM instance profile (AWS only).
type IAMProfileSpec struct {
	// Profile is the AWS IAM Profile to attach to instances in this instance group.
	// Specify the ARN for the IAM instance profile. (AWS only)
	Profile *string `json:"profile,omitempty"`
}

// PerformAssignmentsInstanceGroups populates InstanceGroups with default values
func PerformAssignmentsInstanceGroups(groups []*InstanceGroup) error {
	names := map[string]bool{}
	for _, group := range groups {
		names[group.ObjectMeta.Name] = true
	}

	for _, group := range groups {
		// We want to give them a stable Name as soon as possible
		if group.ObjectMeta.Name == "" {
			// Loop to find the first unassigned name like `nodes-%d`
			i := 0
			for {
				key := fmt.Sprintf("nodes-%d", i)
				if !names[key] {
					group.ObjectMeta.Name = key
					names[key] = true
					break
				}
				i++
			}
		}
	}

	return nil
}

// IsMaster checks if instanceGroup is a master
func (g *InstanceGroup) IsMaster() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleMaster:
		return true
	case InstanceGroupRoleNode:
		return false
	case InstanceGroupRoleBastion:
		return false
	default:
		return false
	}
}

// IsBastion checks if instanceGroup is a bastion
func (g *InstanceGroup) IsBastion() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleMaster:
		return false
	case InstanceGroupRoleNode:
		return false
	case InstanceGroupRoleBastion:
		return true
	default:
		glog.Fatalf("Role not set in group %v", g)
		return false
	}
}

func (g *InstanceGroup) AddInstanceGroupNodeLabel() {
	if g.Spec.NodeLabels == nil {
		nodeLabels := make(map[string]string)
		nodeLabels[NodeLabelInstanceGroup] = g.Name
		g.Spec.NodeLabels = nodeLabels
	} else {
		g.Spec.NodeLabels[NodeLabelInstanceGroup] = g.Name
	}
}
