/*
Copyright 2016 The Kubernetes Authors.

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

	"github.com/blang/semver"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
)

func usesCNI(c *api.Cluster) bool {
	networkConfig := c.Spec.Networking
	if networkConfig == nil || networkConfig.Classic != nil {
		// classic
		return false
	}

	if networkConfig.Kubenet != nil {
		// kubenet
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

	if networkConfig.AmazonVPCIPVlan != nil {
		// AmazonVPCIPVlan uses CNI
		return true
	}

	// Assume other modes also use CNI
	glog.Warningf("Unknown networking mode configured")
	return true
}

// TODO: we really need to sort this out:
// https://github.com/kubernetes/kops/issues/724
// https://github.com/kubernetes/kops/issues/626
// https://github.com/kubernetes/kubernetes/issues/30338

const (
	// 1.5.x k8s uses release 07a8a28637e97b22eb8dfe710eeae1344f69d16e
	defaultCNIAssetK8s1_5           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-07a8a28637e97b22eb8dfe710eeae1344f69d16e.tar.gz"
	defaultCNIAssetHashStringK8s1_5 = "19d49f7b2b99cd2493d5ae0ace896c64e289ccbb"

	// 1.6.x k8s uses release 0799f5732f2a11b329d9e3d51b9c8f2e3759f2ff
	defaultCNIAssetK8s1_6           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-0799f5732f2a11b329d9e3d51b9c8f2e3759f2ff.tar.gz"
	defaultCNIAssetHashStringK8s1_6 = "1d9788b0f5420e1a219aad2cb8681823fc515e7c"

	// defaultCNIAssetK8s1_9 is the CNI tarball for 1.9.x k8s.
	defaultCNIAssetK8s1_9           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-plugins-amd64-v0.6.0.tgz"
	defaultCNIAssetHashStringK8s1_9 = "d595d3ded6499a64e8dac02466e2f5f2ce257c9f"

	// Environment variable for overriding CNI url
	ENV_VAR_CNI_VERSION_URL       = "CNI_VERSION_URL"
	ENV_VAR_CNI_ASSET_HASH_STRING = "CNI_ASSET_HASH_STRING"
)

func findCNIAssets(c *api.Cluster, assetBuilder *assets.AssetBuilder) (*url.URL, string, error) {

	if cniVersionURL := os.Getenv(ENV_VAR_CNI_VERSION_URL); cniVersionURL != "" {
		u, err := url.Parse(cniVersionURL)
		if err != nil {
			return nil, "", fmt.Errorf("unable to parse %q as a URL: %v", cniVersionURL, err)
		}

		glog.Infof("Using CNI asset version %q, as set in %s", cniVersionURL, ENV_VAR_CNI_VERSION_URL)

		if cniAssetHashString := os.Getenv(ENV_VAR_CNI_ASSET_HASH_STRING); cniAssetHashString != "" {

			glog.Infof("Using CNI asset hash %q, as set in %s", cniAssetHashString, ENV_VAR_CNI_ASSET_HASH_STRING)

			return u, cniAssetHashString, nil
		} else {
			return u, "", nil
		}
	}

	sv, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
	if err != nil {
		return nil, "", fmt.Errorf("failed to lookup kubernetes version: %v", err)
	}

	sv.Pre = nil
	sv.Build = nil

	var cniAsset, cniAssetHash string
	if sv.GTE(semver.Version{Major: 1, Minor: 9, Patch: 0, Pre: nil, Build: nil}) {
		cniAsset = defaultCNIAssetK8s1_9
		cniAssetHash = defaultCNIAssetHashStringK8s1_9
		glog.V(2).Infof("Adding default CNI asset for k8s 1.9.x and higher: %s", defaultCNIAssetK8s1_9)
	} else if sv.GTE(semver.Version{Major: 1, Minor: 6, Patch: 0, Pre: nil, Build: nil}) {
		cniAsset = defaultCNIAssetK8s1_6
		cniAssetHash = defaultCNIAssetHashStringK8s1_6
		glog.V(2).Infof("Adding default CNI asset for k8s 1.6.x and higher: %s", defaultCNIAssetK8s1_6)
	} else {
		cniAsset = defaultCNIAssetK8s1_5
		cniAssetHash = defaultCNIAssetHashStringK8s1_5
		glog.V(2).Infof("Adding default CNI asset for k8s 1.5: %s", defaultCNIAssetK8s1_5)
	}

	u, err := url.Parse(cniAsset)
	if err != nil {
		return nil, "", nil
	}

	u, err = assetBuilder.RemapFileAndSHAValue(u, cniAssetHash)
	if err != nil {
		return nil, "", err
	}

	return u, cniAssetHash, nil
}
