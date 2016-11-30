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
	"time"

	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	// How often to Poll pods, nodes and claims.
	Poll = 2 * time.Second

	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures
	// TODO: client should not apply this timeout to Watch calls. Increased from 30s until that is fixed.
	SingleCallTimeout = 5 * time.Minute
)

// NodeAPIAdapter used to retrieve information about Nodes in K8s
// TODO: should we pool the api client connection? My initial thought is no.
type NodeAPIAdapter struct {
	// K8s API client this sucker talks to K8s directly - not kubectl, hard api call

	client interface{}
	// K8s timeout on method call
	timeout time.Duration
	// K8s node name if applicable
	nodeName string
}

// Create a NodeAPIAdapter with K8s client based on the current kubectl config
// TODO I do not really like this .... hrmm
func (nodeAA *NodeAPIAdapter) BuildNodeAPIAdapter(clusterName string, timeout time.Duration, nodeName string) (err error) {

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: clusterName}).ClientConfig()
	if err != nil {
		return fmt.Errorf("cannot load kubecfg settings for %q: %v", clusterName, err)
	}

	c, err := release_1_5.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("cannot build kube client for %q: %v", clusterName, err)
	}

	if err != nil {
		return fmt.Errorf("creating client go boom, %v", err)
	}

	nodeAA.client = c
	nodeAA.timeout = timeout
	nodeAA.nodeName = nodeName

	return nil
}

// GetAllNodes is a access to get all nodes from a cluster api
func (nodeAA *NodeAPIAdapter) GetAllNodes() (nodes *v1.NodeList, err error) {

	c, err := nodeAA.getClient()

	if err != nil {
		glog.V(4).Infof("getClient failed for node %s, %v", nodeAA.nodeName, err)
		return nil, err
	}

	opts := v1.ListOptions{}
	nodes, err = c.Nodes().List(opts)

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
func (nodeAA *NodeAPIAdapter) GetReadySchedulableNodes() (nodes *v1.NodeList, err error) {

	nodes, err = nodeAA.waitListSchedulableNodes()

	if err != nil {
		return nil, fmt.Errorf("GetReadySchedulableNodes go boom %v", err)
	}

	// previous tests may have cause failures of some nodes. Let's skip
	// 'Not Ready' nodes, just in case (there is no need to fail the test).
	FilterNodes(nodes, func(node v1.Node) (bool, error) {
		return isNodeSchedulable(&node)
	})
	return nodes, err

}

// WaitForNodeToBeReady returns whether node name is ready within timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBeReady() (bool, error) {
	return nodeAA.WaitForNodeToBe(v1.NodeReady, true)
}

// WaitForNodeToBeNotReady returns whether node name is not ready (i.e. the
// readiness condition is anything but ready, e.g false or unknown) within
// timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBeNotReady() (bool, error) {
	return nodeAA.WaitForNodeToBe(v1.NodeReady, false)
}

// WaitForNodeToBe returns whether node "name's" condition state matches wantTrue
// within timeout. If wantTrue is true, it will ensure the node condition status
// is ConditionTrue; if it's false, it ensures the node condition is in any state
// other than ConditionTrue (e.g. not true or unknown).
func (nodeAA *NodeAPIAdapter) WaitForNodeToBe(conditionType v1.NodeConditionType, wantTrue bool) (bool, error) {

	if err := nodeAA.isNodeNameDefined(); err != nil {
		return false, fmt.Errorf("isNodeNameDefined failed for node %s, %v", nodeAA.nodeName, err)
	}
	if err := nodeAA.isClientDefined(); err != nil {
		return false, fmt.Errorf("isClientDefined failed for node %s, %v", nodeAA.nodeName, err)
	}

	glog.V(4).Infof("Waiting up to %v for node %s condition %s to be %t", nodeAA.timeout, nodeAA.nodeName, conditionType, wantTrue)

	for start := time.Now(); time.Since(start) < nodeAA.timeout; time.Sleep(Poll) {

		c, err := nodeAA.getClient()

		if err != nil {
			glog.V(4).Infof("getClient failed for node %s, %v", nodeAA.nodeName, err)
			return false, err
		}

		node, err := c.Nodes().Get(nodeAA.nodeName)

		// FIXME this is not erroring on 500's for instance.  We will keep looping
		if err != nil {
			glog.V(4).Infof("Couldn't get node %s", nodeAA.nodeName)
			continue
		}
		iSet, err := IsNodeConditionSetAsExpected(node, conditionType, wantTrue)

		if err != nil {
			glog.V(4).Infof("IsNodeConditionSetAsExpected failed for node %s, %v", nodeAA.nodeName, err)
			return false, err
		}

		if iSet {
			return true, nil
		}
	}
	glog.V(4).Infof("Node %s didn't reach desired %s condition status (%t) within %v", nodeAA.nodeName, conditionType, wantTrue, nodeAA.timeout)
	return false, nil
}

// IsNodeConditionSetAsExpectedSilent node conidtion is
func IsNodeConditionSetAsExpectedSilent(node *v1.Node, conditionType v1.NodeConditionType, wantTrue bool) (bool, error) {
	return isNodeConditionSetAsExpected(node, conditionType, wantTrue, true)
}

