/*
Copyright 2020 The Kubernetes Authors.

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

	"github.com/blang/semver/v4"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

const (
	// containerd packages URLs for v1.4.x+
	containerdVersionUrlAmd64 = "https://github.com/containerd/containerd/releases/download/v%s/cri-containerd-cni-%s-linux-amd64.tar.gz"
	// containerd legacy packages URLs for v1.2.x and v1.3.x
	containerdLegacyUrlAmd64 = "https://storage.googleapis.com/cri-containerd-release/cri-containerd-%s.linux-amd64.tar.gz"
	// containerd version that is available for both AMD64 and ARM64, used in case the selected version is not available for ARM64
	containerdFallbackVersion = "1.2.13"
)

func findContainerdAssets(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	if c.Spec.Containerd == nil || fi.StringValue(c.Spec.Containerd.Version) == "" {
		return nil, nil, fmt.Errorf("unable to find containerd version")
	}

	version := fi.StringValue(c.Spec.Containerd.Version)

	assetUrl, assetHash, err := findContainerdVersionUrlHash(arch, version)
	if err != nil {
		return nil, nil, err
	}

	return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
}

func findContainerdVersionUrlHash(arch architectures.Architecture, version string) (u string, h string, e error) {
	var containerdAssetUrl, containerdAssetHash string

	if findAllContainerdHashesAmd64()[version] != "" {
		var err error
		containerdAssetUrl, err = findContainerdVersionUrl(arch, version)
		if err != nil {
			return "", "", err
		}
		containerdAssetHash, err = findContainerdVersionHash(arch, version)
		if err != nil {
			return "", "", err
		}
	} else {
		// Fall back to Docker packages
		dv := findAllContainerdDockerMappings()[version]
		if dv != "" {
			var err error
			containerdAssetUrl, err = findDockerVersionUrl(arch, dv)
			if err != nil {
				return "", "", err
			}
			containerdAssetHash, err = findDockerVersionHash(arch, dv)
			if err != nil {
				return "", "", err
			}
			println(dv)
		} else {
			return "", "", fmt.Errorf("unknown url and hash for containerd version: %s - %s", arch, version)
		}
	}

	return containerdAssetUrl, containerdAssetHash, nil
}

func findContainerdVersionUrl(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("1.3.4")) {
		return "", fmt.Errorf("unsupported legacy containerd version: %q", version)
	}

	var u string
	switch arch {
	case architectures.ArchitectureAmd64:
		if sv.GTE(semver.MustParse("1.4.0")) {
			u = fmt.Sprintf(containerdVersionUrlAmd64, version, version)
		} else {
			u = fmt.Sprintf(containerdLegacyUrlAmd64, version)
		}
	case architectures.ArchitectureArm64:
		// For now there are only official AMD64 builds, using Default Docker version instead
		if findAllContainerdHashesAmd64()[version] != "" {
			u = fmt.Sprintf(dockerVersionUrlArm64, findAllContainerdDockerMappings()[containerdFallbackVersion])
		}
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	if u == "" {
		return "", fmt.Errorf("unknown url for containerd version: %s - %s", arch, version)
	}

	return u, nil
}

func findContainerdVersionHash(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("1.3.4")) {
		return "", fmt.Errorf("unsupported legacy containerd version: %q", version)
	}

	var h string
	switch arch {
	case architectures.ArchitectureAmd64:
		h = findAllContainerdHashesAmd64()[version]
	case architectures.ArchitectureArm64:
		// For now there are only official AMD64 builds, using Default Docker version instead
		if findAllContainerdHashesAmd64()[version] != "" {
			h = findAllDockerHashesArm64()[findAllContainerdDockerMappings()[containerdFallbackVersion]]
		}
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	if h == "" {
		return "", fmt.Errorf("unknown hash for containerd version: %s - %s", arch, version)
	}

	return h, nil
}

func findAllContainerdHashesAmd64() map[string]string {
	hashes := map[string]string{
		"1.3.4": "4616971c3ad21c24f2f2320fa1c085577a91032a068dd56a41c7c4b71a458087",
		"1.4.0": "b379f29417efd583f77e095173d4d0bd6bb001f0081b2a63d152ee7aef653ce1",
		"1.4.1": "757efb93a4f3161efc447a943317503d8a7ded5cb4cc0cba3f3318d7ce1542ed",
	}

	return hashes
}

func findAllContainerdDockerMappings() map[string]string {
	versions := map[string]string{
		"1.2.6":  "19.03.2",
		"1.2.10": "19.03.5",
		"1.2.12": "19.03.6",
		"1.2.13": "19.03.12",
		"1.3.7":  "19.03.13",
	}

	return versions
}
