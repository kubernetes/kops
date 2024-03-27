/*
Copyright 2024 The Kubernetes Authors.

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

func FindNerdctlAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
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
