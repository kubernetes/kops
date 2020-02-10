/*
Copyright 2020 The Kubernetes Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/kops/pkg/apis/kops"
)

func TestSetInstanceGroupFields(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		fields      []string
		input       *kops.InstanceGroup
		output      *kops.InstanceGroup
	}{
		{
			name:        "one field",
			expectError: false,
			fields: []string{
				"spec.image=test-image-2",
			},
			input: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image: "test-image-1",
				},
			},
			output: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image: "test-image-2",
				},
			},
		},
		{
			name:        "multiple fields",
			expectError: false,
			fields: []string{
				"spec.image=test-image-2",
				"spec.machineType=t3.large",
				"spec.minSize=1",
				"spec.maxSize=10",
				"spec.associatePublicIp=true",
			},
			input: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image: "test-image-1",
				},
			},
			output: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image:             "test-image-2",
					MachineType:       "t3.large",
					MinSize:           newInt32Pointer(int32(1)),
					MaxSize:           newInt32Pointer(int32(10)),
					AssociatePublicIP: newBooleanPointer(true),
				},
			},
		},
		{
			name:        "invalid value",
			expectError: true,
			fields: []string{
				"spec.image=test-image-2",
				"spec.minSize=notAnInt",
			},
			input: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image: "test-image-1",
				},
			},
			output: &kops.InstanceGroup{},
		},
		{
			name:        "invalid field",
			expectError: true,
			fields: []string{
				"spec.image=test-image-2",
				"spec.invalid1.invalid2=value1",
			},
			input: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Image: "test-image-1",
				},
			},
			output: &kops.InstanceGroup{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SetInstancegroupFields(test.fields, test.input)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.output, test.input)
			}
		})
	}
}

func newInt32Pointer(i int32) *int32 { return &i }
func newBooleanPointer(b bool) *bool { return &b }
