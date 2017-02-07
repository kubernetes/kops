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

package validation

import (
	"fmt"
	"time"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/api/v1"
	k8s_clientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// A cluster to validate
type ValidationCluster struct {
	MastersReady         bool              `json:"mastersReady,omitempty"`
	MastersReadyArray    []*ValidationNode `json:"mastersReadyArray,omitempty"`
	MastersNotReadyArray []*ValidationNode `json:"mastersNotReadyArray,omitempty"`
	MastersCount         int               `json:"mastersCount,omitempty"`

	NodesReady         bool              `json:"nodesReady,omitempty"`
	NodesReadyArray    []*ValidationNode `json:"nodesReadyArray,omitempty"`
	NodesNotReadyArray []*ValidationNode `json:"nodesNotReadyArray,omitempty"`
	NodesCount         int               `json:"nodesCount,omitempty"`

	NodeList *v1.NodeList `json:"nodeList,omitempty"`
}

// A K8s node to be validated
type ValidationNode struct {
	Zone     string             `json:"zone,omitempty"`
	Role     string             `json:"role,omitempty"`
	Hostname string             `json:"hostname,omitempty"`
	Status   v1.ConditionStatus `json:"status,omitempty"`
}

// ValidateCluster validate a k8s cluster with a provided instance group list
func ValidateCluster(clusterName string, instanceGroupList *kops.InstanceGroupList, clusterKubernetesClient k8s_clientset.Interface) (*ValidationCluster, error) {
	var instanceGroups []*kops.InstanceGroup
	validationCluster := &ValidationCluster{}

	for i := range instanceGroupList.Items {
		ig := &instanceGroupList.Items[i]
		instanceGroups = append(instanceGroups, ig)
		if ig.Spec.Role == kops.InstanceGroupRoleMaster {
			validationCluster.MastersCount += int(fi.Int32Value(ig.Spec.MinSize))
		} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
			validationCluster.NodesCount += int(fi.Int32Value(ig.Spec.MinSize))
		}
	}

	if len(instanceGroups) == 0 {
		return validationCluster, fmt.Errorf("No InstanceGroup objects found")
	}

	timeout, err := time.ParseDuration("30s")
	if err != nil {
		return nil, fmt.Errorf("Cannot set timeout %q: %v", clusterName, err)
	}

	nodeAA, err := NewNodeAPIAdapter(clusterKubernetesClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("error building node adapter for %q: %v", clusterName, err)
	}

	validationCluster.NodeList, err = nodeAA.GetAllNodes()
	if err != nil {
		return nil, fmt.Errorf("Cannot get nodes for %q: %v", clusterName, err)
	}

	return validateTheNodes(clusterName, validationCluster)

}

func getRoleNode(node *v1.Node) string {
	role := kops.RoleNodeLabelValue
	if val, ok := node.ObjectMeta.Labels[kops.RoleLabelName]; ok {
		role = val
	}

	return role
}

func validateTheNodes(clusterName string, validationCluster *ValidationCluster) (*ValidationCluster, error) {
	nodes := validationCluster.NodeList

	if nodes == nil || len(nodes.Items) == 0 {
		return validationCluster, fmt.Errorf("No nodes found in validationCluster")
	}

	for i := range nodes.Items {
		node := &nodes.Items[i]

		role := getRoleNode(node)

		n := &ValidationNode{
			Zone:     node.ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"],
			Hostname: node.ObjectMeta.Labels["kubernetes.io/hostname"],
			Role:     role,
			Status:   GetNodeConditionStatus(node),
		}

		ready := IsNodeOrMasterReady(node)

		// TODO: Use instance group role instead...
		if n.Role == kops.RoleMasterLabelValue {
			if ready {
				validationCluster.MastersReadyArray = append(validationCluster.MastersReadyArray, n)
			} else {
				validationCluster.MastersNotReadyArray = append(validationCluster.MastersNotReadyArray, n)
			}
		} else if n.Role == kops.RoleNodeLabelValue {
			if ready {
				validationCluster.NodesReadyArray = append(validationCluster.NodesReadyArray, n)
			} else {
				validationCluster.NodesNotReadyArray = append(validationCluster.NodesNotReadyArray, n)
			}

		}
	}

	validationCluster.MastersReady = true
	if len(validationCluster.MastersNotReadyArray) != 0 || validationCluster.MastersCount != len(validationCluster.MastersReadyArray) {
		validationCluster.MastersReady = false
	}

	validationCluster.NodesReady = true
	if len(validationCluster.NodesNotReadyArray) != 0 || validationCluster.NodesCount != len(validationCluster.NodesReadyArray) {
		validationCluster.NodesReady = false
	}

	if validationCluster.MastersReady && validationCluster.NodesReady {
		return validationCluster, nil
	} else {
		// TODO: This isn't an error...
		return validationCluster, fmt.Errorf("Your cluster is NOT ready %s", clusterName)
	}
}
