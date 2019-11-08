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

package cloudinstances

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	api "k8s.io/kops/pkg/apis/kops"
)

// CloudInstanceGroup is the cloud backing of InstanceGroup.
type CloudInstanceGroup struct {
	// HumanName is a user-friendly name for the group
	HumanName     string
	InstanceGroup *api.InstanceGroup
	Ready         []*CloudInstanceGroupMember
	NeedUpdate    []*CloudInstanceGroupMember
	MinSize       int
	MaxSize       int

	// Raw allows for the implementer to attach an object, for tracking additional state
	Raw interface{}
}

// CloudInstanceGroupMember describes an instance in a CloudInstanceGroup group.
type CloudInstanceGroupMember struct {
	// ID is a unique identifier for the instance, meaningful to the cloud
	ID string
	// Node is the associated k8s instance, if it is known
	Node *v1.Node
	// CloudInstanceGroup is the managing CloudInstanceGroup
	CloudInstanceGroup *CloudInstanceGroup
}

// NewCloudInstanceGroupMember creates a new CloudInstanceGroupMember
func (c *CloudInstanceGroup) NewCloudInstanceGroupMember(instanceId string, newGroupName string, currentGroupName string, nodeMap map[string]*v1.Node) error {
	if instanceId == "" {
		return fmt.Errorf("instance id for cloud instance member cannot be empty")
	}
	cm := &CloudInstanceGroupMember{
		ID:                 instanceId,
		CloudInstanceGroup: c,
	}
	node := nodeMap[instanceId]
	if node != nil {
		cm.Node = node
	} else {
		klog.V(8).Infof("unable to find node for instance: %s", instanceId)
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
	}
	return "NeedsUpdate"
}

// GetNodeMap returns a list of nodes keyed by their external id
func GetNodeMap(nodes []v1.Node, cluster *kops.Cluster) map[string]*v1.Node {
	nodeMap := make(map[string]*v1.Node)
	delimiter := "/"
	// Alicloud CCM uses the "{region}.{instance-id}" of a instance as ProviderID.
	// We need to set delimiter to "." for Alicloud.
	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderALI {
		delimiter = "."
	}

	for i := range nodes {
		node := &nodes[i]
		providerIDs := strings.Split(node.Spec.ProviderID, delimiter)
		instanceID := providerIDs[len(providerIDs)-1]
		nodeMap[instanceID] = node
	}

	return nodeMap
}
