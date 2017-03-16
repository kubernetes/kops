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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

const (
	// How often to Poll pods, nodes and claims.
	Poll = 2 * time.Second

	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures
	// TODO: client should not apply this timeout to Watch calls. Increased from 30s until that is fixed.
	SingleCallTimeout = 5 * time.Minute
)

// TODO: Rename to NodeValidator; probably just convert to utility functions
// NodeAPIAdapter used to retrieve information about Nodes in K8s
// TODO: should we pool the api client connection? My initial thought is no.
type NodeAPIAdapter struct {
	// K8s API client this sucker talks to K8s directly - not kubectl, hard api call
	client kubernetes.Interface

	//TODO: convert to arg on WaitForNodeToBe
	// K8s timeout on method call
	timeout time.Duration
}

func NewNodeAPIAdapter(client kubernetes.Interface, timeout time.Duration) (*NodeAPIAdapter, error) {
	if client == nil {
		return nil, fmt.Errorf("client not provided")
	}
	return &NodeAPIAdapter{
		client:  client,
		timeout: timeout,
	}, nil
}

// GetAllNodes is a access to get all nodes from a cluster api
func (nodeAA *NodeAPIAdapter) GetAllNodes() (nodes *v1.NodeList, err error) {
	opts := metav1.ListOptions{}
	nodes, err = nodeAA.client.CoreV1().Nodes().List(opts)
	if err != nil {
		glog.V(4).Infof("getting nodes failed for node %v", err)
		return nil, err
	}

	return nodes, nil
}

// GetReadySchedulableNodesOrDie addresses the common use case of getting nodes you can do work on.
// 1) Needs to be schedulable.
// 2) Needs to be ready.
// If EITHER 1 or 2 is not true, most tests will want to ignore the node entirely.
func (nodeAA *NodeAPIAdapter) GetReadySchedulableNodes() ([]*v1.Node, error) {
	nodeList, err := nodeAA.waitListSchedulableNodes()
	if err != nil {
		return nil, fmt.Errorf("error from listing schedulable nodes: %v", err)
	}

	// previous tests may have cause failures of some nodes. Let's skip
	// 'Not Ready' nodes, just in case (there is no need to fail the test).
	filtered := FilterNodes(nodeList, isNodeSchedulable)
	return filtered, err

}

// WaitForNodeToBeReady returns whether node name is ready within timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBeReady(nodeName string) (bool, error) {
	return nodeAA.WaitForNodeToBe(nodeName, v1.NodeReady, v1.ConditionTrue)
}

// WaitForNodeToBeNotReady returns whether node is not ready (i.e. the
// readiness condition is anything but ready, e.g false or unknown) within
// timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBeNotReady(nodeName string) (bool, error) {
	return nodeAA.WaitForNodeToBe(nodeName, v1.NodeReady, v1.ConditionFalse, v1.ConditionUnknown)
}

// WaitForNodeToBe returns whether the names node condition state matches one of the expected values,
// within timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBe(nodeName string, conditionType v1.NodeConditionType, expected ...v1.ConditionStatus) (bool, error) {
	if nodeName == "" {
		return false, fmt.Errorf("nodeName was empty")
	}

	glog.V(4).Infof("Waiting up to %v for node %s condition %s to be %v", nodeAA.timeout, nodeName, conditionType, expected)

	var cond *v1.NodeCondition
	err := wait.PollImmediate(Poll, nodeAA.timeout, func() (bool, error) {
		node, err := nodeAA.client.Core().Nodes().Get(nodeName, metav1.GetOptions{})
		// FIXME this is not erroring on 500's for instance.  We will keep looping
		if err != nil {
			// TODO: Check if e.g. NotFound
			glog.V(4).Infof("Couldn't get node %s: %v", nodeName, err)
			return false, nil
		}
		cond = findNodeCondition(node, conditionType)
		if cond == nil {
			return false, nil
		}
		return conditionMatchesExpected(cond, expected...), nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			glog.V(4).Infof("Node %s didn't reach desired %s condition status (%v) within %v.  Actual=%v", nodeName, conditionType, expected, nodeAA.timeout, cond)
			return false, nil
		}
		// TODO: Should return error
		return false, nil
	} else {
		return true, nil
	}
}

