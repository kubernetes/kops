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

package protokube

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// bootstrapMasterNodeLabels applies labels to the current node so that it acts as a master
func bootstrapMasterNodeLabels(ctx context.Context, kubeContext *KubernetesContext, nodeName string) error {
	client, err := kubeContext.KubernetesClient()
	if err != nil {
		return err
	}

	klog.V(2).Infof("Querying k8s for node %q", nodeName)
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error querying node %q: %v", nodeName, err)
	}

	labels := map[string]string{
		"node-role.kubernetes.io/master": "",
	}

	shouldPatch := false
	for k, v := range labels {
		actual, found := node.Labels[k]
		if !found || actual != v {
			shouldPatch = true
		}
	}

	if !shouldPatch {
		return nil
	}

	klog.V(2).Infof("patching node %q to add labels %v", nodeName, labels)

	nodePatchMetadata := &nodePatchMetadata{
		Labels: labels,
	}
	nodePatch := &nodePatch{
		Metadata: nodePatchMetadata,
	}

	nodePatchJson, err := json.Marshal(nodePatch)
	if err != nil {
		return fmt.Errorf("error building node patch: %v", err)
	}

	klog.V(2).Infof("sending patch for node %q: %q", node.Name, string(nodePatchJson))
	_, err = client.CoreV1().Nodes().Patch(ctx, node.Name, types.StrategicMergePatchType, nodePatchJson, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch to node: %v", err)
	}

	return nil
}