// IsNodeConditionUnset check that node condition is not set
func IsNodeConditionUnset(node *v1.Node, conditionType v1.NodeConditionType) (bool, error) {

	if err := isNodeStatusDefined(node); err != nil {
		return false, err
	}

	for _, cond := range node.Status.Conditions {
		if cond.Type == conditionType {
			return false, nil
		}
	}
	return true, nil
}

func FilterNodes(nodeList *v1.NodeList, fn func(node v1.Node) (test bool, err error)) {
	var l []v1.Node

	for _, node := range nodeList.Items {
		test, err := fn(node)
		if err != nil {
			// FIXME error handling?
			return
		}
		if test {
			l = append(l, node)
		}
	}
	nodeList.Items = l
}

func IsNodeConditionSetAsExpected(node *v1.Node, conditionType v1.NodeConditionType, wantTrue bool) (bool, error) {
	return isNodeConditionSetAsExpected(node, conditionType, wantTrue, false)
}

// waitListSchedulableNodes is a wrapper around listing nodes supporting retries.
func (nodeAA *NodeAPIAdapter) waitListSchedulableNodes() (nodes *v1.NodeList, err error) {

	if err = nodeAA.isClientDefined(); err != nil {
		return nil, err
	}

	if wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {

		c, err := nodeAA.getClient()

		if err != nil {
			// error logging TODO
			return false, err
		}

		nodes, err = c.Nodes().List(v1.ListOptions{FieldSelector: "spec.unschedulable=false"})
		if err != nil {
			// error logging TODO
			return false, err
		}
		return err == nil, nil
	}) != nil {
		// TODO logging
		return nil, err
	}
	return nodes, err
}

func (nodeAA *NodeAPIAdapter) getClient() (*release_1_5.Clientset, error) {
	// FIXME double check
	if nodeAA.client == nil {
		return nil, errors.New("Client cannot be null")
	}
	c := nodeAA.client.(*release_1_5.Clientset)
	return c, nil
}

// TODO: remove slient bool ... but what is `wantTrue` defined as
func isNodeConditionSetAsExpected(node *v1.Node, conditionType v1.NodeConditionType, wantTrue, silent bool) (bool, error) {

	if err := isNodeStatusDefined(node); err != nil {
		return false, err
	}

	// Check the node readiness condition (logging all).
	for _, cond := range node.Status.Conditions {
		// Ensure that the condition type and the status matches as desired.
		if cond.Type == conditionType {
			if (cond.Status == v1.ConditionTrue) == wantTrue {
				return true, nil
			} else {
				if !silent {
					glog.V(4).Infof(
						"Condition %s of node %s is %v instead of %t. Reason: %v, message: %v",
						conditionType, node.Name, cond.Status == v1.ConditionTrue, wantTrue, cond.Reason, cond.Message)
				}
				return false, nil
			}
		}
	}
	if !silent {
		glog.V(4).Infof("Couldn't find condition %v on node %v", conditionType, node.Name)
	}
	return false, nil
}

// Node is schedulable if:
// 1) doesn't have "unschedulable" field set
// 2) it's Ready condition is set to true
// 3) doesn't have NetworkUnavailable condition set to true
func isNodeSchedulable(node *v1.Node) (bool, error) {
	nodeReady, err := IsNodeConditionSetAsExpected(node, v1.NodeReady, true)

	if err != nil {
		return false, err
	}

	networkUnval, err := IsNodeConditionUnset(node, v1.NodeNetworkUnavailable)

	if err != nil {
		return false, err
	}

	networkUnvalSilent, err := IsNodeConditionSetAsExpectedSilent(node, v1.NodeNetworkUnavailable, false)

	if err != nil {
		return false, err
	}

	networkReady := networkUnval || networkUnvalSilent

	return !node.Spec.Unschedulable && nodeReady && networkReady, nil
}

func (nodeAA *NodeAPIAdapter) isNodeNameDefined() error {

	if nodeAA.nodeName == "" {
		return errors.New("nodeName must be defined in nodeAA struct")
	}
	return nil
}

func (nodeAA *NodeAPIAdapter) isClientDefined() error {

	if nodeAA.client == nil {
		return errors.New("client must be defined in the struct")
	}
	return nil
}

func isNodeStatusDefined(node *v1.Node) error {

	if node == nil {
		return errors.New("node cannot be nil")
	}

	// FIXME how do I test this?
	/*
		if node.Status == nil {
			return errors.New("node.Status cannot be nil")
		}*/
	return nil
}

// Get The Status of a Node
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

// Node is ready if:
// 1) it's Ready condition is set to true
// 2) doesn't have NetworkUnavailable condition set to true
func IsNodeOrMasterReady(node *v1.Node) (bool, error) {
	nodeReady, err := IsNodeConditionSetAsExpected(node, v1.NodeReady, true)

	if err != nil {
		return false, err
	}

	networkUnval, err := IsNodeConditionUnset(node, v1.NodeNetworkUnavailable)

	if err != nil {
		return false, err
	}

	networkUnvalSilent, err := IsNodeConditionSetAsExpectedSilent(node, v1.NodeNetworkUnavailable, false)

	if err != nil {
		return false, err
	}

	networkReady := networkUnval || networkUnvalSilent

	return nodeReady && networkReady, nil
}
