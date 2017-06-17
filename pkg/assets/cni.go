package assets

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops/util"
	"os"
)

// TODO: we really need to sort this out:
// https://github.com/kubernetes/kops/issues/724
// https://github.com/kubernetes/kops/issues/626
// https://github.com/kubernetes/kubernetes/issues/30338

// 1.5.x k8s uses release 07a8a28637e97b22eb8dfe710eeae1344f69d16e
var defaultCNIAssetK8s1_5 = &Asset{
	Origin: "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-07a8a28637e97b22eb8dfe710eeae1344f69d16e.tar.gz",
	Hash:   "19d49f7b2b99cd2493d5ae0ace896c64e289ccbb",
}

// 1.6.x k8s uses release 0799f5732f2a11b329d9e3d51b9c8f2e3759f2ff
var defaultCNIAssetK8s1_6 = &Asset{
	Origin: "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-0799f5732f2a11b329d9e3d51b9c8f2e3759f2ff.tar.gz",
	Hash:   "1d9788b0f5420e1a219aad2cb8681823fc515e7c",
}

// Environment variable for overriding CNI url
const ENV_VAR_CNI_VERSION_URL = "CNI_VERSION_URL"

func (a *AssetBuilder) FindCNIAssets() (string, string, error) {

	if cniVersionURL := os.Getenv(ENV_VAR_CNI_VERSION_URL); cniVersionURL != "" {
		glog.Infof("Using CNI asset version %q, as set in %s", cniVersionURL, ENV_VAR_CNI_VERSION_URL)

		a.addURLAsset(cniVersionURL)

		return cniVersionURL, "", nil
	}

	sv, err := util.ParseKubernetesVersion(a.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return "", "", fmt.Errorf("Failed to lookup kubernetes version: %v", err)
	}

	sv.Pre = nil
	sv.Build = nil

	var asset *Asset
	if sv.GTE(semver.Version{Major: 1, Minor: 6, Patch: 0, Pre: nil, Build: nil}) {
		asset = defaultCNIAssetK8s1_6
	} else {
		asset = defaultCNIAssetK8s1_5
	}

	glog.V(2).Infof("Adding default CNI asset: %s", asset.Origin)

	a.addAsset(asset)

	return asset.Origin, asset.Hash, nil
}
