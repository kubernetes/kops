package v1alpha1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// InstanceGroup represents a group of instances (either nodes or masters) with the same configuration
type InstanceGroup struct {
	unversioned.TypeMeta `json:",inline"`
	ObjectMeta    `json:"metadata,omitempty"`

	Spec InstanceGroupSpec `json:"spec,omitempty"`
}

type InstanceGroupList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole string describes the roles of the nodes in this InstanceGroup (master or nodes)
type InstanceGroupRole string

const (
	InstanceGroupRoleMaster InstanceGroupRole = "Master"
	InstanceGroupRoleNode   InstanceGroupRole = "Node"
)

type InstanceGroupSpec struct {
	// Type determines the role of instances in this group: masters or nodes
	Role InstanceGroupRole `json:"role,omitempty"`

	Image   string `json:"image,omitempty"`
	MinSize *int   `json:"minSize,omitempty"`
	MaxSize *int   `json:"maxSize,omitempty"`
	//NodeInstancePrefix string `json:",omitempty"`
	//NodeLabels         string `json:",omitempty"`
	MachineType string `json:"machineType,omitempty"`
	//NodeTag            string `json:",omitempty"`

	// RootVolumeSize is the size of the EBS root volume to use, in GB
	RootVolumeSize *int `json:"rootVolumeSize,omitempty"`
	// RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
	RootVolumeType *string `json:"rootVolumeType,omitempty"`

	Zones []string `json:"zones,omitempty"`

	// MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
	MaxPrice *string `json:"maxPrice,omitempty"`

	// AssociatePublicIP is true if we want instances to have a public IP
	AssociatePublicIP *bool `json:"associatePublicIp,omitempty"`

	// CloudLabels indicates the labels for instances in this group, at the AWS level
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`

	// NodeLabels indicates the kubernetes labels for nodes in this group
	NodeLabels map[string]string `json:"nodeLabels,omitempty"`
}
