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
	"k8s.io/kops/util/pkg/hashing"
)

func usesCNI(c *kopsapi.Cluster) bool {
	networkConfig := c.Spec.Networking
	if networkConfig == nil || networkConfig.Classic != nil {
		// classic
		return false
	}

	if networkConfig.Kubenet != nil {
		// kubenet is now configured via CNI
		return true
	}

	if networkConfig.GCE != nil {
		// GCE is kubenet at the node level
		return true
	}

	if networkConfig.External != nil {
		// external: assume uses CNI
		return true
	}

	if networkConfig.Kopeio != nil {
		// Kopeio uses kubenet (and thus CNI)
		return true
	}

	if networkConfig.Weave != nil {
		//  Weave uses CNI
		return true
	}

	if networkConfig.Flannel != nil {
		//  Flannel uses CNI
		return true
	}

	if networkConfig.Calico != nil {
		//  Calico uses CNI
		return true
	}

	if networkConfig.Canal != nil {
		// Canal uses CNI
		return true
	}

	if networkConfig.Kuberouter != nil {
		// Kuberouter uses CNI
		return true
	}

	if networkConfig.Romana != nil {
		//  Romana uses CNI
		return true
	}

	if networkConfig.AmazonVPC != nil {
		//  AmazonVPC uses CNI
		return true
	}

	if networkConfig.Cilium != nil {
		// Cilium uses CNI
		return true
	}

	if networkConfig.CNI != nil {
		// CNI definitely uses CNI!
		return true
	}

	if networkConfig.LyftVPC != nil {
		// LyftVPC uses CNI
		return true
	}

	// Assume other modes also use CNI
	klog.Warningf("Unknown networking mode configured")
	return true
}

// TODO: we really need to sort this out:
// https://github.com/kubernetes/kops/issues/724
// https://github.com/kubernetes/kops/issues/626
// https://github.com/kubernetes/kubernetes/issues/30338

const (
	// defaultCNIAssetK8s1_9 is the CNI tarball for 1.9.x k8s.
	defaultCNIAssetK8s1_9           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-amd64-v0.6.0.tgz"
	defaultCNIAssetHashStringK8s1_9 = "d595d3ded6499a64e8dac02466e2f5f2ce257c9f"

	// defaultCNIAssetK8s1_11 is the CNI tarball for k8s >= 1.11
	defaultCNIAssetK8s1_11             = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-amd64-v0.7.5.tgz"
	defaultCNIAssetSHA1StringK8s1_11   = "52e9d2de8a5f927307d9397308735658ee44ab8d"
	defaultCNIAssetSHA256StringK8s1_11 = "3ca15c0a18ee830520cf3a95408be826cbd255a1535a38e0be9608b25ad8bf64"

	// defaultCNIAssetK8s1_18 is the CNI tarball for k8s >= 1.18
	defaultCNIAssetK8s1_18             = "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.5/cni-plugins-linux-amd64-v0.8.5.tgz"
	defaultCNIAssetSHA256StringK8s1_18 = "bd682ffcf701e8f83283cdff7281aad0c83b02a56084d6e601216210732833f9"

	// Environment variable for overriding CNI url
	ENV_VAR_CNI_VERSION_URL       = "CNI_VERSION_URL"
	ENV_VAR_CNI_ASSET_HASH_STRING = "CNI_ASSET_HASH_STRING"
)

func findCNIAssets(c *kopsapi.Cluster, assetBuilder *assets.AssetBuilder) (*url.URL, *hashing.Hash, error) {

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
		} else {
			return u, nil, nil
		}
	}

	sv, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup kubernetes version: %v", err)
	}

	var cniAsset, cniAssetHash string
	if util.IsKubernetesGTE("1.18", *sv) {
		cniAsset = defaultCNIAssetK8s1_18
		cniAssetHash = defaultCNIAssetSHA256StringK8s1_18
		klog.V(2).Infof("Adding default CNI asset for k8s >= 1.18: %s", cniAsset)
	} else if util.IsKubernetesGTE("1.15", *sv) {
		// We're still on the same asset, but we use sha256
		cniAsset = defaultCNIAssetK8s1_11
		cniAssetHash = defaultCNIAssetSHA256StringK8s1_11
		klog.V(2).Infof("Adding default CNI asset for 1.18 > k8s >= 1.11: %s", cniAsset)
	} else if util.IsKubernetesGTE("1.11", *sv) {
		cniAsset = defaultCNIAssetK8s1_11
		cniAssetHash = defaultCNIAssetSHA1StringK8s1_11
		klog.V(2).Infof("Adding default CNI asset for 1.18 > k8s >= 1.11: %s", cniAsset)
	} else {
		cniAsset = defaultCNIAssetK8s1_9
		cniAssetHash = defaultCNIAssetHashStringK8s1_9
		klog.V(2).Infof("Adding default CNI asset for 1.11 > k8s >= 1.9: %s", cniAsset)
	}

	u, err := url.Parse(cniAsset)
	if err != nil {
		return nil, nil, nil
	}

	hash, err := hashing.FromString(cniAssetHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse CNI asset hash %q", cniAssetHash)
	}

	u, err = assetBuilder.RemapFileAndSHAValue(u, cniAssetHash)
	if err != nil {
		return nil, nil, err
	}

	return u, hash, nil
}
