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

package protokube

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/labels"
)

type nodePatch struct {
	Metadata *nodePatchMetadata `json:"metadata,omitempty"`
	Spec     *nodePatchSpec     `json:"spec,omitempty"`
}

type nodePatchMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

type nodePatchSpec struct {
	Unschedulable *bool `json:"unschedulable,omitempty"`
}

// ApplyMasterTaints finds masters that have not yet been tainted, and applies the master taint
// Once the kubelet support --taints (like --labels) this can probably go away entirely.
// It also sets the unschedulable flag to false, so pods (with a toleration) can target the node
func ApplyMasterTaints(kubeContext *KubernetesContext) error {
	client, err := kubeContext.KubernetesClient()
	if err != nil {
		return err
	}

	options := v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{"kubernetes.io/role": "master"}).String(),
	}
	glog.V(2).Infof("Querying k8s for nodes with selector %q", options.LabelSelector)
	nodes, err := client.Core().Nodes().List(options)
	if err != nil {
		return fmt.Errorf("error querying nodes: %v", err)
	}

	taint := []v1.Taint{{Key: "dedicated", Value: "master", Effect: "NoSchedule"}}
	taintJSON, err := json.Marshal(taint)
	if err != nil {
		return fmt.Errorf("error serializing taint: %v", err)
	}

	for i := range nodes.Items {
		node := &nodes.Items[i]

		nodeTaintJSON := node.Annotations[v1.TaintsAnnotationKey]
		if nodeTaintJSON != "" {
			if nodeTaintJSON != string(taintJSON) {
				glog.Infof("Node %q had unexpected taint: %v", node.Name, nodeTaintJSON)
			}
			continue
		}

		nodePatchMetadata := &nodePatchMetadata{
			Annotations: map[string]string{v1.TaintsAnnotationKey: string(taintJSON)},
		}
		unschedulable := false
		nodePatchSpec := &nodePatchSpec{
			Unschedulable: &unschedulable,
		}
		nodePatch := &nodePatch{
			Metadata: nodePatchMetadata,
			Spec:     nodePatchSpec,
		}
		nodePatchJson, err := json.Marshal(nodePatch)
		if err != nil {
			return fmt.Errorf("error building node patch: %v", err)
		}

		glog.V(2).Infof("sending patch for node %q: %q", node.Name, string(nodePatchJson))

		_, err = client.Nodes().Patch(node.Name, api.StrategicMergePatchType, nodePatchJson)
		if err != nil {
			// TODO: Should we keep going?
			return fmt.Errorf("error applying patch to node: %v", err)
		}
	}

	return nil
}