// IsNodeConditionUnset check that node condition is not set
func isNodeConditionUnset(node *v1.Node, conditionType v1.NodeConditionType) bool {
	cond := findNodeCondition(node, conditionType)
	return cond == nil
}

func FilterNodes(nodeList *v1.NodeList, fn func(node *v1.Node) bool) []*v1.Node {
	var matches []*v1.Node
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		if fn(node) {
			matches = append(matches, node)
		}
	}
	return matches
}

// waitListSchedulableNodes is a wrapper around listing nodes supporting retries.
func (nodeAA *NodeAPIAdapter) waitListSchedulableNodes() (*v1.NodeList, error) {
	var nodeList *v1.NodeList
	err := wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {
		var err error
		nodeList, err = nodeAA.client.Core().Nodes().List(metav1.ListOptions{FieldSelector: "spec.unschedulable=false"})
		if err != nil {
			// error logging TODO
			return false, err
		}
		return err == nil, nil
	})

	if err != nil {
		// TODO logging
		return nil, err
	}
	return nodeList, err
}

func findNodeCondition(node *v1.Node, conditionType v1.NodeConditionType) *v1.NodeCondition {
	for i := range node.Status.Conditions {
		cond := &node.Status.Conditions[i]
		if cond.Type == conditionType {
			return cond
		}
	}
	return nil
}

func conditionMatchesExpected(cond *v1.NodeCondition, expected ...v1.ConditionStatus) bool {
	for _, e := range expected {
		if cond.Status == e {
			return true
		}
	}
	return false
}
func isNodeConditionSetAsExpected(node *v1.Node, conditionType v1.NodeConditionType, expected ...v1.ConditionStatus) bool {
	cond := findNodeCondition(node, conditionType)
	if cond == nil {
		glog.V(4).Infof("Couldn't find condition %v on node %v", conditionType, node.Name)
		return false
	}

	if conditionMatchesExpected(cond, expected...) {
		return true
	}

	glog.V(4).Infof(
		"Condition %s of node %s is %v instead of %v. Reason: %v, message: %v",
		conditionType, node.Name, cond.Status, expected, cond.Reason, cond.Message)
	return false
}

// Node is schedulable if:
// 1) doesn't have "unschedulable" field set
// 2) it's Ready condition is set to true
// 3) doesn't have NetworkUnavailable condition set to true
func isNodeSchedulable(node *v1.Node) bool {
	nodeReady := isNodeConditionSetAsExpected(node, v1.NodeReady, v1.ConditionTrue)

	// TODO: Combine
	networkUnavailable := isNodeConditionUnset(node, v1.NodeNetworkUnavailable)
	networkUnavailableSilent := isNodeConditionSetAsExpected(node, v1.NodeNetworkUnavailable, v1.ConditionFalse, v1.ConditionUnknown)

	networkReady := networkUnavailable || networkUnavailableSilent

	return !node.Spec.Unschedulable && nodeReady && networkReady
}

// Get The Status of a Node
func GetNodeConditionStatus(node *v1.Node) v1.ConditionStatus {
	cond := findNodeCondition(node, v1.NodeReady)
	if cond != nil {
		return cond.Status
	}
	return v1.ConditionUnknown
}

// Node is ready if:
// 1) its Ready condition is set to true
// 2) doesn't have NetworkUnavailable condition set to true
func IsNodeOrMasterReady(node *v1.Node) bool {
	nodeReady := isNodeConditionSetAsExpected(node, v1.NodeReady, v1.ConditionTrue)

	// TODO: Combine
	networkUnavailable := isNodeConditionUnset(node, v1.NodeNetworkUnavailable)
	networkUnavailableSilent := isNodeConditionSetAsExpected(node, v1.NodeNetworkUnavailable, v1.ConditionFalse, v1.ConditionUnknown)

	networkReady := networkUnavailable || networkUnavailableSilent

	return nodeReady && networkReady
}
