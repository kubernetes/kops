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

package wellknownassets

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	kopsmodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_FindGVisorAsset(t *testing.T) {
	tests := []struct {
		arch     architectures.Architecture
		wantURL  string
		wantHash string
	}{
		{
			arch:     architectures.ArchitectureAmd64,
			wantURL:  "https://storage.googleapis.com/gvisor/releases/release/20261109.0/x86_64/gvisor.tar.zstd",
			wantHash: "sha512:0a0041fd862f4ea3042a9db8d41c742f48a01f909699adc2a0256e46d85e53f97b82b9fd154b3dce636f62b8d557677639f91c77d13f3f875ff9fa535275393c",
		},
		{
			arch:     architectures.ArchitectureArm64,
			wantURL:  "https://storage.googleapis.com/gvisor/releases/release/20261109.0/aarch64/gvisor.tar.zstd",
			wantHash: "sha512:a0b5d911512ef5fe70e833821e653a90c4d06c7027246ee49423f1d2fe2b2a509389a637797248958b90e49a069bb079ecde860efe1871b1d980ec0fe17e18ac",
		},
	}

	cluster := &kops.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.36.0"
	igModel, err := kopsmodel.ForInstanceGroup(cluster, &kops.InstanceGroup{})
	if err != nil {
		t.Fatalf("building instance group model: %v", err)
	}
	assetBuilder := assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, false)

	for _, test := range tests {
		t.Run(string(test.arch), func(t *testing.T) {
			asset, err := FindGVisorAsset(igModel, assetBuilder, test.arch)
			if err != nil {
				t.Fatalf("FindGVisorAsset: %v", err)
			}
			if asset.DownloadURL.String() != test.wantURL {
				t.Errorf("got URL %q, want %q", asset.DownloadURL, test.wantURL)
			}
			if asset.SHAValue.String() != test.wantHash {
				t.Errorf("got hash %q, want %q", asset.SHAValue, test.wantHash)
			}
		})
	}
}
