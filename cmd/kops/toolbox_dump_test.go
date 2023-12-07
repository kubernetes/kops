/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTruncateNodeList(t *testing.T) {
	cases := []struct {
		name     string
		input    []corev1.Node
		max      int
		expected []corev1.Node
		err      bool
	}{
		{
			name: "less than max",
			input: []corev1.Node{
				makeNode(),
				makeNode(),
				makeControlPlaneNode(),
			},
			max: 5,
			expected: []corev1.Node{
				makeControlPlaneNode(),
				makeNode(),
				makeNode(),
			},
		},
		{
			name: "truncate",
			input: []corev1.Node{
				makeNode(),
				makeNode(),
				makeNode(),
				makeControlPlaneNode(),
				makeNode(),
			},
			max: 4,
			expected: []corev1.Node{
				makeControlPlaneNode(),
				makeNode(),
				makeNode(),
				makeNode(),
			},
		},
		{
			name: "less than zero",
			input: []corev1.Node{
				makeNode(),
				makeNode(),
				makeNode(),
				makeControlPlaneNode(),
				makeNode(),
			},
			max: -1,
			err: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodeList := corev1.NodeList{Items: tc.input}
			err := truncateNodeList(&nodeList, tc.max)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, nodeList.Items)
			}
		})
	}
}

func makeControlPlaneNode() corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/control-plane": "",
			},
		},
	}
}

func makeNode() corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/node": "",
			},
		},
	}
}
