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

package util

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNodeRole(t *testing.T) {
	tests := []struct {
		name     string
		node     *v1.Node
		expected string
	}{
		{
			name: "RoleNone",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
			},
			expected: "",
		},
		{
			name: "RoleNewerLabel",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
					Labels: map[string]string{
						"node-role.kubernetes.io/node": "node-role",
					},
				},
			},
			expected: "node",
		},
		{
			name: "RoleOlderLabel",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-3",
					Labels: map[string]string{
						"kubernetes.io/role": "master",
					},
				},
			},
			expected: "master",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			role := GetNodeRole(test.node)
			if role != test.expected {
				t.Fatalf("Got role \"%s\", expected \"%s\"", role, test.expected)
			}
		})
	}
}
