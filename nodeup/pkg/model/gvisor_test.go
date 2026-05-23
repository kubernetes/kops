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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestGVisorBuilderBuild(t *testing.T) {
	tests := []struct {
		name         string
		distribution distributions.Distribution
		gvisor       *kops.GVisorConfig
		wantTasks    []string
	}{
		{
			name:         "disabled",
			distribution: distributions.DistributionDebian13,
			gvisor:       &kops.GVisorConfig{Enabled: fi.PtrTo(false)},
		},
		{
			name:         "enabled debian",
			distribution: distributions.DistributionDebian13,
			gvisor:       &kops.GVisorConfig{Enabled: fi.PtrTo(true)},
			wantTasks:    []string{"AptSource/gvisor", "Package/runsc"},
		},
		{
			name:         "enabled non debian",
			distribution: distributions.DistributionRhel9,
			gvisor:       &kops.GVisorConfig{Enabled: fi.PtrTo(true)},
		},
		{
			name:         "unset",
			distribution: distributions.DistributionDebian13,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := &fi.NodeupModelBuilderContext{
				Tasks: make(map[string]fi.NodeupTask),
			}
			builder := &GVisorBuilder{
				NodeupModelContext: &NodeupModelContext{
					Distribution: test.distribution,
					NodeupConfig: &nodeup.Config{
						GVisor: test.gvisor,
					},
				},
			}

			if err := builder.Build(context); err != nil {
				t.Fatalf("Build returned error: %v", err)
			}
			if len(context.Tasks) != len(test.wantTasks) {
				t.Fatalf("got %d tasks, want %d: %v", len(context.Tasks), len(test.wantTasks), context.Tasks)
			}
			for _, key := range test.wantTasks {
				if _, ok := context.Tasks[key]; !ok {
					t.Errorf("missing task %q", key)
				}
			}
		})
	}
}
