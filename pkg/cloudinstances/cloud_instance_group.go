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
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
)

// TerminationPolicy is a terminating policy for the cloud instance group
type TerminationPolicy string

const (
	// TerminateDefault indicates set the default policy for the cloud provider
	TerminateDefault = "default"
	// TerminateNewest indicates we should terminate the newest
	TerminateNewest = "newest"
	// TerminateOldest inducated we should terminate the oldest
	TerminateOldest = "oldest"
)

var (
	// TerminateNewestPolicy is just a wrapper to save typing
	TerminateNewestPolicy = []TerminationPolicy{TerminateNewest}
	// TerminateOldestPolicy is just a wrapper to save typing
	TerminateOldestPolicy = []TerminationPolicy{TerminateOldest}
)

// CloudInstanceGroup is the cloud backing of InstanceGroup.
type CloudInstanceGroup struct {
	// HumanName is a user-friendly name for the group
	HumanName string
	// InstanceGroup is a reference to the actual instanceGroup
	InstanceGroup *api.InstanceGroup
	// MaxSize is the maximum size of the instanceGroup
	MaxSize int
	// MinSize is the minimum size of the instanceGroup
	MinSize int
	// NeedUpdate is the number of instances which are required to be updated
	NeedUpdate []*CloudInstanceGroupMember
	// Raw allows for the implementer to attach an object, for tracking additional state
	Raw interface{}
	// Ready is the number of instances ready
	Ready []*CloudInstanceGroupMember
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
func (c *CloudInstanceGroup) NewCloudInstanceGroupMember(id string, newGroupName string, currentGroupName string, nodeMap map[string]*v1.Node) error {
	if id == "" {
		return fmt.Errorf("instance id for cloud instance member cannot be empty")
	}
	cm := &CloudInstanceGroupMember{ID: id, CloudInstanceGroup: c}

	if node := nodeMap[id]; node != nil {
		cm.Node = node
	} else {
		glog.V(8).Infof("unable to find node for instance: %s", id)
	}

	switch newGroupName {
	case currentGroupName:
		c.Ready = append(c.Ready, cm)
	default:
		c.NeedUpdate = append(c.NeedUpdate, cm)
	}

	return nil
}

// Size returns the size of the group
func (c *CloudInstanceGroup) Size() int {
	return len(c.Ready) + len(c.NeedUpdate)
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
	sv, _ := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)

	nodeMap := make(map[string]*v1.Node)
	for i := range nodes {
		node := &nodes[i]
		//ExternalID is deprecated in kubernetes 1.11 https://github.com/kubernetes/kubernetes/pull/61877
		//Mappings from ExternalID https://github.com/kubernetes/kubernetes/issues/61966#issuecomment-377659476
		if sv.Major == 1 && sv.Minor < 10 {
			nodeMap[node.Spec.ExternalID] = node
		} else {
			providerIDs := strings.Split(node.Spec.ProviderID, "/")
			instanceID := providerIDs[len(providerIDs)-1]
			nodeMap[instanceID] = node
		}
	}

	return nodeMap
}
