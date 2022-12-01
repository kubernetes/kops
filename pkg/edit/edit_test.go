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

package edit

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
)

func TestHasExtraFields(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name: "minimal",
			yaml: heredoc.Doc(`
			apiVersion: kops.k8s.io/v1alpha2
			kind: Cluster
			metadata:
			  creationTimestamp: "2017-01-01T00:00:00Z"
			  name: hello
			spec:
			  kubernetesVersion: 1.2.3
			`),
			expected: "",
		},

		{
			name: "extraFields",
			yaml: heredoc.Doc(`
			apiVersion: kops.k8s.io/v1alpha2
			kind: Cluster
			metadata:
			  creationTimestamp: "2017-01-01T00:00:00Z"
			  name: hello
			extraFields: true
			spec:
			  kubernetesVersion: 1.2.3
			`),
			expected: heredoc.Doc(`
			  apiVersion: kops.k8s.io/v1alpha2
			+ extraFields: true
			  kind: Cluster
			  metadata:
			...
			`),
		},
		{
			name: "spec2",
			yaml: heredoc.Doc(`
			apiVersion: kops.k8s.io/v1alpha2
			kind: Cluster
			metadata:
			  creationTimestamp: "2017-01-01T00:00:00Z"
			  name: hello
			spec2:
			  kubernetesVersion: 1.2.3
			`),
			expected: heredoc.Doc(`...
			    creationTimestamp: "2017-01-01T00:00:00Z"
			    name: hello
			+ spec2:
			+   kubernetesVersion: 1.2.3
			- spec: {}
			`),
		},
		{
			name: "isolateMasters",
			yaml: heredoc.Doc(`
			apiVersion: kops.k8s.io/v1alpha2
			kind: Cluster
			metadata:
			  creationTimestamp: "2017-01-01T00:00:00Z"
			  name: hello
			spec:
			  kubernetesVersion: 1.2.3
			  isolateMasters: false
			`),
			expected: "",
		},
		{
			name: "instanceGroup",
			yaml: heredoc.Doc(`
			apiVersion: kops.k8s.io/v1alpha2
			kind: InstanceGroup
			metadata:
			  creationTimestamp: "2017-01-01T00:00:00Z"
			  name: hello
			spec:
			  role: Node
			`),
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := HasExtraFields(test.yaml)
			if err != nil {
				t.Errorf("Error from HasExtraFields: %v", err)
				return
			}
			if result != test.expected {
				t.Errorf("Actual result:\n %s \nExpect:\n %s", result, test.expected)
			}
		})
	}
}
