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
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

// InstanceGroup represents a group of instances (either nodes or masters) with the same configuration
type InstanceGroup struct {
	unversioned.TypeMeta `json:",inline"`
	ObjectMeta           api.ObjectMeta `json:"metadata,omitempty"`

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

func (g *InstanceGroup) IsMaster() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleMaster:
		return true
	case InstanceGroupRoleNode:
		return false

	default:
		glog.Fatalf("Role not set in group %v", g)
		return false
	}
}

func (g *InstanceGroup) Validate() error {
	if g.ObjectMeta.Name == "" {
		return field.Required(field.NewPath("Name"), "")
	}

	if g.Spec.Role == "" {
		return field.Required(field.NewPath("Role"), "Role must be set")
	}

	switch g.Spec.Role {
	case InstanceGroupRoleMaster:
	case InstanceGroupRoleNode:

	default:
		return field.Invalid(field.NewPath("Role"), g.Spec.Role, "Unknown role")
	}

	if g.IsMaster() {
		if len(g.Spec.Zones) == 0 {
			return fmt.Errorf("Master InstanceGroup %s did not specify any Zones", g.ObjectMeta.Name)
		}
	}

	return nil
}

// CrossValidate performs validation of the instance group, including that it is consistent with the Cluster
// It calls Validate, so all that validation is included.
func (g *InstanceGroup) CrossValidate(cluster *Cluster, strict bool) error {
	err := g.Validate()
	if err != nil {
		return err
	}

	// Check that instance groups are defined in valid zones
	{
		clusterZones := make(map[string]*ClusterZoneSpec)
		for _, z := range cluster.Spec.Zones {
			if clusterZones[z.Name] != nil {
				return fmt.Errorf("Zones contained a duplicate value: %v", z.Name)
			}
			clusterZones[z.Name] = z
		}

		for _, z := range g.Spec.Zones {
			if clusterZones[z] == nil {
				return fmt.Errorf("InstanceGroup %q is configured in %q, but this is not configured as a Zone in the cluster", g.ObjectMeta.Name, z)
			}
		}
	}

	return nil
}
