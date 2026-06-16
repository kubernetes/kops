/*
Copyright 2026 The Kubernetes Authors.

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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
)

func newLinodeClusterForClusterValidation(name string) *kops.Cluster {
	return &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Linode: &kops.LinodeSpec{},
			},
		},
	}
}

func TestLinodeValidateCluster(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *kops.Cluster
		expected []*field.Error
	}{
		{
			name:    "accepts cluster name within limit",
			cluster: newLinodeClusterForClusterValidation("linode.example.com"),
		},
		{
			name:    "rejects cluster name longer than 32 characters",
			cluster: newLinodeClusterForClusterValidation(strings.Repeat("a", 33)),
			expected: []*field.Error{
				{
					Type:   field.ErrorTypeInvalid,
					Field:  "objectMeta.name",
					Detail: "cluster name must be no more than 32 characters long",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errList := linodeValidateCluster(tt.cluster)
			testFieldErrors(t, errList, tt.expected)
		})
	}
}
