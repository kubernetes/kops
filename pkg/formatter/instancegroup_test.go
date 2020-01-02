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

package formatter

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestRenderInstanceGroupZones(t *testing.T) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "subnet1", Zone: "subnet1zone"},
				{Name: "subnet2", Zone: "subnet2zone"},
				{Name: "subnet3", Zone: "subnet3zone"},
			},
		},
	}

	grid := []struct {
		ig       *kops.InstanceGroup
		expected string
	}{
		{
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Zones: []string{"test1", "test2"},
				},
			},
			expected: "test1,test2",
		},
		{
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"subnet1", "subnet2"},
				},
			},
			expected: "subnet1zone,subnet2zone",
		},
		{
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"badsubnet"},
				},
			},
			expected: "",
		},
	}
	for _, g := range grid {
		f := RenderInstanceGroupZones(cluster)
		actual := f(g.ig)
		if actual != g.expected {
			t.Errorf("unexpected output: %q vs %q", g.expected, actual)
			continue
		}
	}
}

func TestRenderInstanceGroupSubnets(t *testing.T) {
	cluster := &kops.Cluster{}
	grid := []struct {
		ig       *kops.InstanceGroup
		expected string
	}{
		{
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"subnet"},
				},
			},
			expected: "subnet",
		},
		{
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Subnets: []string{"subnet1", "subnet2"},
				},
			},
			expected: "subnet1,subnet2",
		},
	}
	for _, g := range grid {
		f := RenderInstanceGroupSubnets(cluster)
		actual := f(g.ig)
		if actual != g.expected {
			t.Errorf("unexpected output: %q vs %q", g.expected, actual)
			continue
		}
	}
}
