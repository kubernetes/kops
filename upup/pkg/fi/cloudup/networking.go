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
	defaultCNIAssetAmd64K8s_11 = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-amd64-v0.7.5.tgz"
	defaultCNIAssetArm64K8s_11 = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-arm64-v0.7.5.tgz"

	// defaultCNIAssetAmd64K8s_15 is the CNI tarball for k8s >= 1.15
	defaultCNIAssetAmd64K8s_15 = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-amd64-v0.8.6.tgz"
	defaultCNIAssetArm64K8s_15 = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-arm64-v0.8.6.tgz"

	// Environment variable for overriding CNI url
	ENV_VAR_CNI_ASSET_URL  = "CNI_VERSION_URL"
	ENV_VAR_CNI_ASSET_HASH = "CNI_ASSET_HASH_STRING"

	// Default LyftVPC packages
	defaultLyftVPCAssetAmd64       = "https://github.com/lyft/cni-ipvlan-vpc-k8s/releases/download/v0.6.0/cni-ipvlan-vpc-k8s-amd64-v0.6.0.tar.gz"
	defaultLyftVPCAssetAmd64SHA256 = "871757d381035f64020a523e7a3e139b6177b98eb7a61b547813ff25957fc566"
	defaultLyftVPCAssetArm64       = "https://github.com/lyft/cni-ipvlan-vpc-k8s/releases/download/v0.6.0/cni-ipvlan-vpc-k8s-arm64-v0.6.0.tar.gz"
	defaultLyftVPCAssetArm64SHA256 = "3aadcb32ffda53990153790203eb72898e55a985207aa5b4451357f9862286f0"

	// Environment variable for overriding LyftVPC url
	ENV_VAR_LYFT_VPC_ASSET_URL  = "LYFT_VPC_DOWNLOAD_URL"
	ENV_VAR_LYFT_VPC_ASSET_HASH = "LYFT_VPC_DOWNLOAD_HASH"
)

func findCNIAssets(c *kopsapi.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	// Override CNI packages from env vars
	cniAssetURL := os.Getenv(ENV_VAR_CNI_ASSET_URL)
	cniAssetHash := os.Getenv(ENV_VAR_CNI_ASSET_HASH)

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

	sv, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find Kubernetes version: %v", err)
	}

	switch arch {
	case architectures.ArchitectureAmd64:
		if util.IsKubernetesGTE("1.15", *sv) {
			cniAssetURL = defaultCNIAssetAmd64K8s_15
		} else {
			cniAssetURL = defaultCNIAssetAmd64K8s_11
		}
		klog.V(2).Infof("Adding default ARM64 CNI plugin binaries asset: %s", cniAssetURL)
	case architectures.ArchitectureArm64:
		if util.IsKubernetesGTE("1.15", *sv) {
			cniAssetURL = defaultCNIAssetArm64K8s_15
		} else {
			cniAssetURL = defaultCNIAssetArm64K8s_11
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

func findLyftVPCAssets(c *kopsapi.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	// Override LyftVPC packages from env vars
	lyftAssetURL := os.Getenv(ENV_VAR_LYFT_VPC_ASSET_URL)
	lyftAssetHash := os.Getenv(ENV_VAR_LYFT_VPC_ASSET_HASH)

	if lyftAssetURL != "" && lyftAssetHash != "" {
		klog.V(2).Infof("Using LyftVPC package URL %q, as set in %s", lyftAssetURL, ENV_VAR_LYFT_VPC_ASSET_URL)
		klog.V(2).Infof("Using LyftVPC package hash %q, as set in %s", lyftAssetHash, ENV_VAR_LYFT_VPC_ASSET_HASH)
	} else {
		switch arch {
		case architectures.ArchitectureAmd64:
			lyftAssetURL = defaultLyftVPCAssetAmd64
			lyftAssetHash = defaultLyftVPCAssetAmd64SHA256
		case architectures.ArchitectureArm64:
			lyftAssetURL = defaultLyftVPCAssetArm64
			lyftAssetHash = defaultLyftVPCAssetArm64SHA256
		default:
			return nil, nil, fmt.Errorf("unknown arch for LyftVPC asset: %s", arch)
		}
	}

	u, err := url.Parse(lyftAssetURL)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse LyftVPC asset URL %q: %v", lyftAssetURL, err)
	}

	h, err := hashing.FromString(lyftAssetHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse LyftVPC asset hash %q: %v", lyftAssetHash, err)
	}

	u, err = assetBuilder.RemapFileAndSHAValue(u, lyftAssetHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to remap LyftVPC asset: %v", err)
	}

	return u, h, nil
}
