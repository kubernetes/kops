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

package nodelabeler

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/nodelabels"
)

type nodePatch struct {
	Metadata *nodePatchMetadata `json:"metadata,omitempty"`
}

type nodePatchMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// BootstrapControlPlaneNodeLabels applies labels to the current node so that it acts as a control-plane.
// Safe to call repeatedly: the patch is skipped when the labels already match.
func BootstrapControlPlaneNodeLabels(ctx context.Context, client kubernetes.Interface, nodeName string, nodeLabel string) error {
	if nodeName == "" {
		return fmt.Errorf("node name is required")
	}

	klog.V(2).Infof("querying k8s for node %q", nodeName)
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("querying node %q: %w", nodeName, err)
	}

	labels := nodelabels.BuildMandatoryControlPlaneLabels(nodeLabel)

	shouldPatch := false
	for k, v := range labels {
		if actual, found := node.Labels[k]; !found || actual != v {
			shouldPatch = true
			break
		}
	}
	if !shouldPatch {
		return nil
	}

	klog.V(2).Infof("patching node %q to add labels %v", nodeName, labels)
	patch, err := json.Marshal(&nodePatch{
		Metadata: &nodePatchMetadata{Labels: labels},
	})
	if err != nil {
		return fmt.Errorf("building node patch: %w", err)
	}

	klog.V(2).Infof("sending patch for node %q: %s", node.Name, patch)
	if _, err := client.CoreV1().Nodes().Patch(ctx, node.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{FieldManager: "kops-channels"}); err != nil {
		return fmt.Errorf("patching node %q: %w", node.Name, err)
	}
	return nil
}
