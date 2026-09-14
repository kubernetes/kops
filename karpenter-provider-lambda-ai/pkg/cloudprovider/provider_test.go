/*
Copyright The Kubernetes Authors.

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

package cloudprovider

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

func TestResolveRegion(t *testing.T) {
	tests := []struct {
		name         string
		requirements scheduling.Requirements
		want         string
	}{
		{
			name:         "defaults to us-west-1",
			requirements: scheduling.NewRequirements(),
			want:         "us-west-1",
		},
		{
			name: "uses topology requirement",
			requirements: scheduling.NewRequirements(
				scheduling.NewRequirement(v1.LabelTopologyZone, v1.NodeSelectorOpIn, "us-east-1"),
			),
			want: "us-east-1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveRegion(test.requirements); got != test.want {
				t.Errorf("resolveRegion() = %q, want %q", got, test.want)
			}
		})
	}
}
