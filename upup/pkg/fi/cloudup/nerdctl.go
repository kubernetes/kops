package cloudup

import (
	"fmt"
	"net/url"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

const (
	nerdctlAssetUrlAmd64  = "https://github.com/containerd/nerdctl/releases/download/v1.7.4/nerdctl-1.7.4-linux-amd64.tar.gz"
	nerdctlAssetUrlArm64  = "https://github.com/containerd/nerdctl/releases/download/v1.7.4/nerdctl-1.7.4-linux-arm64.tar.gz"
	nerdctlAssetHashAmd64 = "71aee9d987b7fad0ff2ade50b038ad7e2356324edc02c54045960a3521b3e6a7"
	nerdctlAssetHashArm64 = "d8df47708ca57b9cd7f498055126ba7dcfc811d9ba43aae1830c93a09e70e22d"
)

func findNerdctlAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	if c.Spec.Containerd == nil {
		return nil, nil, fmt.Errorf("unable to find containerd config")
	}

	var assetURL, assetHash string
	switch arch {
	case architectures.ArchitectureAmd64:
		assetURL = nerdctlAssetUrlAmd64
		assetHash = nerdctlAssetHashAmd64
	case architectures.ArchitectureArm64:
		assetURL = nerdctlAssetUrlArm64
		assetHash = nerdctlAssetHashArm64
	default:
		return nil, nil, fmt.Errorf("unknown arch for nerdctl binaries asset: %s", arch)
	}

	return findAssetsUrlHash(assetBuilder, assetURL, assetHash)
}
