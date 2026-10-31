/*
Copyright 2026 The Kubernetes Authors.

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

package bootstrapchannelbuilder

import (
	"slices"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
)

func TestBuildPruneDirectivesProtectedInstanceGroups(t *testing.T) {
	controlPlane := testInstanceGroup("control-plane", kops.InstanceGroupRoleControlPlane, kops.InstanceManagerCloudGroup)
	karpenterA := testInstanceGroup("karpenter-a", kops.InstanceGroupRoleNode, kops.InstanceManagerKarpenter)
	karpenterB := testInstanceGroup("karpenter-b", kops.InstanceGroupRoleNode, kops.InstanceManagerKarpenter)
	// Not managed by Karpenter, for example after being moved to CloudGroup.
	nodes := testInstanceGroup("nodes", kops.InstanceGroupRoleNode, kops.InstanceManagerCloudGroup)
	allInstanceGroups := []*kops.InstanceGroup{controlPlane, karpenterA, karpenterB, nodes}
	// The objects rendered when karpenter-a is updated; kOps names them after the InstanceGroup.
	manifest := []byte("apiVersion: karpenter.k8s.aws/v1\nkind: EC2NodeClass\nmetadata:\n  name: karpenter-a\n---\napiVersion: karpenter.sh/v1\nkind: NodePool\nmetadata:\n  name: karpenter-a\n")

	grid := []struct {
		desc              string
		allInstanceGroups []*kops.InstanceGroup
		instanceGroups    []*kops.InstanceGroup
		// expectedFieldSelector is expected on the EC2NodeClass and NodePool prune directives only.
		expectedFieldSelector string
	}{
		{
			desc:                  "all instance groups",
			allInstanceGroups:     allInstanceGroups,
			instanceGroups:        allInstanceGroups,
			expectedFieldSelector: "metadata.name!=karpenter-a,metadata.name!=karpenter-b",
		},
		{
			desc:                  "control plane only",
			allInstanceGroups:     allInstanceGroups,
			instanceGroups:        []*kops.InstanceGroup{controlPlane},
			expectedFieldSelector: "metadata.name!=karpenter-a,metadata.name!=karpenter-b,metadata.name!=nodes",
		},
		{
			desc:                  "node instance groups",
			allInstanceGroups:     allInstanceGroups,
			instanceGroups:        []*kops.InstanceGroup{karpenterA, karpenterB, nodes},
			expectedFieldSelector: "metadata.name!=karpenter-a,metadata.name!=karpenter-b",
		},
		{
			desc:                  "one Karpenter instance group",
			allInstanceGroups:     allInstanceGroups,
			instanceGroups:        []*kops.InstanceGroup{karpenterA},
			expectedFieldSelector: "metadata.name!=karpenter-a,metadata.name!=karpenter-b,metadata.name!=nodes",
		},
		{
			desc:              "no Karpenter instance groups",
			allInstanceGroups: []*kops.InstanceGroup{controlPlane, nodes},
			instanceGroups:    []*kops.InstanceGroup{controlPlane, nodes},
		},
		{
			desc:                  "control plane only, without Karpenter instance groups",
			allInstanceGroups:     []*kops.InstanceGroup{controlPlane, nodes},
			instanceGroups:        []*kops.InstanceGroup{controlPlane},
			expectedFieldSelector: "metadata.name!=nodes",
		},
	}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			spec := testAddonSpec("karpenter.sh")
			modelContext := &model.KopsModelContext{AllInstanceGroups: g.allInstanceGroups, InstanceGroups: g.instanceGroups}
			hashPruneSpec, err := buildPruneDirectives(spec, manifest, modelContext)
			if err != nil {
				t.Fatalf("building prune directives: %v", err)
			}
			if expected := g.expectedFieldSelector != ""; hashPruneSpec != expected {
				t.Errorf("hash prune spec = %v, want %v", hashPruneSpec, expected)
			}

			karpenterKinds := 0
			for _, kind := range spec.Prune.Kinds {
				gk := schema.GroupKind{Group: kind.Group, Kind: kind.Kind}
				expected := ""
				if slices.Contains(karpenterInstanceGroupKinds, gk) {
					expected = g.expectedFieldSelector
					karpenterKinds++
				}
				if kind.FieldSelector != expected {
					t.Errorf("field selector for %v = %q, want %q", gk, kind.FieldSelector, expected)
				}
			}
			if karpenterKinds != len(karpenterInstanceGroupKinds) {
				t.Errorf("found %d Karpenter kinds in the prune directives, want %d", karpenterKinds, len(karpenterInstanceGroupKinds))
			}
		})
	}
}
