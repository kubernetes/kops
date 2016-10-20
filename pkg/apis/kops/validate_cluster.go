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

	"github.com/golang/glog"
	//client_simple "k8s.io/kops/pkg/client/simple"
	//k8s_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
)

const (
	Node   = "node"
	Master = "master"
)

// A cluster to validate
type ValidationCluster struct {
	MastersReady    []*ValidationNode
	MastersNotReady []*ValidationNode
	// Leaving here if we are to determine which nodes are down
	//MastersInstanceGroups []*api.InstanceGroup
	MastersCount int

	NodesReady    []*ValidationNode
	NodesNotReady []*ValidationNode
	// Leaving here if we are to determine which nodes are down
	//NodesInstanceGroups []*api.InstanceGroup
	NodesCount int
}

// A K8s node to be validated
type ValidationNode struct {
	Zone     string
	Role     string
	Hostname string
	Status   v1.ConditionStatus
}

// ValidateClusterWithIg validate a k8s clsuter with a provided instance group list
func ValidateClusterWithIg(clusterName string, instanceGroupList *InstanceGroupList) (*v1.NodeList, error) {

	var instancegroups []*InstanceGroup
	validationCluster := &ValidationCluster{}
	for i := range instanceGroupList.Items {
		ig := &instanceGroupList.Items[i]
		instancegroups = append(instancegroups, ig)
		if ig.Spec.Role == InstanceGroupRoleMaster {
			// Leaving here if we are to determine which nodes are down
			//validationCluster.mastersInstanceGroups = append(validationCluster.mastersInstanceGroups, ig)
			validationCluster.MastersCount += *ig.Spec.MinSize
		} else {
			// Leaving here if we are to determine which nodes are down
			//validationCluster.nodesInstanceGroups = append(validationCluster.nodesInstanceGroups, ig)
			validationCluster.NodesCount += *ig.Spec.MinSize
		}
	}
	nodeAA := &NodeAPIAdapter{}

	timeout, err := time.ParseDuration("30s")

	if err != nil {
		return nil, fmt.Errorf("Cannot set timeout %q: %v", clusterName, err)
	}

	nodeAA.BuildNodeAPIAdapter(clusterName, timeout, "")

	nodes, err := nodeAA.GetAllNodes()

	if err != nil {
		return nil, fmt.Errorf("Cannot get nodes for %q: %v", clusterName, err)
	}

	if len(instancegroups) == 0 {
		return nodes, errors.New("No InstanceGroup objects found\n")
	}

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
			return nodes, fmt.Errorf("Cannot test if node is ready: %s", node.Name)
		}
		if n.Role == Master {
			if ready {
				validationCluster.MastersReady = append(validationCluster.MastersReady, n)
			} else {
				validationCluster.MastersNotReady = append(validationCluster.MastersNotReady, n)
			}
		} else if n.Role == Node {
			if ready {
				validationCluster.NodesReady = append(validationCluster.NodesReady, n)
			} else {
				validationCluster.NodesNotReady = append(validationCluster.NodesNotReady, n)
			}

		}

	}

	mastersReady := true
	nodesReady := true
	if len(validationCluster.MastersNotReady) != 0 || validationCluster.MastersCount !=
		len(validationCluster.MastersReady) {
		mastersReady = false
	}

	if len(validationCluster.NodesNotReady) != 0 || validationCluster.NodesCount !=
		len(validationCluster.NodesReady) {
		nodesReady = false
	}

	glog.Infof("validationCluster %+v", validationCluster)

	if mastersReady && nodesReady {
		return nodes, nil
	} else {
		return nodes, fmt.Errorf("You cluster is NOT ready %s", clusterName)
	}
}

// ValidateCluster does what it is named, validate a K8s cluster
/*
func FullValidateCluster(clusterName string, clientset client_simple.Clientset) (*v1.NodeList, error) {

	list, err := clientset.InstanceGroups(clusterName).List(k8s_api.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Cannot get instnacegroups for %q: %v", clusterName, err)
	}

	return ValidateClusterWithIg(clusterName, list)

}*/

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
