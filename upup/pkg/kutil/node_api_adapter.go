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

package kutil

import (
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/util/wait"
	"time"
)

const (
	// How often to Poll pods, nodes and claims.
	Poll = 2 * time.Second

	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures from failing tests.
	// TODO: client should not apply this timeout to Watch calls. Increased from 30s until that is fixed.
	SingleCallTimeout = 5 * time.Minute
)

type NodeAPIAdapter struct {
	client  *client.Client
	timeout time.Duration
}

// GetReadySchedulableNodesOrDie addresses the common use case of getting nodes you can do work on.
// 1) Needs to be schedulable.
// 2) Needs to be ready.
// If EITHER 1 or 2 is not true, most tests will want to ignore the node entirely.
func (nodeAA *NodeAPIAdapter) GetReadySchedulableNodes() (nodes *api.NodeList, err error) {

	nodes, err = nodeAA.waitListSchedulableNodes()

	if err != nil {
		return nil, err
	}

	// previous tests may have cause failures of some nodes. Let's skip
	// 'Not Ready' nodes, just in case (there is no need to fail the test).
	FilterNodes(nodes, func(node api.Node) bool {
		return isNodeSchedulable(&node), nil
	})
	return nodes, err

}

// WaitForNodeToBeReady returns whether node name is ready within timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBeReady(name string) bool {
	return nodeAA.WaitForNodeToBe(name, api.NodeReady, true)
}

// WaitForNodeToBeNotReady returns whether node name is not ready (i.e. the
// readiness condition is anything but ready, e.g false or unknown) within
// timeout.
func (nodeAA *NodeAPIAdapter) WaitForNodeToBeNotReady(name string) bool {
	return nodeAA.WaitForNodeToBe(name, api.NodeReady, false)
}

// WaitForNodeToBe returns whether node "name's" condition state matches wantTrue
// within timeout. If wantTrue is true, it will ensure the node condition status
// is ConditionTrue; if it's false, it ensures the node condition is in any state
// other than ConditionTrue (e.g. not true or unknown).
func (nodeAA *NodeAPIAdapter) WaitForNodeToBe(name string, conditionType api.NodeConditionType, wantTrue bool) bool {
	glog.V(4).Infof("Waiting up to %v for node %s condition %s to be %t", nodeAA.timeout, name, conditionType, wantTrue)
	for start := time.Now(); time.Since(start) < nodeAA.timeout; time.Sleep(Poll) {
		node, err := nodeAA.client.Nodes().Get(name)

		// FIXME this is not erroring on 500's for instance.  We will keep looping
		if err != nil {
			glog.V(4).Infof("Couldn't get node %s", name)
			continue
		}

		if IsNodeConditionSetAsExpected(node, conditionType, wantTrue) {
			return true
		}
	}
	glog.V(4).Infof("Node %s didn't reach desired %s condition status (%t) within %v", name, conditionType, wantTrue, nodeAA.timeout)
	return false
}

func IsNodeConditionSetAsExpectedSilent(node *api.Node, conditionType api.NodeConditionType, wantTrue bool) bool {
	return isNodeConditionSetAsExpected(node, conditionType, wantTrue, true)
}

func IsNodeConditionUnset(node *api.Node, conditionType api.NodeConditionType) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == conditionType {
			return false
		}
	}
	return true
}
func FilterNodes(nodeList *api.NodeList, fn func(node api.Node) bool) {
	var l []api.Node

	for _, node := range nodeList.Items {
		if fn(node) {
			l = append(l, node)
		}
	}
	nodeList.Items = l
}

func IsNodeConditionSetAsExpected(node *api.Node, conditionType api.NodeConditionType, wantTrue bool) bool {
	return isNodeConditionSetAsExpected(node, conditionType, wantTrue, false)
}

// waitListSchedulableNodes is a wrapper around listing nodes supporting retries.
func (nodeAA *NodeAPIAdapter) waitListSchedulableNodes() (nodes *api.NodeList, err error) {

	if wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {
		nodes, err = nodeAA.client.Nodes().List(api.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector()})
		if err != nil {
			return nil, err
		}
		return nodes, nil
	}) != nil {
		return nil, err
	}
	return nodes, nil
}

func isNodeConditionSetAsExpected(node *api.Node, conditionType api.NodeConditionType, wantTrue, silent bool) bool {
	// Check the node readiness condition (logging all).
	for _, cond := range node.Status.Conditions {
		// Ensure that the condition type and the status matches as desired.
		if cond.Type == conditionType {
			if (cond.Status == api.ConditionTrue) == wantTrue {
				return true
			} else {
				if !silent {
					glog.V(4).Infof(
						"Condition %s of node %s is %v instead of %t. Reason: %v, message: %v",
						conditionType, node.Name, cond.Status == api.ConditionTrue, wantTrue, cond.Reason, cond.Message)
				}
				return false
			}
		}
	}
	if !silent {
		glog.V(4).Infof("Couldn't find condition %v on node %v", conditionType, node.Name)
	}
	return false
}

// Node is schedulable if:
// 1) doesn't have "unschedulable" field set
// 2) it's Ready condition is set to true
// 3) doesn't have NetworkUnavailable condition set to true
func isNodeSchedulable(node *api.Node) bool {
	nodeReady := IsNodeConditionSetAsExpected(node, api.NodeReady, true)
	networkReady := IsNodeConditionUnset(node, api.NodeNetworkUnavailable) ||
		IsNodeConditionSetAsExpectedSilent(node, api.NodeNetworkUnavailable, false)
	return !node.Spec.Unschedulable && nodeReady && networkReady
}
