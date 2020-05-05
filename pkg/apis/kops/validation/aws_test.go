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

package validation

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
)

func TestValidateInstanceGroupSpec(t *testing.T) {
	grid := []struct {
		Input          kops.InstanceGroupSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"sg-1234abcd"},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"sg-1234abcd", ""},
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalSecurityGroups[1]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{" ", ""},
			},
			ExpectedErrors: []string{
				"Invalid value::spec.additionalSecurityGroups[0]",
				"Invalid value::spec.additionalSecurityGroups[1]",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"--invalid"},
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalSecurityGroups[0]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "t2.micro",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "t2.invalidType",
			},
			ExpectedErrors: []string{"Invalid value::test-nodes.spec.machineType"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "m5.large",
				Image:       "k8s-1.9-debian-stretch-amd64-hvm-ebs-2018-03-11",
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "c5.large",
				Image:       "k8s-1.9-debian-stretch-amd64-hvm-ebs-2018-03-11",
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(55),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(380),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(125),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(120),
			},
			ExpectedErrors: []string{},
		},
	}
	for _, g := range grid {
		ig := &kops.InstanceGroup{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-nodes",
			},
			Spec: g.Input,
		}
		errs := awsValidateInstanceGroup(ig)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}
