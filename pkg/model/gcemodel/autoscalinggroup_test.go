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

package gcemodel

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
)

func TestSplitToZones(t *testing.T) {
	testcases := []struct {
		name     string
		minSize  *int32
		zones    []string
		expected map[string]int
		role     kops.InstanceGroupRole
		mustErr  bool
	}{
		{
			name:    "no zones, minSize 2",
			minSize: fi.PtrTo(int32(2)),
			mustErr: true,
		},
		{
			name:    "1 zone, minSize 2",
			minSize: fi.PtrTo(int32(2)),
			zones:   []string{"us-central1-a"},
			expected: map[string]int{
				"us-central1-a": 2,
			},
		},
		{
			name:    "2 zones, minSize 2",
			minSize: fi.PtrTo(int32(2)),
			zones:   []string{"us-central1-a", "us-central1-b"},
			expected: map[string]int{
				"us-central1-a": 1,
				"us-central1-b": 1,
			},
		},
		{
			name:    "2 zones, minSize 3",
			minSize: fi.PtrTo(int32(3)),
			zones:   []string{"us-central1-a", "us-central1-b"},
			expected: map[string]int{
				"us-central1-a": 2,
				"us-central1-b": 1,
			},
		},
		{
			name:    "3 zones, minSize 2",
			minSize: fi.PtrTo(int32(2)),
			zones:   []string{"us-central1-a", "us-central1-b", "us-central1-c"},
			expected: map[string]int{
				"us-central1-a": 1,
				"us-central1-b": 1,
				"us-central1-c": 0,
			},
		},
		{
			name:    "1 zone, default minSize",
			minSize: nil, // Defaults to 1 by default
			zones:   []string{"us-central1-a"},
			expected: map[string]int{
				"us-central1-a": 1,
			},
		},
		{
			name:    "2 zones, default minSize (Node)",
			minSize: nil, // Defaults to 2 for Node
			role:    kops.InstanceGroupRoleNode,
			zones:   []string{"us-central1-a", "us-central1-b"},
			expected: map[string]int{
				"us-central1-a": 1,
				"us-central1-b": 1,
			},
		},
	}

	for _, g := range testcases {
		t.Run(g.name, func(t *testing.T) {
			ig := &kops.InstanceGroup{}
			ig.ObjectMeta.Name = "nodes"
			ig.Spec.Role = g.role
			ig.Spec.MinSize = g.minSize
			ig.Spec.Zones = g.zones

			b := &AutoscalingGroupModelBuilder{
				GCEModelContext: &GCEModelContext{
					KopsModelContext: &model.KopsModelContext{},
				},
			}

			actual, err := b.splitToZones(ig)

			if g.mustErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if !reflect.DeepEqual(actual, g.expected) {
				t.Errorf("unexpected result. expected=%v, actual=%v", g.expected, actual)
			}
		})
	}
}
