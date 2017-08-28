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

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/apis/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
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

	ComponentFailures []string `json:"componentFailures,omitempty"`
	PodFailures       []string `json:"podFailures,omitempty"`
}

// A K8s node to be validated
type ValidationNode struct {
	Zone     string             `json:"zone,omitempty"`
	Role     string             `json:"role,omitempty"`
	Hostname string             `json:"hostname,omitempty"`
	Status   v1.ConditionStatus `json:"status,omitempty"`
}

type ValidateClusterRetries struct {
	Cluster         *api.Cluster
	Clientset       simple.Clientset
	Interval        time.Duration
	ValidateRetries int
	K8sClient       kubernetes.Interface
}

// GetInstanceGroupsAndValidateCluster uses client set to get all instances groups, and then validates the cluster.
func (v ValidateClusterRetries) GetInstanceGroupsAndValidateCluster() error {
	// get the new list of ig
	list, err := v.Clientset.InstanceGroupsFor(v.Cluster).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to get instance groups: %v", err)
	}

	// validate the cluster
	if err := v.ValidateClusterWithRetries(list); err != nil {
		return fmt.Errorf("unable to validate cluster: %v", err)
	}
	return nil
}

// ValidateClusterWithRetries runs our validation methods on the K8s Cluster x times and then fails.
func (v ValidateClusterRetries) ValidateClusterWithRetries(instanceGroupList *api.InstanceGroupList) (err error) {

	// TODO - We are going to need to improve Validate to allow for more than one node, not master
	// TODO - going down at a time.
	for i := 0; i <= v.ValidateRetries; i++ {

		if _, err = ValidateCluster(v.Cluster.ObjectMeta.Name, instanceGroupList, v.K8sClient); err != nil {
			glog.Infof("Cluster did not validate, and waiting longer: %v.", err)
			time.Sleep(v.Interval / 2)
		} else {
			glog.Infof("Cluster validated.")
			return nil
		}

	}

	// FIXME handle force and cloud only
	// for loop is done, and did not end when the cluster validated
	return fmt.Errorf("cluster validation failed: %v", err)
}

// ValidateCluster calls k8s APIs to validate a cluster.
func (v *ValidateClusterRetries) ValidateCluster(instanceGroupList *api.InstanceGroupList) error {

	if _, err := ValidateCluster(v.Cluster.ObjectMeta.Name, instanceGroupList, v.K8sClient); err != nil {
		return fmt.Errorf("cluster %q did not pass validation: %v", v.Cluster.ObjectMeta.Name, err)
	}

	return nil

}

// ValidateCluster validate a k8s cluster with a provided instance group list.
func ValidateCluster(clusterName string, instanceGroupList *kops.InstanceGroupList, clusterKubernetesClient kubernetes.Interface) (*ValidationCluster, error) {
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
		return validationCluster, fmt.Errorf("no InstanceGroup objects found")
	}

	timeout, err := time.ParseDuration("10s")
	if err != nil {
		return nil, fmt.Errorf("cannot set timeout %q: %v", clusterName, err)
	}

	nodeAA, err := NewNodeAPIAdapter(clusterKubernetesClient, timeout, nil)
	if err != nil {
		return nil, fmt.Errorf("error building node adapter for %q: %v", clusterName, err)
	}

	validationCluster.NodeList, err = nodeAA.GetAllNodes()
	if err != nil {
		return nil, fmt.Errorf("cannot get nodes for %q: %v", clusterName, err)
	}

	validationCluster.ComponentFailures, err = collectComponentFailures(clusterKubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get component status for %q: %v", clusterName, err)
	}

	validationCluster.PodFailures, err = collectPodFailures(clusterKubernetesClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get pod health for %q: %v", clusterName, err)
	}

	return validateTheNodes(clusterName, validationCluster)

}

func collectComponentFailures(client kubernetes.Interface) (failures []string, err error) {
	componentList, err := client.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
	if err == nil {
		for _, component := range componentList.Items {
			for _, condition := range component.Conditions {
				if condition.Status != v1.ConditionTrue {
					failures = append(failures, component.Name)
				}
			}
		}
	}
	return
}

func collectPodFailures(client kubernetes.Interface) (failures []string, err error) {
	pods, err := client.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.Phase == v1.PodSucceeded {
				continue
			}
			for _, status := range pod.Status.ContainerStatuses {
				if !status.Ready {
					failures = append(failures, pod.Name)
				}
			}
		}
	}
	return
}

func validateTheNodes(clusterName string, validationCluster *ValidationCluster) (*ValidationCluster, error) {
	nodes := validationCluster.NodeList

	if nodes == nil || len(nodes.Items) == 0 {
		return validationCluster, fmt.Errorf("No nodes found in validationCluster")
	}

	for i := range nodes.Items {
		node := &nodes.Items[i]

		role := util.GetNodeRole(node)
		if role == "" {
			role = "node"
		}

		n := &ValidationNode{
			Zone:     node.ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"],
			Hostname: node.ObjectMeta.Labels["kubernetes.io/hostname"],
			Role:     role,
			Status:   GetNodeConditionStatus(node),
		}

		ready := IsNodeOrMasterReady(node)

		// TODO: Use instance group role instead...
		if n.Role == "master" {
			if ready {
				validationCluster.MastersReadyArray = append(validationCluster.MastersReadyArray, n)
			} else {
				validationCluster.MastersNotReadyArray = append(validationCluster.MastersNotReadyArray, n)
			}
		} else if n.Role == "node" {
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
	if len(validationCluster.NodesNotReadyArray) != 0 || validationCluster.NodesCount > len(validationCluster.NodesReadyArray) {
		validationCluster.NodesReady = false
	}

	if !validationCluster.MastersReady {
		return validationCluster, fmt.Errorf("your masters are NOT ready %s", clusterName)
	}

	if !validationCluster.NodesReady {
		return validationCluster, fmt.Errorf("your nodes are NOT ready %s", clusterName)
	}

	if len(validationCluster.ComponentFailures) != 0 {
		return validationCluster, fmt.Errorf("your components are NOT healthy %s", clusterName)
	}

	if len(validationCluster.PodFailures) != 0 {
		return validationCluster, fmt.Errorf("your kube-system pods are NOT healthy %s", clusterName)
	}

	return validationCluster, nil
}
