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

	"github.com/blang/semver/v4"

	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
)

const (
	runcVersionUrlAmd64 = "https://github.com/opencontainers/runc/releases/download/v%s/runc.amd64"
	runcVersionUrlArm64 = "https://github.com/opencontainers/runc/releases/download/v%s/runc.arm64"
)

func FindRuncAsset(ig model.InstanceGroup, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*assets.FileAsset, error) {
	containerd := ig.RawClusterSpec().Containerd
	if containerd == nil {
		return nil, fmt.Errorf("unable to find containerd config")
	}

	containerdVersion, err := semver.ParseTolerant(fi.ValueOf(containerd.Version))
	if err != nil {
		return nil, fmt.Errorf("unable to parse version string: %q", fi.ValueOf(containerd.Version))
	}
	// A compatible runc binary is bundled with containerd builds < v1.6.0
	// https://github.com/containerd/containerd/issues/6541
	if containerdVersion.LT(semver.MustParse("1.6.0")) {
		return nil, nil
	}

	if containerd.Runc == nil {
		return nil, fmt.Errorf("unable to find runc config")
	}
	runc := containerd.Runc

	canonicalURL := ""
	knownHash := ""

	if runc.Packages != nil {
		if arch == architectures.ArchitectureAmd64 && runc.Packages.UrlAmd64 != nil && runc.Packages.HashAmd64 != nil {
			canonicalURL = fi.ValueOf(runc.Packages.UrlAmd64)
			knownHash = fi.ValueOf(runc.Packages.HashAmd64)
		}
		if arch == architectures.ArchitectureArm64 && runc.Packages.UrlArm64 != nil && runc.Packages.HashArm64 != nil {
			canonicalURL = fi.ValueOf(runc.Packages.UrlArm64)
			knownHash = fi.ValueOf(runc.Packages.HashArm64)
		}
	}

	if canonicalURL == "" {
		version := fi.ValueOf(runc.Version)
		if version == "" {
			return nil, fmt.Errorf("unable to find runc version")
		}
		u, err := findRuncVersionUrl(arch, version)
		if err != nil {
			return nil, err
		}
		canonicalURL = u
	}

	return buildFileAsset(assetBuilder, canonicalURL, knownHash)
}

func findRuncVersionUrl(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("1.1.0")) {
		return "", fmt.Errorf("unsupported runc version: %q", version)
	}

	var u string
	switch arch {
	case architectures.ArchitectureAmd64:
		u = fmt.Sprintf(runcVersionUrlAmd64, version)
	case architectures.ArchitectureArm64:
		u = fmt.Sprintf(runcVersionUrlArm64, version)
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	if u == "" {
		return "", fmt.Errorf("unknown url for runc version: %s - %s", arch, version)
	}

	return u, nil
}
