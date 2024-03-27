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

package wellknownassets

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
	// containerd packages URLs for v1.6.x+
	containerdReleaseUrlAmd64 = "https://github.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-amd64.tar.gz"
	containerdReleaseUrlArm64 = "https://github.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-arm64.tar.gz"
	// containerd packages URLs for v1.4.x+
	containerdBundleUrlAmd64 = "https://github.com/containerd/containerd/releases/download/v%s/cri-containerd-cni-%s-linux-amd64.tar.gz"
)

func FindContainerdAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	if c.Spec.Containerd == nil {
		return nil, nil, fmt.Errorf("unable to find containerd config")
	}
	containerd := c.Spec.Containerd

	if containerd.Packages != nil {
		if arch == architectures.ArchitectureAmd64 && containerd.Packages.UrlAmd64 != nil && containerd.Packages.HashAmd64 != nil {
			assetUrl := fi.ValueOf(containerd.Packages.UrlAmd64)
			assetHash := fi.ValueOf(containerd.Packages.HashAmd64)
			return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
		}
		if arch == architectures.ArchitectureArm64 && containerd.Packages.UrlArm64 != nil && containerd.Packages.HashArm64 != nil {
			assetUrl := fi.ValueOf(containerd.Packages.UrlArm64)
			assetHash := fi.ValueOf(containerd.Packages.HashArm64)
			return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
		}
	}

	version := fi.ValueOf(containerd.Version)
	if version == "" {
		return nil, nil, fmt.Errorf("unable to find containerd version")
	}

	return findContainerdVersionUrlHash(assetBuilder, arch, version)
}

func findContainerdVersionUrlHash(assetBuilder *assets.AssetBuilder, arch architectures.Architecture, version string) (u *url.URL, h *hashing.Hash, e error) {
	assetUrl, err := findContainerdVersionUrl(arch, version)
	if err != nil {
		return nil, nil, err
	}
	fileUrl, err := url.Parse(assetUrl)
	if err != nil {
		return nil, nil, err
	}

	return assetBuilder.RemapFileAndSHA(fileUrl)
}

func findContainerdVersionUrl(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("1.4.0")) {
		return "", fmt.Errorf("unsupported legacy containerd version: %q", version)
	}

	var u string
	switch arch {
	case architectures.ArchitectureAmd64:
		if sv.GTE(semver.MustParse("1.6.0")) {
			u = fmt.Sprintf(containerdReleaseUrlAmd64, version, version)
		} else {
			u = fmt.Sprintf(containerdBundleUrlAmd64, version, version)
		}
	case architectures.ArchitectureArm64:
		if sv.GTE(semver.MustParse("1.6.0")) {
			u = fmt.Sprintf(containerdReleaseUrlArm64, version, version)
		}
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	if u == "" {
		return "", fmt.Errorf("unknown url for containerd version: %s - %s", arch, version)
	}

	return u, nil
}

func findAssetsUrlHash(assetBuilder *assets.AssetBuilder, assetUrl string, assetHash string) (*url.URL, *hashing.Hash, error) {
	u, err := url.Parse(assetUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse asset URL %q: %v", assetUrl, err)
	}

	h, err := hashing.FromString(assetHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse asset hash %q: %v", assetHash, err)
	}

	u, err = assetBuilder.RemapFileAndSHAValue(u, assetHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to remap asset: %v", err)
	}

	return u, h, nil
}
