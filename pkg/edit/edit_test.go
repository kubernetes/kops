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

	"github.com/MakeNowJust/heredoc"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
)

var testTimestamp = metav1.Time{Time: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)}
var testObj = v1alpha2.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		CreationTimestamp: testTimestamp,
		Name:              "hello",
	},
	Spec: v1alpha2.ClusterSpec{
		KubernetesVersion: "1.2.3",
	},
}

func TestHasExtraFields(t *testing.T) {
	tests := []struct {
		obj      runtime.Object
		yaml     string
		expected string
	}{
		{
			obj: &testObj,
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
			obj: &testObj,
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
