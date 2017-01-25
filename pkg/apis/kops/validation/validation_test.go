/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/validation"
	"k8s.io/kubernetes/pkg/util/validation/field"
	"testing"
)

func Test_Validate_DNS(t *testing.T) {
	for _, name := range []string{"test.-", "!", "-"} {
		errs := validation.IsDNS1123Subdomain(name)
		if len(errs) == 0 {
			t.Fatalf("Expected errors validating name %q", name)
		}
	}
}

func TestValidateSubnets(t *testing.T) {
	grid := []struct {
		Input          []kops.ClusterSubnetSpec
		ExpectedErrors []string
	}{
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a"},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: ""},
			},
			ExpectedErrors: []string{"Required value::Subnets[0].Name"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a"},
				{Name: "a"},
			},
			ExpectedErrors: []string{"Invalid value::Subnets"},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ProviderID: "a"},
				{Name: "b", ProviderID: "b"},
			},
		},
		{
			Input: []kops.ClusterSubnetSpec{
				{Name: "a", ProviderID: "a"},
				{Name: "b", ProviderID: ""},
			},
			ExpectedErrors: []string{"Invalid value::Subnets"},
		},
	}
	for _, g := range grid {
		errs := validateSubnets(g.Input, field.NewPath("Subnets"))

		if len(g.ExpectedErrors) == 0 {
			if len(errs) != 0 {
				t.Errorf("unexpected errors from %q: %v", g.Input, errs)
			}
		} else {
			errStrings := sets.NewString()
			for _, err := range errs {
				errStrings.Insert(err.Type.String() + "::" + err.Field)
			}

			for _, expected := range g.ExpectedErrors {
				if !errStrings.Has(expected) {
					t.Errorf("expected error %v from %q, was not found in %q", expected, g.Input, errStrings.List())
				}
			}
		}
	}
}

func Test_Validate_DockerConfig_Storage(t *testing.T) {
	for _, name := range []string{"aufs", "zfs", "overlay"} {
		config := &kops.DockerConfig{Storage: &name}
		errs := ValidateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 0 {
			t.Fatalf("Unexpected errors validating DockerConfig %q", errs)
		}
	}

	for _, name := range []string{"overlayfs", "", "au"} {
		config := &kops.DockerConfig{Storage: &name}
		errs := ValidateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 1 {
			t.Fatalf("Expected errors validating DockerConfig %q", config)
		}
		if errs[0].Field != "docker.storage" || errs[0].Type != field.ErrorTypeNotSupported {
			t.Fatalf("Not the expected error validating DockerConfig %q", errs)
		}
	}
}
