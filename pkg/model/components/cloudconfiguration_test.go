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

package components

import (
	"testing"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestCloudConfigurationOptionsBuilder(t *testing.T) {
	ob := &CloudConfigurationOptionsBuilder{
		Context: nil,
	}
	disabled := fi.Bool(false)
	enabled := fi.Bool(true)
	for _, test := range []struct {
		description              string
		generalManageSCs         *bool
		openStackManageSCs       *bool
		expectedGeneralManageSCs *bool
	}{
		{
			"neither",
			nil,
			nil,
			enabled,
		},
		{
			"all false",
			disabled,
			nil,
			disabled,
		},
		{
			"all true",
			enabled,
			nil,
			enabled,
		},
		{
			"os false",
			nil,
			disabled,
			disabled,
		},
		{
			"os true",
			nil,
			enabled,
			enabled,
		},
		{
			"all false, os false",
			disabled,
			disabled,
			disabled,
		},
		{
			"all false, os true",
			// Caught as conflict during validation.
			disabled,
			enabled,
			disabled,
		},
		{
			"all true, os false",
			// Caught as conflict during validation.
			enabled,
			disabled,
			enabled,
		},
		{
			"all true, os true",
			enabled,
			enabled,
			enabled,
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			spec := kopsapi.ClusterSpec{
				CloudProvider: kopsapi.CloudProviderSpec{
					Openstack: &kopsapi.OpenstackSpec{},
				},
				CloudConfig: &kopsapi.CloudConfiguration{
					Openstack: &kopsapi.OpenstackConfiguration{
						BlockStorage: &kopsapi.OpenstackBlockStorageConfig{},
					},
				},
			}
			if p := test.generalManageSCs; p != nil {
				spec.CloudConfig.ManageStorageClasses = p
			}
			if p := test.openStackManageSCs; p != nil {
				spec.CloudConfig.Openstack.BlockStorage.CreateStorageClass = p
			}
			if err := ob.BuildOptions(&spec); err != nil {
				t.Fatalf("failed to build options: %v", err)
			}
			if want, got := test.expectedGeneralManageSCs, spec.CloudConfig.ManageStorageClasses; (want == nil) != (got == nil) || (got != nil && *got != *want) {
				switch {
				case want == nil:
					t.Errorf("spec.cloudConfig.manageStorageClasses: want nil, got %t", *got)
				case got == nil:
					t.Errorf("spec.cloudConfig.manageStorageClasses: want %t, got nil", *want)
				default:
					t.Errorf("spec.cloudConfig.manageStorageClasses: want %t, got %t", *want, *got)
				}
			}
		})
	}
}
