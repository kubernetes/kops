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
	"net"
	"net/url"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
)

// ValidationCluster a cluster to validate.
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

	ComponentFailures []string              `json:"componentFailures,omitempty"`
	PodFailures       []string              `json:"podFailures,omitempty"`
	ErrorMessage      string                `json:"errorMessage,omitempty"`
	Status            string                `json:"status"`
	ClusterName       string                `json:"clusterName"`
	InstanceGroups    []*kops.InstanceGroup `json:"instanceGroups,omitempty"`
}

// ValidationNode is A K8s node to be validated.
type ValidationNode struct {
	Zone     string             `json:"zone,omitempty"`
	Role     string             `json:"role,omitempty"`
	Hostname string             `json:"hostname,omitempty"`
	Status   v1.ConditionStatus `json:"status,omitempty"`
}

const (
	ClusterValidationFailed = "FAILED"
	ClusterValidationPassed = "PASSED"
)

// HasPlaceHolderIP checks if the API DNS has been updated.
func HasPlaceHolderIP(clusterName string) (bool, error) {

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: clusterName}).ClientConfig()

	apiAddr, err := url.Parse(config.Host)
	if err != nil {
		return true, fmt.Errorf("unable to parse Kubernetes cluster API URL: %v", err)
	}

	hostAddrs, err := net.LookupHost(apiAddr.Host)
	if err != nil {
		return true, fmt.Errorf("unable to resolve Kubernetes cluster API URL dns: %v", err)
	}

	for _, h := range hostAddrs {
		if h == "203.0.113.123" {
			return true, nil
		}
	}

	return false, nil
}

// ValidateCluster validate a k8s cluster with a provided instance group list
func ValidateCluster(clusterName string, instanceGroupList *kops.InstanceGroupList, clusterKubernetesClient kubernetes.Interface) (*ValidationCluster, error) {
	var instanceGroups []*kops.InstanceGroup

	for i := range instanceGroupList.Items {
		ig := &instanceGroupList.Items[i]
		instanceGroups = append(instanceGroups, ig)
	}

	if len(instanceGroups) == 0 {
		return nil, fmt.Errorf("no InstanceGroup objects found")
	}

	timeout, err := time.ParseDuration("10s")
	if err != nil {
		return nil, fmt.Errorf("cannot set timeout %q: %v", clusterName, err)
	}

	nodeAA, err := NewNodeAPIAdapter(clusterKubernetesClient, timeout)
	if err != nil {
		return nil, fmt.Errorf("error building node adapter for %q: %v", clusterName, err)
	}

	validationCluster := &ValidationCluster{
		ClusterName:    clusterName,
		ErrorMessage:   ClusterValidationPassed,
		InstanceGroups: instanceGroups,
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
		return nil, fmt.Errorf("No nodes found in validationCluster")
	}
	// Needed for when NodesCount and MastersCounts are predefined, i.e tests
	presetNodeCount := validationCluster.NodesCount == 0
	presetMasterCount := validationCluster.MastersCount == 0
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
			if presetMasterCount {
				validationCluster.MastersCount++
			}
			if ready {
				validationCluster.MastersReadyArray = append(validationCluster.MastersReadyArray, n)
			} else {
				validationCluster.MastersNotReadyArray = append(validationCluster.MastersNotReadyArray, n)
			}
		} else if n.Role == "node" {
			if presetNodeCount {
				validationCluster.NodesCount++
			}
			if ready {
				validationCluster.NodesReadyArray = append(validationCluster.NodesReadyArray, n)
			} else {
				validationCluster.NodesNotReadyArray = append(validationCluster.NodesNotReadyArray, n)
			}
		} else {
			glog.Warningf("ignoring node with role %q", n.Role)
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
		validationCluster.Status = ClusterValidationFailed
		validationCluster.ErrorMessage = fmt.Sprintf("your masters are NOT ready %s", clusterName)
		return validationCluster, fmt.Errorf(validationCluster.ErrorMessage)
	}

	if !validationCluster.NodesReady {
		validationCluster.Status = ClusterValidationFailed
		validationCluster.ErrorMessage = fmt.Sprintf("your nodes are NOT ready %s", clusterName)
		return validationCluster, fmt.Errorf(validationCluster.ErrorMessage)
	}

	if len(validationCluster.ComponentFailures) != 0 {
		validationCluster.Status = ClusterValidationFailed
		validationCluster.ErrorMessage = fmt.Sprintf("your components are NOT healthy %s", clusterName)
		return validationCluster, fmt.Errorf(validationCluster.ErrorMessage)
	}

	if len(validationCluster.PodFailures) != 0 {
		validationCluster.Status = ClusterValidationFailed
		validationCluster.ErrorMessage = fmt.Sprintf("your kube-system pods are NOT healthy %s", clusterName)
		return validationCluster, fmt.Errorf(validationCluster.ErrorMessage)
	}

	return validationCluster, nil
}
