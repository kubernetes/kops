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

	"k8s.io/kops/util/pkg/architectures"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/hashing"
)

const crioVersionUrl = "https://storage.googleapis.com/k8s-conform-cri-o/artifacts/cri-o.%s.v%s.tar.gz"

func findCrioAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	if c.Spec.Crio == nil || fi.StringValue(c.Spec.Crio.Version) == "" {
		return nil, nil, fmt.Errorf("unable to find crio version")
	}

	crio := c.Spec.Crio

	if crio.Packages != nil {
		if arch == architectures.ArchitectureAmd64 && crio.Packages.UrlAmd64 != nil && crio.Packages.HashAmd64 != nil {
			assetUrl := fi.StringValue(crio.Packages.UrlAmd64)
			assetHash := fi.StringValue(crio.Packages.HashAmd64)
			return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
		}
		if arch == architectures.ArchitectureArm64 && crio.Packages.UrlArm64 != nil && crio.Packages.HashArm64 != nil {
			assetUrl := fi.StringValue(crio.Packages.UrlArm64)
			assetHash := fi.StringValue(crio.Packages.HashArm64)
			return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
		}
	}

	version := fi.StringValue(c.Spec.Crio.Version)

	assetUrl, assetHash, err := findCrioVersionUrlHash(arch, version)
	if err != nil {
		return nil, nil, err
	}

	return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
}

func findCrioVersionUrlHash(arch architectures.Architecture, version string) (u string, h string, e error) {
	crioUrl, err := findCrioVersionUrl(arch, version)
	if err != nil {
		return "", "", err
	}
	crioHash, err := findCrioVersionHash(arch, version)
	if err != nil {
		return "", "", err
	}

	return crioUrl, crioHash, nil
}

func findCrioVersionUrl(arch architectures.Architecture, version string) (string, error) {
	if crioVersionNotSupported(version) {
		return "", fmt.Errorf("crio version not supported")
	}

	return fmt.Sprintf(crioVersionUrl, arch, version), nil
}

func findCrioVersionHash(arch architectures.Architecture, version string) (string, error) {
	versionHash, ok := getCrioVersionHash(arch)[version]

	if ok {
		return versionHash, nil
	} else {
		return "", fmt.Errorf("version hash not found")
	}
}

func getCrioVersionHash(arch architectures.Architecture) map[string]string {
	versionHash := map[architectures.Architecture]map[string]string{
		architectures.ArchitectureAmd64: {
			"1.21.0": "75d1aefd93d9b6eea9627d20c1a8b65307e10396b7c28d440308bb351ca041bb",
		},
		architectures.ArchitectureArm64: {
			"1.21.0": "a3dc626ecd8ecd0561b6219abe49e5e64d86e5d45a1a6647eb10d9fd0dd3c8f7",
		},
	}

	return versionHash[arch]
}

func crioVersionNotSupported(version string) bool {
	switch version {
	case
		"1.21.0":
		return false
	}

	return true
}
