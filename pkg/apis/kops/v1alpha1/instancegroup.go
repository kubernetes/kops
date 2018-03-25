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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceGroup represents a group of instances (either nodes or masters) with the same configuration
type InstanceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceGroupSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstanceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole string describes the roles of the nodes in this InstanceGroup (master or nodes)
type InstanceGroupRole string

const (
	InstanceGroupRoleMaster InstanceGroupRole = "Master"
	InstanceGroupRoleNode   InstanceGroupRole = "Node"
)

// InstanceGroupSpec is the specification for a instanceGroup
type InstanceGroupSpec struct {
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
	// A collection of files assets for deployed cluster wide
	FileAssets []FileAssetSpec `json:"fileAssets,omitempty"`
	// Describes the tenancy of the instance group. Can be either default or dedicated.
	// Currently only applies to AWS.
	Tenancy string `json:"tenancy,omitempty"`
	// Kubelet overrides kubelet config from the ClusterSpec
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`
	// Taints indicates the kubernetes taints for nodes in this group
	Taints []string `json:"taints,omitempty"`
	// AdditionalUserData is any aditional user-data to be passed to the host
	AdditionalUserData []UserData `json:"additionalUserData,omitempty"`
	// Zones is the names of the Zones where machines in this instance group should be placed
	// This is needed for regional subnets (e.g. GCE), to restrict placement to particular zones
	Zones []string `json:"zones,omitempty"`
	// SuspendProcesses disables the listed Scaling Policies
	SuspendProcesses []string `json:"suspendProcesses,omitempty"`
	// DetailedInstanceMonitoring defines if detailed-monitoring is enabled (AWS only)
	DetailedInstanceMonitoring *bool `json:"detailedInstanceMonitoring,omitempty"`
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
