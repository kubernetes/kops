/*
Copyright 2019 The Kubernetes Authors.

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

package etcdmanager

import (
	"fmt"
	"testing"

	"github.com/blang/semver/v4"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/util/pkg/vfs"
)

func TestBuildOptionsEtcdVersionCheck(t *testing.T) {
	latest36 := components.LatestEtcd36Version
	sv := semver.MustParse(latest36)
	next36 := fmt.Sprintf("%d.%d.%d", sv.Major, sv.Minor, sv.Patch+1)

	tests := []struct {
		name        string
		version     string
		image       string
		expectError bool
	}{
		{
			name:    "bundled version",
			version: latest36,
		},
		{
			name:        "version not bundled",
			version:     next36,
			expectError: true,
		},
		{
			name:    "version not bundled, with custom image",
			version: next36,
			image:   "gcr.io/etcd-development/etcd:v" + next36,
		},
		{
			name:    "bundled version, with custom image",
			version: latest36,
			image:   "gcr.io/etcd-development/etcd:v" + latest36,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.34.0",
					ConfigStore: kops.ConfigStoreSpec{
						Base: "memfs://clusters.example.com/minimal.example.com",
					},
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Name:    "main",
							Version: test.version,
							Image:   test.image,
						},
					},
				},
			}

			assetBuilder := assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, false)
			optionsContext, err := components.NewOptionsContext(cluster, assetBuilder, assetBuilder.KubeletSupportedVersion)
			if err != nil {
				t.Fatalf("unexpected error from NewOptionsContext: %v", err)
			}

			builder := &EtcdManagerOptionsBuilder{OptionsContext: optionsContext}
			err = builder.BuildOptions(cluster)
			if test.expectError && err == nil {
				t.Errorf("expected error from BuildOptions, got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("unexpected error from BuildOptions: %v", err)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		base        string
		other1      string
		other2      string
		expectedStr string
	}{
		{
			base:        "/test",
			other1:      "z1",
			other2:      "/z2",
			expectedStr: "/test/z1/z2",
		},
		{
			base:        "test/",
			other1:      "z1",
			other2:      "/z2",
			expectedStr: "test/z1/z2",
		},
	}
	for _, test := range tests {
		result := join(test.base, test.other1, test.other2)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}
