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

// Test_FindZonesForInstanceGroup tests FindZonesForInstanceGroup
func Test_FindZonesForInstanceGroup(t *testing.T) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "zonea", Zone: "zonea"},
				{Name: "zoneb", Zone: "zoneb"},
			},
		},
	}

	grid := []struct {
		cluster     *kops.Cluster
		ig          *kops.InstanceGroup
		expected    []string
		expectError bool
	}{
		{
			cluster: cluster,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zonea"},
				},
			},
			expected: []string{"zonea"},
		},
		{
			cluster: cluster,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zonea", "zoneb"},
				},
			},
			expected: []string{"zonea", "zoneb"},
		},
		{
			cluster: cluster,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zoneb", "zonea"},
				},
			},
			// Order is not preserved (they are in fact sorted)
			expected: []string{"zonea", "zoneb"},
		},
		{
			cluster: cluster,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"zoneb", "nope"},
				},
			},
			expectError: true,
		},
		{
			cluster: cluster,
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Zones: []string{"directa", "directb"},
				},
			},
			expected: []string{"directa", "directb"},
		},
		{
			cluster: cluster,
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
	for i, g := range grid {
		actual, err := FindZonesForInstanceGroup(g.cluster, g.ig)
		if err != nil {
			if g.expectError {
				continue
			}
			t.Errorf("unexpected error for %d: %v", i, err)
			continue
		}
		if g.expectError {
			t.Errorf("expected error for %d, result was %v", i, actual)
			continue
		}
		if !reflect.DeepEqual(actual, g.expected) {
			t.Errorf("unexpected result for %d: actual=%v, expected=%v", i, actual, g.expected)
			continue
		}
	}
}
