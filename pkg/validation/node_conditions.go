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

package validation

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func getNodeReadyStatus(node *v1.Node) v1.ConditionStatus {
	cond := findNodeCondition(node, v1.NodeReady)
	if cond != nil {
		return cond.Status
	}
	return v1.ConditionUnknown
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

// isNodeReady returns if a Node is considered ready.
// It is considered ready if:
// 1) its Ready condition is set to true
// 2) doesn't have NetworkUnavailable condition set to true
func isNodeReady(node *v1.Node) bool {
	nodeReadyCondition := findNodeCondition(node, v1.NodeReady)
	if nodeReadyCondition == nil {
		klog.Warningf("v1.NodeReady condition not set on node %s", node.Name)
		return false
	}
	if nodeReadyCondition.Status != v1.ConditionTrue {
		klog.V(4).Infof("node %q not ready: %v", node.Name, nodeReadyCondition)
		return false
	}

	networkUnavailableCondition := findNodeCondition(node, v1.NodeNetworkUnavailable)
	if networkUnavailableCondition != nil {
		if networkUnavailableCondition.Status != v1.ConditionFalse && networkUnavailableCondition.Status != v1.ConditionUnknown {
			klog.V(4).Infof("node %q not ready: %v", node.Name, networkUnavailableCondition)
			return false
		}
	} else {
		klog.V(4).Infof("v1.NodeNetworkUnavailable condition not set on node %s", node.Name)
	}

	return true
}
