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
	runcVersionUrlAmd64 = "https://github.com/opencontainers/runc/releases/download/v%s/runc.amd64"
	runcVersionUrlArm64 = "https://github.com/opencontainers/runc/releases/download/v%s/runc.arm64"
)

func findRuncAsset(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	if c.Spec.Containerd == nil {
		return nil, nil, fmt.Errorf("unable to find containerd config")
	}
	containerd := c.Spec.Containerd

	containerdVersion, err := semver.ParseTolerant(fi.ValueOf(containerd.Version))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse version string: %q", fi.ValueOf(containerd.Version))
	}
	// A compatible runc binary is bundled with containerd builds < v1.6.0
	// https://github.com/containerd/containerd/issues/6541
	if containerdVersion.LT(semver.MustParse("1.6.0")) {
		return nil, nil, nil
	}

	if containerd.Runc == nil {
		return nil, nil, fmt.Errorf("unable to find runc config")
	}
	runc := containerd.Runc

	if runc.Packages != nil {
		if arch == architectures.ArchitectureAmd64 && runc.Packages.UrlAmd64 != nil && runc.Packages.HashAmd64 != nil {
			assetUrl := fi.ValueOf(runc.Packages.UrlAmd64)
			assetHash := fi.ValueOf(runc.Packages.HashAmd64)
			return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
		}
		if arch == architectures.ArchitectureArm64 && runc.Packages.UrlArm64 != nil && runc.Packages.HashArm64 != nil {
			assetUrl := fi.ValueOf(runc.Packages.UrlArm64)
			assetHash := fi.ValueOf(runc.Packages.HashArm64)
			return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
		}
	}

	version := fi.ValueOf(runc.Version)
	if version == "" {
		return nil, nil, fmt.Errorf("unable to find runc version")
	}
	assetUrl, assetHash, err := findRuncVersionUrlHash(arch, version)
	if err != nil {
		return nil, nil, err
	}

	return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
}

func findRuncVersionUrlHash(arch architectures.Architecture, version string) (u string, h string, e error) {
	var runcAssetUrl, runcAssetHash string

	if findAllRuncHashesAmd64()[version] != "" {
		var err error
		runcAssetUrl, err = findRuncVersionUrl(arch, version)
		if err != nil {
			return "", "", err
		}
		runcAssetHash, err = findRuncVersionHash(arch, version)
		if err != nil {
			return "", "", err
		}
	} else {
		return "", "", fmt.Errorf("unknown url and hash for runc version: %s - %s", arch, version)
	}

	return runcAssetUrl, runcAssetHash, nil
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

func findRuncVersionHash(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("1.1.0")) {
		return "", fmt.Errorf("unsupported runc version: %q", version)
	}

	var h string
	switch arch {
	case architectures.ArchitectureAmd64:
		h = findAllRuncHashesAmd64()[version]
	case architectures.ArchitectureArm64:
		h = findAllRuncHashesArm64()[version]
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	if h == "" {
		return "", fmt.Errorf("unknown hash for runc version: %s - %s", arch, version)
	}

	return h, nil
}

func findAllRuncHashesAmd64() map[string]string {
	hashes := map[string]string{
		"1.1.0":  "ab1c67fbcbdddbe481e48a55cf0ef9a86b38b166b5079e0010737fd87d7454bb",
		"1.1.1":  "5798c85d2c8b6942247ab8d6830ef362924cd72a8e236e77430c3ab1be15f080",
		"1.1.2":  "e0436dfc5d26ca88f00e84cbdab5801dd9829b1e5ded05dcfc162ce5718c32ce",
		"1.1.3":  "6e8b24be90fffce6b025d254846da9d2ca6d65125f9139b6354bab0272253d01",
		"1.1.4":  "db772be63147a4e747b4fe286c7c16a2edc4a8458bd3092ea46aaee77750e8ce",
		"1.1.5":  "f00b144e86f8c1db347a2e8f22caade07d55382c5f76dd5c0a5b1ab64eaec8bb",
		"1.1.6":  "868bee5b8dc2a01df0ca41d0accfad6a3372dc1165ebfb76143d2c6672e86115",
		"1.1.7":  "c3aadb419e5872af49504b6de894055251d2e685fddddb981a79703e7f895cbd",
		"1.1.8":  "1d05ed79854efc707841dfc7afbf3b86546fc1d0b3a204435ca921c14af8385b",
		"1.1.9":  "b9bfdd4cb27cddbb6172a442df165a80bfc0538a676fbca1a6a6c8f4c6933b43",
		"1.1.10": "81f73a59be3d122ab484d7dfe9ddc81030f595cc59968f61c113a9a38a2c113a",
		"1.1.11": "77ae134de014613c44d25e6310a57a219a7a91155cd47d069a0f22a2cad5caea",
		"1.1.12": "aadeef400b8f05645768c1476d1023f7875b78f52c7ff1967a6dbce236b8cbd8",
	}

	return hashes
}

func findAllRuncHashesArm64() map[string]string {
	hashes := map[string]string{
		"1.1.0":  "9ec8e68feabc4e7083a4cfa45ebe4d529467391e0b03ee7de7ddda5770b05e68",
		"1.1.1":  "20c436a736547309371c7ac2a335f5fe5a42b450120e497d09c8dc3902c28444",
		"1.1.2":  "6ebd968d46d00a3886e9a0cae2e0a7b399e110cf5d7b26e63ce23c1d81ea10ef",
		"1.1.3":  "00c9ad161a77a01d9dcbd25b1d76fa9822e57d8e4abf26ba8907c98f6bcfcd0f",
		"1.1.4":  "dbb71e737eaef454a406ce21fd021bd8f1b35afb7635016745992bbd7c17a223",
		"1.1.5":  "54e79e4d48b9e191767e4abc08be1a8476a1c757e9a9f8c45c6ded001226867f",
		"1.1.6":  "da5b2ed26a173a69ea66eae7c369feebf59c1031e14985f512a0a293bb5f76fb",
		"1.1.7":  "1b309c4d5aa4cc7b888b2f79c385ecee26ca3d55dae0852e7c4a692196d5faab",
		"1.1.8":  "7c22cb618116d1d5216d79e076349f93a672253d564b19928a099c20e4acd658",
		"1.1.9":  "b43e9f561e85906f469eef5a7b7992fc586f750f44a0e011da4467e7008c33a0",
		"1.1.10": "4830afd426bdeacbdf9cb8729524aa2ed51790b8c4b28786995925593708f1c8",
		"1.1.11": "9f1ee53f06b78cc4a115ca6ae4eec10567999539ce828a22c5351edba043ed12",
		"1.1.12": "879f910a05c95c10c64ad8eb7d5e3aa8e4b30e65587b3d68e009a3565aed5bb8",
	}

	return hashes
}
