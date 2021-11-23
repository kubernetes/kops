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
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/kops/pkg/apis/kops"
)

var (
	testTimestamp  = metav1.Time{Time: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)}
	testClusterObj = kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: testTimestamp,
			Name:              "hello",
		},
		Spec: kops.ClusterSpec{
			KubernetesVersion: "1.2.3",
		},
	}
	testIGObj = kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: testTimestamp,
			Name:              "hello",
		},
		Spec: kops.InstanceGroupSpec{
			Role: kops.InstanceGroupRoleNode,
		},
	}
)

func TestHasExtraFields(t *testing.T) {
	tests := []struct {
		obj      runtime.Object
		yaml     string
		expected string
	}{
		{
			obj: &testClusterObj,
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
			obj: &testClusterObj,
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
			obj: &testClusterObj,
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
			- spec:
			    kubernetesVersion: 1.2.3
			`),
		},
		{
			obj: &testClusterObj,
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
			obj: &testIGObj,
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
		result, err := HasExtraFields(test.yaml, test.obj)
		if err != nil {
			t.Errorf("Error from HasExtraFields: %v", err)
			continue
		}
		if result != test.expected {
			t.Errorf("Actual result:\n %s \nExpect:\n %s", result, test.expected)
		}
	}
}
