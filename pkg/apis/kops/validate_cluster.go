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
	"errors"
	"time"

	"fmt"

	//"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api/v1"
)

const (
	Node   = "node"
	Master = "master"
)

// A cluster to validate
type ValidationCluster struct {
	MastersReady         bool
	MastersReadyArray    []*ValidationNode
	MastersNotReadyArray []*ValidationNode
	MastersCount         int

	NodesReady         bool
	NodesReadyArray    []*ValidationNode
	NodesNotReadyArray []*ValidationNode
	NodesCount         int

	NodeList *v1.NodeList
}

// A K8s node to be validated
type ValidationNode struct {
	Zone     string
	Role     string
	Hostname string
	Status   v1.ConditionStatus
}

// ValidateClusterWithIg validate a k8s clsuter with a provided instance group list
func ValidateCluster(clusterName string, instanceGroupList *InstanceGroupList) (*ValidationCluster, error) {

	var instanceGroups []*InstanceGroup
	validationCluster := &ValidationCluster{}
	for i := range instanceGroupList.Items {
		ig := &instanceGroupList.Items[i]
		instanceGroups = append(instanceGroups, ig)
		if ig.Spec.Role == InstanceGroupRoleMaster {
			validationCluster.MastersCount += *ig.Spec.MinSize
		} else if ig.Spec.Role == InstanceGroupRoleNode {
			validationCluster.NodesCount += *ig.Spec.MinSize
		}
	}

	if len(instanceGroups) == 0 {
		return validationCluster, errors.New("No InstanceGroup objects found\n")
	}

	nodeAA := &NodeAPIAdapter{}

	timeout, err := time.ParseDuration("30s")

	if err != nil {
		return nil, fmt.Errorf("Cannot set timeout %q: %v", clusterName, err)
	}

	nodeAA.BuildNodeAPIAdapter(clusterName, timeout, "")

	validationCluster.NodeList, err = nodeAA.GetAllNodes()

	if err != nil {
		return nil, fmt.Errorf("Cannot get nodes for %q: %v", clusterName, err)
	}

	nodes := validationCluster.NodeList

	for _, node := range nodes.Items {

		role := Node
		if val, ok := node.ObjectMeta.Labels["kubernetes.io/role"]; ok {
			role = val
		}

		n := &ValidationNode{
			Zone:     node.ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"],
			Hostname: node.ObjectMeta.Labels["kubernetes.io/hostname"],
			Role:     role,
			Status:   GetNodeConditionStatus(node.Status.Conditions),
		}

		ready, err := IsNodeOrMasterReady(&node)
		if err != nil {
			return validationCluster, fmt.Errorf("Cannot test if node is ready: %s", node.Name)
		}
		if n.Role == Master {
			if ready {
				validationCluster.MastersReadyArray = append(validationCluster.MastersReadyArray, n)
			} else {
				validationCluster.MastersNotReadyArray = append(validationCluster.MastersNotReadyArray, n)
			}
		} else if n.Role == Node {
			if ready {
				validationCluster.NodesReadyArray = append(validationCluster.NodesReadyArray, n)
			} else {
				validationCluster.NodesNotReadyArray = append(validationCluster.NodesNotReadyArray, n)
			}

		}

	}

	validationCluster.MastersReady = true
	if len(validationCluster.MastersNotReadyArray) != 0 || validationCluster.MastersCount !=
		len(validationCluster.MastersReadyArray) {
		validationCluster.MastersReady = false
	}

	validationCluster.NodesReady = true
	if len(validationCluster.NodesNotReadyArray) != 0 || validationCluster.NodesCount !=
		len(validationCluster.NodesReadyArray) {
		validationCluster.NodesReady = false
	}

	if validationCluster.MastersReady && validationCluster.NodesReady {
		return validationCluster, nil
	} else {
		return validationCluster, fmt.Errorf("You cluster is NOT ready %s", clusterName)
	}
}

func GetNodeConditionStatus(nodeConditions []v1.NodeCondition) v1.ConditionStatus {
	s := v1.ConditionUnknown
	for _, element := range nodeConditions {
		if element.Type == "Ready" {
			s = element.Status
			break
		}
	}
	return s

}
