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
	crictlAssetUrlAmd64 = "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-amd64.tar.gz"
	crictlAssetUrlArm64 = "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-arm64.tar.gz"
)

func FindCrictlAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
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
