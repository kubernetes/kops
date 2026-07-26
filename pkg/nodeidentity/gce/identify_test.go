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

package gce

import (
	"testing"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
)

func TestIGNameFromInstanceMetadata(t *testing.T) {
	grid := []struct {
		name        string
		clusterName string
		metadata    map[string]string
		expected    string
		expectErr   bool
	}{
		{
			name:        "karpenter instance",
			clusterName: "cluster.example.com",
			metadata: map[string]string{
				gcemetadata.MetadataKeyClusterName: "cluster.example.com",
				MetadataKeyInstanceGroupName:       "karpenter-nodes",
			},
			expected: "karpenter-nodes",
		},
		{
			name:        "cluster name mismatch",
			clusterName: "cluster.example.com",
			metadata: map[string]string{
				gcemetadata.MetadataKeyClusterName: "other.example.com",
				MetadataKeyInstanceGroupName:       "karpenter-nodes",
			},
			expectErr: true,
		},
		{
			name:        "missing cluster name",
			clusterName: "cluster.example.com",
			metadata: map[string]string{
				MetadataKeyInstanceGroupName: "karpenter-nodes",
			},
			expectErr: true,
		},
		{
			name:        "missing instance group name",
			clusterName: "cluster.example.com",
			metadata: map[string]string{
				gcemetadata.MetadataKeyClusterName: "cluster.example.com",
			},
			expectErr: true,
		},
	}

	for _, tc := range grid {
		t.Run(tc.name, func(t *testing.T) {
			instance := &compute.Instance{
				Name:     "instance-1",
				Metadata: &compute.Metadata{},
			}
			for k, v := range tc.metadata {
				instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{
					Key:   k,
					Value: new(v),
				})
			}

			igName, err := igNameFromInstanceMetadata(tc.clusterName, instance)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got igName %q", igName)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if igName != tc.expected {
				t.Errorf("igName = %q, want %q", igName, tc.expected)
			}
		})
	}
}
