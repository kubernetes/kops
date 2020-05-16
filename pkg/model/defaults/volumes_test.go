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

package defaults

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestDefaultInstanceGroupVolumeSize(t *testing.T) {
	tests := []struct {
		role     kops.InstanceGroupRole
		expected int32
	}{
		{
			role:     "Node2",
			expected: -1,
		},
	}
	for _, test := range tests {
		result, _ := DefaultInstanceGroupVolumeSize(test.role)
		if test.expected != result {
			t.Errorf("Expected %d, got %d", test.expected, result)
		}
	}
}
