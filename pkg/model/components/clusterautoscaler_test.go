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

package components

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/vfs"
)

func TestClusterAutoscalerImage(t *testing.T) {
	grid := []struct {
		kubernetesVersion string
		image             *string
		expected          string
	}{
		{
			kubernetesVersion: "v1.32.0",
			expected:          "registry.k8s.io/autoscaling/cluster-autoscaler:v1.34.6",
		},
		{
			kubernetesVersion: "v1.33.5",
			expected:          "registry.k8s.io/autoscaling/cluster-autoscaler:v1.34.6",
		},
		{
			kubernetesVersion: "v1.34.0",
			expected:          "registry.k8s.io/autoscaling/cluster-autoscaler:v1.37.0",
		},
		{
			kubernetesVersion: "v1.37.0",
			expected:          "registry.k8s.io/autoscaling/cluster-autoscaler:v1.37.0",
		},
		{
			kubernetesVersion: "v1.38.0-alpha.1",
			expected:          "registry.k8s.io/autoscaling/cluster-autoscaler:v1.37.0",
		},
		{
			kubernetesVersion: "v1.32.0",
			image:             new("example.com/cluster-autoscaler:custom"),
			expected:          "example.com/cluster-autoscaler:custom",
		},
	}

	for _, g := range grid {
		t.Run(g.kubernetesVersion, func(t *testing.T) {
			c := buildCluster()
			c.Spec.KubernetesVersion = g.kubernetesVersion
			c.Spec.ClusterAutoscaler = &kops.ClusterAutoscalerConfig{
				Enabled: new(true),
				Image:   g.image,
			}
			b := assets.NewAssetBuilder(vfs.Context, c.Spec.Assets, false)

			optionsContext, err := NewOptionsContext(c, b, b.KubeletSupportedVersion)
			if err != nil {
				t.Fatalf("error from NewOptionsContext: %v", err)
			}

			cas := &ClusterAutoscalerOptionsBuilder{OptionsContext: optionsContext}
			if err := cas.BuildOptions(c); err != nil {
				t.Fatalf("unexpected error from BuildOptions: %v", err)
			}

			actual := c.Spec.ClusterAutoscaler.Image
			if actual == nil || *actual != g.expected {
				t.Errorf("expected image %q, got %v", g.expected, actual)
			}
		})
	}
}
