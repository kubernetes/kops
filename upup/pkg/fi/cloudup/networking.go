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
	"fmt"
	"net/url"
	"os"

	"k8s.io/klog"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

// TODO: we really need to sort this out:
// https://github.com/kubernetes/kops/issues/724
// https://github.com/kubernetes/kops/issues/626
// https://github.com/kubernetes/kubernetes/issues/30338

const (
	// defaultCNIAssetAmd64K8s_11 is the CNI tarball for k8s >= 1.11
	defaultCNIAssetAmd64K8s_11              = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-amd64-v0.7.5.tgz"
	defaultCNIAssetAmd64SHA256StringK8s1_11 = "3ca15c0a18ee830520cf3a95408be826cbd255a1535a38e0be9608b25ad8bf64"
	defaultCNIAssetArm64K8s_11              = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-arm64-v0.7.5.tgz"
	defaultCNIAssetArm64SHA256StringK8s1_11 = "7fec91af78e9548df306f0ec43bea527c8c10cc3a9682c33e971c8522a7fcded"

	// defaultCNIAssetAmd64K8s_15 is the CNI tarball for k8s >= 1.15
	defaultCNIAssetAmd64K8s_15              = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-amd64-v0.8.6.tgz"
	defaultCNIAssetAmd64SHA256StringK8s1_15 = "994fbfcdbb2eedcfa87e48d8edb9bb365f4e2747a7e47658482556c12fd9b2f5"
	defaultCNIAssetArm64K8s_15              = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-arm64-v0.8.6.tgz"
	defaultCNIAssetArm64SHA256StringK8s1_15 = "43fbf750c5eccb10accffeeb092693c32b236fb25d919cf058c91a677822c999"

	// Environment variable for overriding CNI url
	ENV_VAR_CNI_VERSION_URL       = "CNI_VERSION_URL"
	ENV_VAR_CNI_ASSET_HASH_STRING = "CNI_ASSET_HASH_STRING"
)

func findCNIAssets(c *kopsapi.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {

	if cniVersionURL := os.Getenv(ENV_VAR_CNI_VERSION_URL); cniVersionURL != "" {
		u, err := url.Parse(cniVersionURL)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse %q as a URL: %v", cniVersionURL, err)
		}

		klog.Infof("Using CNI asset version %q, as set in %s", cniVersionURL, ENV_VAR_CNI_VERSION_URL)

		if cniAssetHashString := os.Getenv(ENV_VAR_CNI_ASSET_HASH_STRING); cniAssetHashString != "" {

			klog.Infof("Using CNI asset hash %q, as set in %s", cniAssetHashString, ENV_VAR_CNI_ASSET_HASH_STRING)

			hash, err := hashing.FromString(cniAssetHashString)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to parse CNI asset hash %q", cniAssetHashString)
			}
			return u, hash, nil
		}
		return u, nil, nil
	}

	sv, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup kubernetes version: %v", err)
	}

	var cniAsset, cniAssetHash string
	switch arch {
	case architectures.ArchitectureAmd64:
		if util.IsKubernetesGTE("1.15", *sv) {
			cniAsset = defaultCNIAssetAmd64K8s_15
			cniAssetHash = defaultCNIAssetAmd64SHA256StringK8s1_15
		} else {
			cniAsset = defaultCNIAssetAmd64K8s_11
			cniAssetHash = defaultCNIAssetAmd64SHA256StringK8s1_11
		}
		klog.V(2).Infof("Adding default ARM64 CNI plugin binaries asset : %s", cniAsset)
	case architectures.ArchitectureArm64:
		if util.IsKubernetesGTE("1.15", *sv) {
			cniAsset = defaultCNIAssetArm64K8s_15
			cniAssetHash = defaultCNIAssetArm64SHA256StringK8s1_15
		} else {
			cniAsset = defaultCNIAssetArm64K8s_11
			cniAssetHash = defaultCNIAssetArm64SHA256StringK8s1_11
		}
		klog.V(2).Infof("Adding default AMD64 CNI plugin binaries asset : %s", cniAsset)
	default:
		return nil, nil, fmt.Errorf("unknown arch for CNI plugin binaries asset %s", arch)
	}

	u, err := url.Parse(cniAsset)
	if err != nil {
		return nil, nil, nil
	}

	hash, err := hashing.FromString(cniAssetHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse CNI plugin binaries asset hash %q", cniAssetHash)
	}

	u, err = assetBuilder.RemapFileAndSHAValue(u, cniAssetHash)
	if err != nil {
		return nil, nil, err
	}

	return u, hash, nil
}
