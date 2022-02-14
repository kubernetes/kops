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
	"k8s.io/klog/v2"
	kopsapi "k8s.io/kops/pkg/apis/kops"
)

// CloudInstanceGroup is the cloud backing of InstanceGroup.
type CloudInstanceGroup struct {
	// HumanName is a user-friendly name for the group
	HumanName     string
	InstanceGroup *kopsapi.InstanceGroup
	Ready         []*CloudInstance
	NeedUpdate    []*CloudInstance
	MinSize       int
	TargetSize    int
	MaxSize       int

	// Raw allows for the implementer to attach an object, for tracking additional state
	Raw interface{}
}

// NewCloudInstance creates a new CloudInstance
func (c *CloudInstanceGroup) NewCloudInstance(instanceId string, status string, node *v1.Node) (*CloudInstance, error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instance id for cloud instance member cannot be empty")
	}
	cm := &CloudInstance{
		ID:                 instanceId,
		CloudInstanceGroup: c,
		Roles:              []string{},
	}

	if status == CloudInstanceStatusUpToDate {
		c.Ready = append(c.Ready, cm)
	} else {
		c.NeedUpdate = append(c.NeedUpdate, cm)
	}

	cm.Status = status

	if node != nil {
		cm.Node = node
	} else {
		klog.V(8).Infof("unable to find node for instance: %s", instanceId)
	}
	return cm, nil
}

// Status returns a human-readable Status indicating whether an update is needed
func (c *CloudInstanceGroup) Status() string {
	if len(c.NeedUpdate) == 0 {
		return "Ready"
	}
	return "NeedsUpdate"
}

func (group *CloudInstanceGroup) AdjustNeedUpdate() {
	if group.Ready != nil {
		var newReady []*CloudInstance
		for _, member := range group.Ready {
			makeNotReady := false
			if member.Node != nil && member.Node.Annotations != nil {
				if _, ok := member.Node.Annotations["kops.k8s.io/needs-update"]; ok {
					makeNotReady = true
				}
			}

			if makeNotReady {
				group.NeedUpdate = append(group.NeedUpdate, member)
				member.Status = CloudInstanceStatusNeedsUpdate
			} else {
				newReady = append(newReady, member)
			}
		}
		group.Ready = newReady
	}
}

// GetNodeMap returns a list of nodes keyed by their external id
func GetNodeMap(nodes []v1.Node, cluster *kopsapi.Cluster) map[string]*v1.Node {
	nodeMap := make(map[string]*v1.Node)

	if kopsapi.CloudProviderID(cluster.Spec.CloudProvider) == kopsapi.CloudProviderAzure {
		for i := range nodes {
			node := &nodes[i]
			vmName, err := toAzureVMName(node.Spec.ProviderID)
			if err != nil {
				klog.Errorf("ignoring node %q with malformed provider ID: %s", node.Name, err)
				continue
			}
			nodeMap[vmName] = node
		}
		return nodeMap
	}

	for i := range nodes {
		node := &nodes[i]
		providerIDs := strings.Split(node.Spec.ProviderID, "/")
		instanceID := providerIDs[len(providerIDs)-1]
		nodeMap[instanceID] = node
	}

	return nodeMap
}

// toAzureVMName returns a VM name from the resource path stored in the provider ID.
func toAzureVMName(providerID string) (string, error) {
	l := strings.Split(providerID, "/")
	if len(l) != 13 {
		return "", fmt.Errorf("unexpected form of resource path: %q", providerID)
	}
	vmssName := l[10]
	idx := l[12]
	return fmt.Sprintf("%s_%s", vmssName, idx), nil
}
