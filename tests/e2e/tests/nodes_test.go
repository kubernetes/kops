/*
Copyright 2021 The Kubernetes Authors.

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

package tests

import (
	"testing"

	"k8s.io/kops/tests/e2e/tests/fluentest"
)

func TestSufficientControlPlaneNodes(t *testing.T) {
	h := fluentest.Harness{T: t}

	igs := h.InstanceGroups()

	var controlPlaneNodes []*fluentest.Node
	for _, node := range h.Nodes().MustItems(t) {
		if node.IsControlPlane() {
			controlPlaneNodes = append(controlPlaneNodes, node)
		}
	}

	minControlPlaneNodes := 0
	maxControlPlaneNodes := 0
	for _, ig := range igs {
		if ig.IsControlPlane() {
			minControlPlaneNodes += ig.MinSize()
			maxControlPlaneNodes += ig.MaxSize()
		}
	}

	t.Logf("found %d controlPlaneNodes", len(controlPlaneNodes))
	if len(controlPlaneNodes) < minControlPlaneNodes || len(controlPlaneNodes) > maxControlPlaneNodes {
		t.Errorf("unexpected number of control plane nodes; got %d, want in range [%d, %d]", len(controlPlaneNodes), minControlPlaneNodes, maxControlPlaneNodes)
	}
}
