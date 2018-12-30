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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestDefaultTaintsEnforcedBefore160(t *testing.T) {
	type param struct {
		ver       string
		role      kops.InstanceGroupRole
		taints    []string
		shouldErr bool
	}

	params := []param{
		{"1.5.0", kops.InstanceGroupRoleNode, []string{kops.TaintNoScheduleMaster15}, true},
		{"1.5.1", kops.InstanceGroupRoleNode, nil, false},
		{"1.5.2", kops.InstanceGroupRoleNode, []string{}, false},
		{"1.6.0", kops.InstanceGroupRoleNode, []string{kops.TaintNoScheduleMaster15}, false},
		{"1.6.1", kops.InstanceGroupRoleNode, []string{"Foo"}, false},
	}

	for _, p := range params {
		cluster := &kops.Cluster{Spec: kops.ClusterSpec{KubernetesVersion: p.ver}}
		ig := &kops.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: kops.InstanceGroupSpec{
				Taints: p.taints,
				Role:   p.role,
			},
		}

		err := CrossValidateInstanceGroup(ig, cluster, false)
		if p.shouldErr {
			if err == nil {
				t.Fatal("Expected error building kubelet config, received nil.")
			} else if !strings.Contains(err.Error(), "User-specified taints are not supported before kubernetes version 1.6.0") {
				t.Fatalf("Received an unexpected error validating taints: '%s'", err.Error())
			}
		} else {
			if err != nil {
				t.Fatalf("Received an unexpected error validating taints: '%s', params: '%v'", err.Error(), p)
			}
		}
	}
}

func s(v string) *string {
	return fi.String(v)
}
func TestValidateInstanceProfile(t *testing.T) {
	grid := []struct {
		Input          *kops.IAMProfileSpec
		ExpectedErrors []string
		ExpectedDetail string
	}{
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws:iam::123456789012:instance-profile/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws:iam::123456789012:instance-profile/has/path/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws-cn:iam::123456789012:instance-profile/has/path/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws-us-gov:iam::123456789012:instance-profile/has/path/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("42"),
			},
			ExpectedErrors: []string{"Invalid value::IAMProfile.Profile"},
			ExpectedDetail: "Instance Group IAM Instance Profile must be a valid aws arn such as arn:aws:iam::123456789012:instance-profile/KopsExampleRole",
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws:iam::123456789012:group/division_abc/subdivision_xyz/product_A/Developers"),
			},
			ExpectedErrors: []string{"Invalid value::IAMProfile.Profile"},
			ExpectedDetail: "Instance Group IAM Instance Profile must be a valid aws arn such as arn:aws:iam::123456789012:instance-profile/KopsExampleRole",
		},
	}

	for _, g := range grid {
		err := validateInstanceProfile(g.Input, field.NewPath("IAMProfile"))
		allErrs := field.ErrorList{}
		if err != nil {
			allErrs = append(allErrs, err)
		}
		testErrors(t, g.Input, allErrs, g.ExpectedErrors)

		if g.ExpectedDetail != "" {
			found := false
			for _, err := range allErrs {
				if err.Detail == g.ExpectedDetail {
					found = true
				}
			}
			if !found {
				for _, err := range allErrs {
					t.Logf("found detail: %q", err.Detail)
				}

				t.Errorf("did not find expected error %q", g.ExpectedDetail)
			}
		}
	}
}
