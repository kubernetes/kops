/*
Copyright 2017 The Kubernetes Authors.

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
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

// Test_FindSubnet tests FindSubnet
func Test_FindSubnet(t *testing.T) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "a"},
				{Name: "b"},
			},
		},
	}

	grid := []struct {
		cluster     *kops.Cluster
		subnet      string
		expectFound bool
	}{
		{
			cluster:     cluster,
			subnet:      "a",
			expectFound: true,
		},
		{
			cluster:     cluster,
			subnet:      "a",
			expectFound: true,
		},
		{
			cluster:     cluster,
			subnet:      "ab",
			expectFound: false,
		},
	}
	for _, g := range grid {
		actual := FindSubnet(g.cluster, g.subnet)
		if g.expectFound {
			if actual == nil {
				t.Errorf("did not find %q", g.subnet)
				continue
			}
			if actual.Name != g.subnet {
				t.Errorf("found but had wrong name: %q vs %q", g.subnet, actual.Name)
			}
		} else {
			if actual != nil {
				t.Errorf("unexpectedly found %q", g.subnet)
				continue
			}
		}
	}
}

// TestW_FindZonesOrRegionForInstanceGroup tests FindZonesOrRegionForInstanceGroup for a cluster
func Test_FindZonesOrRegionForInstanceGroup(t *testing.T) {
	clusterWithZones := &kops.Cluster{
		Spec: kops.ClusterSpec{
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "zonea", Zone: "zonea"},
				{Name: "zoneb", Zone: "zoneb"},
			},
		},
	}

	clusterWithRegion := &kops.Cluster{
		Spec: kops.ClusterSpec{
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "region1", Region: "region1"},
				{Name: "region2", Region: "region2"},
			},
		},
	}

	testcases := []struct {
		name        string
		cluster     *kops.Cluster
		ig          *kops.InstanceGroup
		expected    []string
		expectError bool
	}{
		{
			name:    "cluster with only 1 zone",
			cluster: clusterWithZones,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zonea"},
				},
			},
			expected: []string{"zonea"},
		},
		{
			name:    "cluster with multiple zones",
			cluster: clusterWithZones,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zonea", "zoneb"},
				},
			},
			expected: []string{"zonea", "zoneb"},
		},
		{
			name:    "cluster with 1 region",
			cluster: clusterWithRegion,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"region1"},
				},
			},
			expected: []string{"region1"},
		},
		{
			name:    "cluster with multiple regions",
			cluster: clusterWithRegion,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"region1", "region2"},
				},
			},
			expected: []string{"region1", "region2"},
		},
		{
			name:    "cluster with invalid zone",
			cluster: clusterWithZones,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zoneb", "nope"},
				},
			},
			expectError: true,
		},
		{
			name:    "cluster with invalid region",
			cluster: clusterWithRegion,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"region1", "nope"},
				},
			},
			expectError: true,
		},
		{
			name:    "cluster with existing zones",
			cluster: clusterWithZones,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Zones: []string{"directa", "directb"},
				},
			},
			expected: []string{"directa", "directb"},
		},
		{
			name:    "cluster with new and existing zones",
			cluster: clusterWithZones,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Zones:   []string{"directa", "directb"},
					Subnets: []string{"zonea", "zoneb"},
				},
			},
			// Not sure if we actually should merge here - should the IG zones actually act as a restriction on the Subnet zones?
			// For now, this isn't likely to come up, so we test the current behaviour
			expected: []string{"directa", "directb", "zonea", "zoneb"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := FindZonesOrRegionForInstanceGroup(tc.cluster, tc.ig)
			if err != nil {
				if tc.expectError {
					return
				}
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectError {
				t.Errorf("expected error, result was %v", actual)
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("unexpected result: actual=%v, expected=%v", actual, tc.expected)
			}
		})
	}
}
