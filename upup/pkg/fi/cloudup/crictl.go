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
	crictlAssetUrlAmd64 = "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-amd64.tar.gz"
	crictlAssetUrlArm64 = "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-arm64.tar.gz"
)

func findCrictlAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	var assetURL string
	switch arch {
	case architectures.ArchitectureAmd64:
		assetURL = crictlAssetUrlAmd64
	case architectures.ArchitectureArm64:
		assetURL = crictlAssetUrlArm64
	default:
		return nil, nil, fmt.Errorf("unknown arch for crictl binaries asset: %s", arch)
	}

	u, err := url.Parse(assetURL)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse crictl binaries asset URL %q: %v", assetURL, err)
	}

	u, h, err := assetBuilder.RemapFileAndSHA(u)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to remap crictl binaries asset: %v", err)
	}

	return u, h, err
}
