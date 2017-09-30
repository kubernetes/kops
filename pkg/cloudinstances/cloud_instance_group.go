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

package cloudinstances

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/client-go/pkg/api/v1"
	api "k8s.io/kops/pkg/apis/kops"
)

// CloudInstanceGroup is the cloud backing of InstanceGroup.
type CloudInstanceGroup struct {
	InstanceGroup     *api.InstanceGroup
	GroupName         string
	GroupTemplateName string
	Ready             []*CloudInstanceGroupMember
	NeedUpdate        []*CloudInstanceGroupMember
	MinSize           int
	MaxSize           int
}

// CloudInstanceGroupMember describes an instance in a CloudInstanceGroup group.
type CloudInstanceGroupMember struct {
	ID   *string
	Node *v1.Node
}

// NewCloudInstanceGroup creates a CloudInstanceGroup and validates its initial values.
func NewCloudInstanceGroup(groupName string, groupTemplateName string, ig *api.InstanceGroup, minSize int, maxSize int) (*CloudInstanceGroup, error) {
	if groupName == "" {
		return nil, fmt.Errorf("group name for cloud instance group must be set")
	}
	if groupTemplateName == "" {
		return nil, fmt.Errorf("group template name for cloud instance group must be set")
	}
	if ig == nil {
		return nil, fmt.Errorf("kops instance group for cloud instance group must be set")
	}

	if minSize < 0 {
		return nil, fmt.Errorf("cloud instance group min size must be zero or greater")
	}
	if maxSize < 0 {
		return nil, fmt.Errorf("cloud instance group max size must be zero or greater")
	}

	cg := &CloudInstanceGroup{
		GroupName:         groupName,
		GroupTemplateName: groupTemplateName,
		InstanceGroup:     ig,
		MinSize:           minSize,
		MaxSize:           maxSize,
	}

	return cg, nil
}

// NewCloudInstanceGroupMember creates a new CloudInstanceGroupMember
func (c *CloudInstanceGroup) NewCloudInstanceGroupMember(instanceId *string, newGroupName string, currentGroupName string, nodeMap map[string]*v1.Node) error {
	if instanceId == nil {
		return fmt.Errorf("instance id for cloud instance member cannot be nil")
	}
	cm := &CloudInstanceGroupMember{
		ID: instanceId,
	}
	id := *instanceId
	node := nodeMap[id]
	if node != nil {
		cm.Node = node
	} else {
		glog.V(8).Infof("unable to find node for instance: %s", id)
	}

	if newGroupName == currentGroupName {
		c.Ready = append(c.Ready, cm)
	} else {
		c.NeedUpdate = append(c.NeedUpdate, cm)
	}

	return nil
}

// Status returns a human-readable Status indicating whether an update is needed
func (c *CloudInstanceGroup) Status() string {
	if len(c.NeedUpdate) == 0 {
		return "Ready"
	} else {
		return "NeedsUpdate"
	}
}

// GetNodeMap returns a list of nodes keyed by there external id
func GetNodeMap(nodes []v1.Node) map[string]*v1.Node {
	nodeMap := make(map[string]*v1.Node)
	for i := range nodes {
		node := &nodes[i]
		nodeMap[node.Spec.ExternalID] = node
	}

	return nodeMap
}

// GetInstanceGroup filters a list of instancegroups for recognized cloud groups
func GetInstanceGroup(name string, clusterName string, instancegroups []*api.InstanceGroup) (*api.InstanceGroup, error) {
	var instancegroup *api.InstanceGroup
	for _, g := range instancegroups {
		var groupName string
		switch g.Spec.Role {
		case api.InstanceGroupRoleMaster:
			groupName = g.ObjectMeta.Name + ".masters." + clusterName
		case api.InstanceGroupRoleNode:
			groupName = g.ObjectMeta.Name + "." + clusterName
		case api.InstanceGroupRoleBastion:
			groupName = g.ObjectMeta.Name + "." + clusterName
		default:
			glog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
			continue
		}

		if name == groupName {
			if instancegroup != nil {
				return nil, fmt.Errorf("found multiple instance groups matching ASG %q", groupName)
			}
			instancegroup = g
		}
	}

	return instancegroup, nil
}
