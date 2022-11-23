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

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

// TODO: we really need to sort this out:
// https://github.com/kubernetes/kops/issues/724
// https://github.com/kubernetes/kops/issues/626
// https://github.com/kubernetes/kubernetes/issues/30338

const (
	// defaultCNIAssetAmd64K8s_15 is the CNI tarball for k8s >= 1.15
	defaultCNIAssetAmd64K8s_15 = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.7/cni-plugins-linux-amd64-v0.8.7.tgz"
	defaultCNIAssetArm64K8s_15 = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.7/cni-plugins-linux-arm64-v0.8.7.tgz"
	// defaultCNIAssetAmd64K8s_22 is the CNI tarball for k8s >= 1.22
	defaultCNIAssetAmd64K8s_22 = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz"
	defaultCNIAssetArm64K8s_22 = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.9.1/cni-plugins-linux-arm64-v0.9.1.tgz"

	// Environment variable for overriding CNI url
	ENV_VAR_CNI_ASSET_URL  = "CNI_VERSION_URL"
	ENV_VAR_CNI_ASSET_HASH = "CNI_ASSET_HASH_STRING"
)

func findCNIAssets(c *kopsapi.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	// Override CNI packages from env vars
	cniAssetURL := viper.GetString(ENV_VAR_CNI_ASSET_URL)
	cniAssetHash := viper.GetString(ENV_VAR_CNI_ASSET_HASH)

	if cniAssetURL != "" && cniAssetHash != "" {
		klog.V(2).Infof("Using CNI asset URL %q, as set in %s", cniAssetURL, ENV_VAR_CNI_ASSET_URL)
		klog.V(2).Infof("Using CNI asset hash %q, as set in %s", cniAssetHash, ENV_VAR_CNI_ASSET_HASH)

		u, err := url.Parse(cniAssetURL)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse CNI plugin binaries asset URL %q: %v", cniAssetURL, err)
		}

		h, err := hashing.FromString(cniAssetHash)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse CNI plugin binaries asset hash %q: %v", cniAssetHash, err)
		}

		u, err = assetBuilder.RemapFileAndSHAValue(u, cniAssetHash)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to remap CNI plugin binaries asset: %v", err)
		}

		return u, h, nil
	}

	switch arch {
	case architectures.ArchitectureAmd64:
		if c.IsKubernetesLT("1.22") {
			cniAssetURL = defaultCNIAssetAmd64K8s_15
		} else {
			cniAssetURL = defaultCNIAssetAmd64K8s_22
		}
		klog.V(2).Infof("Adding default ARM64 CNI plugin binaries asset: %s", cniAssetURL)
	case architectures.ArchitectureArm64:
		if c.IsKubernetesLT("1.22") {
			cniAssetURL = defaultCNIAssetArm64K8s_15
		} else {
			cniAssetURL = defaultCNIAssetArm64K8s_22
		}
		klog.V(2).Infof("Adding default AMD64 CNI plugin binaries asset: %s", cniAssetURL)
	default:
		return nil, nil, fmt.Errorf("unknown arch for CNI plugin binaries asset: %s", arch)
	}

	u, err := url.Parse(cniAssetURL)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse CNI plugin binaries asset URL %q: %v", cniAssetURL, err)
	}

	u, h, err := assetBuilder.RemapFileAndSHA(u)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to remap CNI plugin binaries asset: %v", err)
	}

	return u, h, nil
}
