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
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"os"
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

	if networkConfig.CNI != nil {
		// CNI definitely uses CNI!
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
	defaultCNIAsset           = "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-07a8a28637e97b22eb8dfe710eeae1344f69d16e.tar.gz"
	defaultCNIAssetHashString = "19d49f7b2b99cd2493d5ae0ace896c64e289ccbb"
	ENV_VAR_CNI_VERSION_URL   = "CNI_VERSION_URL"
)

func findCNIAssets(c *api.Cluster) (string, string) {

	if cniVersionURL := os.Getenv(ENV_VAR_CNI_VERSION_URL); cniVersionURL != "" {
		glog.Infof("Using CNI asset version %q, as set in %s", cniVersionURL, ENV_VAR_CNI_VERSION_URL)
		return cniVersionURL, ""
	}

	glog.V(2).Infof("Adding default CNI asset: %s", defaultCNIAsset)
	return defaultCNIAsset, defaultCNIAssetHashString
}
