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

package commands

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestUnsetInstanceGroupsBadInput(t *testing.T) {
	fields := []string{
		"bad-unset-input",
	}

	err := UnsetInstancegroupFields(fields, &kops.InstanceGroup{})
	if err == nil {
		t.Errorf("expected a field parsing error, but received none")
	}
}

func TestUnsetInstanceGroupsFields(t *testing.T) {
	grid := []struct {
		Fields []string
		Input  kops.InstanceGroup
		Output kops.InstanceGroup
	}{
		{
			Fields: []string{
				"spec.image",
			},
			Input: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image: "ami-test-1",
				},
			},
			Output: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{},
			},
		},
		{
			Fields: []string{
				"spec.machineType",
			},
			Input: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					MachineType: "m5.large",
				},
			},
			Output: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{},
			},
		},
		{
			Fields: []string{
				"spec.minSize",
				"spec.maxSize",
			},
			Input: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					MinSize: fi.PtrTo(int32(1)),
					MaxSize: fi.PtrTo(int32(3)),
				},
			},
			Output: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{},
			},
		},
		{
			Fields: []string{
				"spec.additionalSecurityGroups",
			},
			Input: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					AdditionalSecurityGroups: []string{
						"group1",
						"group2",
						"group3",
					},
				},
			},
			Output: kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{},
			},
		},
	}

	for _, g := range grid {
		ig := g.Input

		err := UnsetInstancegroupFields(g.Fields, &ig)
		if err != nil {
			t.Errorf("unexpected error from unsetClusterFields %v: %v", g.Fields, err)
			continue
		}

		if !reflect.DeepEqual(ig, g.Output) {
			t.Errorf("unexpected output from unsetClusterFields %v.  expected=%v, actual=%v", g.Fields, g.Output, ig)
			continue
		}

	}
}
