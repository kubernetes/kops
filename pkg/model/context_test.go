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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
)

func TestCloudTagsForInstanceGroup_Taints(t *testing.T) {
	grid := []struct {
		name     string
		taints   []string
		wantTags map[string]string
	}{
		{
			name:   "taint with value and effect",
			taints: []string{"foo=bar:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": "bar:NoSchedule",
			},
		},
		{
			name:   "taint without value (key:effect)",
			taints: []string{"foo:NoSchedule"},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": ":NoSchedule",
			},
		},
		{
			name: "mix of taint formats",
			taints: []string{
				"foo:NoSchedule",
				"bar=baz:PreferNoSchedule",
			},
			wantTags: map[string]string{
				"k8s.io/cluster-autoscaler/node-template/taint/foo": ":NoSchedule",
				"k8s.io/cluster-autoscaler/node-template/taint/bar": "baz:PreferNoSchedule",
			},
		},
		{
			name:     "taint without effect is skipped",
			taints:   []string{"foo"},
			wantTags: map[string]string{},
		},
	}

	for _, tc := range grid {
		t.Run(tc.name, func(t *testing.T) {
			cluster := testutils.BuildMinimalClusterAWS("testcluster.test.com")
			ig := &kops.InstanceGroup{}
			ig.ObjectMeta.Name = "nodes"
			ig.Spec.Role = kops.InstanceGroupSubRoleNode.Role()
			ig.Spec.Taints = tc.taints

			b := &KopsModelContext{
				IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
				AllInstanceGroups: []*kops.InstanceGroup{ig},
				InstanceGroups:    []*kops.InstanceGroup{ig},
			}

			tags, err := b.CloudTagsForInstanceGroup(ig)
			if err != nil {
				t.Fatalf("CloudTagsForInstanceGroup() error = %v", err)
			}

			for k, want := range tc.wantTags {
				if got := tags[k]; got != want {
					t.Errorf("tag %q = %q, want %q", k, got, want)
				}
			}
		})
	}
}
