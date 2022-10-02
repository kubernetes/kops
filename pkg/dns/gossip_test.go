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

package dns

import (
	"testing"
)

func TestIsGossipHostname(t *testing.T) {
	tests := []struct {
		clusterName string
		expected    bool
	}{
		{
			clusterName: "mycluster.k8s.local",
			expected:    true,
		},
		{
			clusterName: "mycluster.k8s.io",
			expected:    false,
		},
		{
			clusterName: "mycluster.k8s.local.",
			expected:    true,
		},
		{
			clusterName: "k8s.local",
			expected:    false,
		},
	}

	for _, test := range tests {
		result := IsGossipClusterName(test.clusterName)
		if result != test.expected {
			t.Errorf("Actual result %v, expected %v", result, test.expected)
		}
	}
}
