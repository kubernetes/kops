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
	"fmt"
	"net/url"
	"os"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

// TODO: we really need to sort this out:
// https://github.com/kubernetes/kops/issues/724
// https://github.com/kubernetes/kops/issues/626
// https://github.com/kubernetes/kubernetes/issues/30338

const (
	defaultCNIVersionK8s_32 = "1.6.2"
	defaultCNIVersionK8s_34 = "1.7.1"
	defaultCNIVersionK8s_35 = "1.8.0"
	defaultCNIVersionK8s_36 = "1.9.1"
	defaultCNIAssetURL      = "https://github.com/containernetworking/plugins/releases/download/v%[1]s/cni-plugins-linux-%[2]s-v%[1]s.tgz"

	// Environment variable for overriding CNI url
	ENV_VAR_CNI_ASSET_URL  = "CNI_VERSION_URL"
	ENV_VAR_CNI_ASSET_HASH = "CNI_ASSET_HASH_STRING"
)

func FindCNIAssets(ig model.InstanceGroup, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*assets.FileAsset, error) {
	// Override CNI packages from env vars
	cniAssetURL := os.Getenv(ENV_VAR_CNI_ASSET_URL)
	cniAssetHash := os.Getenv(ENV_VAR_CNI_ASSET_HASH)

	if cniAssetURL != "" && cniAssetHash != "" {
		klog.V(2).Infof("Using CNI asset URL %q, as set in %s", cniAssetURL, ENV_VAR_CNI_ASSET_URL)
		klog.V(2).Infof("Using CNI asset hash %q, as set in %s", cniAssetHash, ENV_VAR_CNI_ASSET_HASH)

		u, err := url.Parse(cniAssetURL)
		if err != nil {
			return nil, fmt.Errorf("unable to parse CNI plugin binaries asset URL %q: %v", cniAssetURL, err)
		}

		h, err := hashing.FromString(cniAssetHash)
		if err != nil {
			return nil, fmt.Errorf("unable to parse CNI plugin binaries asset hash %q: %v", cniAssetHash, err)
		}

		asset, err := assetBuilder.RemapFileWithInfo(u, h, assets.FileAssetInfo{
			Family:       "cni-plugins",
			Architecture: string(arch),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to remap CNI plugin binaries asset: %v", err)
		}

		return asset, nil
	}

	switch arch {
	case architectures.ArchitectureAmd64, architectures.ArchitectureArm64:
	default:
		return nil, fmt.Errorf("unknown arch for CNI plugin binaries asset: %s", arch)
	}

	var cniVersion string
	switch {
	case ig.KubernetesVersion().IsGTE("1.36"):
		cniVersion = defaultCNIVersionK8s_36
	case ig.KubernetesVersion().IsGTE("1.35"):
		cniVersion = defaultCNIVersionK8s_35
	case ig.KubernetesVersion().IsGTE("1.34"):
		cniVersion = defaultCNIVersionK8s_34
	case ig.KubernetesVersion().IsGTE("1.32"):
		cniVersion = defaultCNIVersionK8s_32
	default:
		return nil, fmt.Errorf("unknown CNI plugin binaries asset: %s", arch)
	}
	cniAssetURL = fmt.Sprintf(defaultCNIAssetURL, cniVersion, arch)
	klog.V(2).Infof("Adding CNI plugin binaries asset: %s", cniAssetURL)

	u, err := url.Parse(cniAssetURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse CNI plugin binaries asset URL %q: %v", cniAssetURL, err)
	}

	asset, err := assetBuilder.RemapFileWithInfo(u, nil, assets.FileAssetInfo{
		Family:       "cni-plugins",
		Version:      cniVersion,
		Architecture: string(arch),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to remap CNI plugin binaries asset: %v", err)
	}

	return asset, nil
}
