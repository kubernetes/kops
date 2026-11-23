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

package model

import (
	"path"
	"slices"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

func TestGVisorBuilderBuild(t *testing.T) {
	// The files of the gVisor release tarball, as the asset store lists them.
	gVisorAssets := map[string]string{
		"runsc":                     "runsc",
		"containerd-shim-runsc-v1":  "containerd-shim-runsc-v1",
		"gvisor-bin/gvisor_sentry":  "gvisor_sentry",
		"gvisor-bin/runsc-netgofer": "runsc-netgofer",
	}

	tests := []struct {
		name         string
		distribution distributions.Distribution
		role         kops.InstanceGroupRole
		gvisor       *kops.GVisorConfig
		assets       map[string]string
		wantFiles    []string
		wantErr      bool
	}{
		{
			name:         "disabled",
			distribution: distributions.DistributionDebian13,
			role:         kops.InstanceGroupRoleNode,
			gvisor:       &kops.GVisorConfig{Enabled: new(false)},
			assets:       gVisorAssets,
		},
		{
			name:         "enabled debian",
			distribution: distributions.DistributionDebian13,
			role:         kops.InstanceGroupRoleNode,
			gvisor:       &kops.GVisorConfig{Enabled: new(true)},
			assets:       gVisorAssets,
			wantFiles: []string{
				"/usr/local/bin/containerd-shim-runsc-v1",
				"/usr/local/bin/gvisor-bin",
				"/usr/local/bin/gvisor-bin/gvisor_sentry",
				"/usr/local/bin/gvisor-bin/runsc-netgofer",
				"/usr/local/bin/runsc",
			},
		},
		{
			name:         "enabled rhel",
			distribution: distributions.DistributionRhel9,
			role:         kops.InstanceGroupRoleNode,
			gvisor:       &kops.GVisorConfig{Enabled: new(true)},
			assets:       gVisorAssets,
			wantFiles: []string{
				"/usr/local/bin/containerd-shim-runsc-v1",
				"/usr/local/bin/gvisor-bin",
				"/usr/local/bin/gvisor-bin/gvisor_sentry",
				"/usr/local/bin/gvisor-bin/runsc-netgofer",
				"/usr/local/bin/runsc",
			},
		},
		{
			name:         "enabled flatcar",
			distribution: distributions.DistributionFlatcar,
			role:         kops.InstanceGroupRoleNode,
			gvisor:       &kops.GVisorConfig{Enabled: new(true)},
			assets:       gVisorAssets,
			wantFiles: []string{
				"/opt/kops/bin/containerd-shim-runsc-v1",
				"/opt/kops/bin/gvisor-bin",
				"/opt/kops/bin/gvisor-bin/gvisor_sentry",
				"/opt/kops/bin/gvisor-bin/runsc-netgofer",
				"/opt/kops/bin/runsc",
			},
		},
		{
			name:         "enabled without the gVisor asset",
			distribution: distributions.DistributionFlatcar,
			role:         kops.InstanceGroupRoleNode,
			gvisor:       &kops.GVisorConfig{Enabled: new(true)},
			wantErr:      true,
		},
		{
			name:         "enabled without sidecar binaries",
			distribution: distributions.DistributionFlatcar,
			role:         kops.InstanceGroupRoleNode,
			gvisor:       &kops.GVisorConfig{Enabled: new(true)},
			assets: map[string]string{
				"runsc":                    "runsc",
				"containerd-shim-runsc-v1": "containerd-shim-runsc-v1",
			},
			wantErr: true,
		},
		{
			name:         "enabled control plane",
			distribution: distributions.DistributionDebian13,
			role:         kops.InstanceGroupRoleControlPlane,
			gvisor:       &kops.GVisorConfig{Enabled: new(true)},
			assets:       gVisorAssets,
		},
		{
			name:         "unset",
			distribution: distributions.DistributionDebian13,
			role:         kops.InstanceGroupRoleNode,
			assets:       gVisorAssets,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assetStore := fi.NewAssetStore(t.TempDir())
			for assetPath, contents := range test.assets {
				assetStore.AddForTest(path.Base(assetPath), assetPath, contents)
			}

			context := &fi.NodeupModelBuilderContext{
				Tasks: make(map[string]fi.NodeupTask),
			}
			builder := &GVisorBuilder{
				NodeupModelContext: &NodeupModelContext{
					Assets:       assetStore,
					Distribution: test.distribution,
					BootConfig: &nodeup.BootConfig{
						InstanceGroupRole: test.role,
					},
					NodeupConfig: &nodeup.Config{
						GVisor: test.gvisor,
					},
				},
			}

			err := builder.Build(context)
			if test.wantErr {
				if err == nil {
					t.Fatalf("Build returned no error, want one")
				}
				return
			}
			if err != nil {
				t.Fatalf("Build returned error: %v", err)
			}

			var files []string
			for _, task := range context.Tasks {
				file, ok := task.(*nodetasks.File)
				if !ok {
					t.Fatalf("unexpected task %v", task)
				}
				if fi.ValueOf(file.Mode) != "0755" {
					t.Errorf("%s has mode %q, want 0755", file.Path, fi.ValueOf(file.Mode))
				}
				files = append(files, file.Path)
			}
			slices.Sort(files)
			if !slices.Equal(files, test.wantFiles) {
				t.Errorf("got files %v, want %v", files, test.wantFiles)
			}
		})
	}
}
