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
	"os"
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
)

func Test_FindCNIAssetFromEnvironmentVariable(t *testing.T) {

	desiredCNIVersion := "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-TEST-VERSION.tar.gz"
	desiredCNIVersionHash := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	os.Setenv(ENV_VAR_CNI_ASSET_URL, desiredCNIVersion)
	os.Setenv(ENV_VAR_CNI_ASSET_HASH, desiredCNIVersionHash)
	defer func() {
		os.Unsetenv(ENV_VAR_CNI_ASSET_URL)
		os.Unsetenv(ENV_VAR_CNI_ASSET_HASH)
	}()

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	assetBuilder := assets.NewAssetBuilder(cluster, "")
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

func Test_FindCNIAssetFromDefaults(t *testing.T) {

	desiredCNIVersion := "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-amd64-v0.8.6.tgz"
	desiredCNIVersionHash := "sha256:994fbfcdbb2eedcfa87e48d8edb9bb365f4e2747a7e47658482556c12fd9b2f5"

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	cniAsset, cniAssetHash, err := findCNIAssets(cluster, assetBuilder, architectures.ArchitectureAmd64)

	if err != nil {
		t.Errorf("Unable to parse CNI version %s", err)
	}

	if cniAsset.String() != desiredCNIVersion {
		t.Errorf("Expected default CNI version %q, but got %q instead", desiredCNIVersion, cniAsset)
	}

	if cniAssetHash.String() != desiredCNIVersionHash {
		t.Errorf("Expected default CNI version hash %q, but got %q instead", desiredCNIVersionHash, cniAssetHash)
	}
}

func Test_FindLyftAssetFromEnvironmentVariable(t *testing.T) {

	desiredLyftVersion := "https://github.com/lyft/cni-ipvlan-vpc-k8s/releases/download/TEST-VERSION/cni-TEST-VERSION.tar.gz"
	desiredLyftVersionHash := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	os.Setenv(ENV_VAR_LYFT_VPC_ASSET_URL, desiredLyftVersion)
	os.Setenv(ENV_VAR_LYFT_VPC_ASSET_HASH, desiredLyftVersionHash)
	defer func() {
		os.Unsetenv(ENV_VAR_LYFT_VPC_ASSET_URL)
		os.Unsetenv(ENV_VAR_LYFT_VPC_ASSET_HASH)
	}()

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	lyftAsset, lyftAssetHash, err := findLyftVPCAssets(cluster, assetBuilder, architectures.ArchitectureAmd64)

	if err != nil {
		t.Errorf("Unable to parse Lyft version %s", err)
	}

	if lyftAsset.String() != desiredLyftVersion {
		t.Errorf("Expected Lyft version from env var %q, but got %q instead", desiredLyftVersion, lyftAsset)
	}

	if lyftAssetHash.String() != desiredLyftVersionHash {
		t.Errorf("Expected Lyft version hash from env var %q, but got %q instead", desiredLyftVersionHash, lyftAssetHash)
	}
}

func Test_FindLyftAssetFromDefaults(t *testing.T) {

	desiredLyftVersion := "https://github.com/lyft/cni-ipvlan-vpc-k8s/releases/download/v0.6.0/cni-ipvlan-vpc-k8s-amd64-v0.6.0.tar.gz"
	desiredLyftVersionHash := "sha256:871757d381035f64020a523e7a3e139b6177b98eb7a61b547813ff25957fc566"

	cluster := &api.Cluster{}
	cluster.Spec.KubernetesVersion = "v1.18.0"

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	lyftAsset, lyftAssetHash, err := findLyftVPCAssets(cluster, assetBuilder, architectures.ArchitectureAmd64)

	if err != nil {
		t.Errorf("Unable to parse Lyft version %s", err)
	}

	if lyftAsset.String() != desiredLyftVersion {
		t.Errorf("Expected default Lyft version %q, but got %q instead", desiredLyftVersion, lyftAsset)
	}

	if lyftAssetHash.String() != desiredLyftVersionHash {
		t.Errorf("Expected default Lyft version hash %q, but got %q instead", desiredLyftVersionHash, lyftAssetHash)
	}
}
