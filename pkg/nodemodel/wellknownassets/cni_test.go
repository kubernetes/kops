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

package wellknownassets

import (
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	kopsmodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_FindCNIAssetFromEnvironmentVariable(t *testing.T) {
	desiredCNIVersion := "https://dl.k8s.io/network-plugins/cni-TEST-VERSION.tar.gz"
	desiredCNIVersionHash := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	t.Setenv(ENV_VAR_CNI_ASSET_URL, desiredCNIVersion)
	t.Setenv(ENV_VAR_CNI_ASSET_HASH, desiredCNIVersionHash)

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	ig := &api.InstanceGroup{}

	assetBuilder := assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, false)

	igModel, err := kopsmodel.ForInstanceGroup(cluster, ig)
	if err != nil {
		t.Fatalf("building instance group model: %v", err)
	}

	asset, err := FindCNIAssets(igModel, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Fatalf("Unable to parse CNI version: %v", err)
	}

	if asset.DownloadURL.String() != desiredCNIVersion {
		t.Errorf("Expected CNI version from env var %q, but got %q instead", desiredCNIVersion, asset.DownloadURL.String())
	}

	if asset.SHAValue.String() != desiredCNIVersionHash {
		t.Errorf("Expected empty CNI version hash, but got %v instead", asset.SHAValue.String())
	}
}

func TestFindCNIAssetsUsesSelectedVersionInOCIReference(t *testing.T) {
	t.Setenv(ENV_VAR_CNI_ASSET_URL, "")
	t.Setenv(ENV_VAR_CNI_ASSET_HASH, "")
	repository := "oci://registry.example.com/prefix"
	cluster := &api.Cluster{Spec: api.ClusterSpec{
		KubernetesVersion: "v1.36.0",
		Assets:            &api.AssetsSpec{FileRepository: &repository},
	}}
	igModel, err := kopsmodel.ForInstanceGroup(cluster, &api.InstanceGroup{})
	if err != nil {
		t.Fatalf("building instance group model: %v", err)
	}
	asset, err := FindCNIAssets(igModel, assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, false), architectures.ArchitectureAmd64)
	if err != nil {
		t.Fatalf("FindCNIAssets() error = %v", err)
	}
	want := "oci://registry.example.com/prefix/cni-plugins:v1.9.1-amd64"
	if got := asset.DownloadURL.String(); got != want {
		t.Fatalf("DownloadURL = %q, want %q", got, want)
	}
}

func Test_FindCNIAssetFromDefaults134(t *testing.T) {
	desiredCNIVersionURL := "https://github.com/containernetworking/plugins/releases/download/v1.7.1/cni-plugins-linux-amd64-v1.7.1.tgz"
	desiredCNIVersionHash := "sha256:1a28a0506bfe5bcdc981caf1a49eeab7e72da8321f1119b7be85f22621013098"

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.34.0"

	ig := &api.InstanceGroup{}

	igModel, err := kopsmodel.ForInstanceGroup(cluster, ig)
	if err != nil {
		t.Fatalf("building instance group model: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(vfs.Context, cluster.Spec.Assets, false)

	asset, err := FindCNIAssets(igModel, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Fatalf("Unable to parse CNI version: %s", err)
	}

	if asset.DownloadURL.String() != desiredCNIVersionURL {
		t.Errorf("Expected default CNI version %q, but got %q instead", desiredCNIVersionURL, asset.DownloadURL)
	}

	if asset.SHAValue.String() != desiredCNIVersionHash {
		t.Errorf("Expected default CNI version hash %q, but got %q instead", desiredCNIVersionHash, asset.SHAValue.String())
	}
}
