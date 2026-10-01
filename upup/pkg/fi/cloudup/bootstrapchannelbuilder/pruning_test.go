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
)

func TestBuildPruneDirectivesSkipPruning(t *testing.T) {
	manifest := []byte("apiVersion: karpenter.sh/v1\nkind: NodePool\nmetadata:\n  name: nodes-a\n")
	karpenterKinds := []schema.GroupKind{
		{Group: "karpenter.k8s.aws", Kind: "EC2NodeClass"},
		{Group: "karpenter.sh", Kind: "NodePool"},
	}

	grid := []struct {
		desc                    string
		protectedInstanceGroups []string
		// expectedFieldSelector is expected on the EC2NodeClass and NodePool prune directives only.
		expectedFieldSelector string
	}{
		{
			desc: "prune all objects that are not in the manifest",
		},
		{
			desc:                    "keep the objects of the protected instance groups",
			protectedInstanceGroups: []string{"nodes-b", "nodes-c"},
			expectedFieldSelector:   "metadata.name!=nodes-b,metadata.name!=nodes-c",
		},
	}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			spec := testAddonSpec("karpenter.sh")
			if err := buildPruneDirectives(spec, manifest, g.protectedInstanceGroups); err != nil {
				t.Fatalf("building prune directives: %v", err)
			}

			pruned := make(map[schema.GroupKind]bool)
			for _, kind := range spec.Prune.Kinds {
				gk := schema.GroupKind{Group: kind.Group, Kind: kind.Kind}
				pruned[gk] = true
				expected := ""
				if slices.Contains(karpenterKinds, gk) {
					expected = g.expectedFieldSelector
				}
				if kind.FieldSelector != expected {
					t.Errorf("field selector for %v = %q, want %q", gk, kind.FieldSelector, expected)
				}
			}
			for _, gk := range karpenterKinds {
				if !pruned[gk] {
					t.Errorf("%v is not pruned", gk)
				}
			}
		})
	}
}
