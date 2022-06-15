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

package cloudup

import (
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
)

func Test_FindCNIAssetFromEnvironmentVariable(t *testing.T) {
	desiredCNIVersion := "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-TEST-VERSION.tar.gz"
	desiredCNIVersionHash := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	t.Setenv(ENV_VAR_CNI_ASSET_URL, desiredCNIVersion)
	t.Setenv(ENV_VAR_CNI_ASSET_HASH, desiredCNIVersionHash)

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	assetBuilder := assets.NewAssetBuilder(cluster, false)
	cniAsset, cniAssetHash, err := findCNIAssets(cluster, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Errorf("Unable to parse CNI version %s", err)
	}

	if cniAsset.String() != desiredCNIVersion {
		t.Errorf("Expected CNI version from env var %q, but got %q instead", desiredCNIVersion, cniAsset)
	}

	if cniAssetHash.String() != desiredCNIVersionHash {
		t.Errorf("Expected empty CNI version hash, but got %v instead", cniAssetHash)
	}
}

func Test_FindCNIAssetFromDefaults118(t *testing.T) {
	desiredCNIVersionURL := "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.7/cni-plugins-linux-amd64-v0.8.7.tgz"
	desiredCNIVersionHash := "sha256:977824932d5667c7a37aa6a3cbba40100a6873e7bd97e83e8be837e3e7afd0a8"

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	assetBuilder := assets.NewAssetBuilder(cluster, false)
	cniAsset, cniAssetHash, err := findCNIAssets(cluster, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Errorf("Unable to parse CNI version %s", err)
	}

	if cniAsset.String() != desiredCNIVersionURL {
		t.Errorf("Expected default CNI version %q, but got %q instead", desiredCNIVersionURL, cniAsset)
	}

	if cniAssetHash.String() != desiredCNIVersionHash {
		t.Errorf("Expected default CNI version hash %q, but got %q instead", desiredCNIVersionHash, cniAssetHash)
	}
}

func Test_FindCNIAssetFromDefaults122(t *testing.T) {
	desiredCNIVersionURL := "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz"
	desiredCNIVersionHash := "sha256:962100bbc4baeaaa5748cdbfce941f756b1531c2eadb290129401498bfac21e7"

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.22.0"

	assetBuilder := assets.NewAssetBuilder(cluster, false)
	cniAsset, cniAssetHash, err := findCNIAssets(cluster, assetBuilder, architectures.ArchitectureAmd64)
	if err != nil {
		t.Errorf("Unable to parse CNI version %s", err)
	}

	if cniAsset.String() != desiredCNIVersionURL {
		t.Errorf("Expected default CNI version %q, but got %q instead", desiredCNIVersionURL, cniAsset)
	}

	if cniAssetHash.String() != desiredCNIVersionHash {
		t.Errorf("Expected default CNI version hash %q, but got %q instead", desiredCNIVersionHash, cniAssetHash)
	}
}
