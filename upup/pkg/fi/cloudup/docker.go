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
	// Docker packages URLs for v18.09.0+
	dockerVersionUrlAmd64 = "https://download.docker.com/linux/static/stable/x86_64/docker-%s.tgz"
	dockerVersionUrlArm64 = "https://download.docker.com/linux/static/stable/aarch64/docker-%s.tgz"
	// Docker legacy AMD64 packages URLs for v17.03.0 to v18.06.3
	dockerLegacyUrlAmd64 = "https://download.docker.com/linux/static/stable/x86_64/docker-%s-ce.tgz"
	// Docker legacy ARM64 packages URLs for v17.09.0 to v18.06.3
	dockerLegacyUrlArm64 = "https://download.docker.com/linux/static/stable/aarch64/docker-%s-ce.tgz"
	// Docker version that is available for both AMD64 and ARM64, used in case the selected version is too old and not available for ARM64
	dockerFallbackVersion = "17.09.0"
)

func findDockerAssets(c *kops.Cluster, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*url.URL, *hashing.Hash, error) {
	if c.Spec.Docker == nil || fi.StringValue(c.Spec.Docker.Version) == "" {
		return nil, nil, fmt.Errorf("unable to find Docker version")
	}

	version := fi.StringValue(c.Spec.Docker.Version)

	assetUrl, assetHash, err := findDockerVersionUrlHash(arch, version)
	if err != nil {
		return nil, nil, err
	}

	return findAssetsUrlHash(assetBuilder, assetUrl, assetHash)
}

func findDockerVersionUrlHash(arch architectures.Architecture, version string) (u string, h string, e error) {
	dockerAssetUrl, err := findDockerVersionUrl(arch, version)
	if err != nil {
		return "", "", err
	}
	dockerAssetHash, err := findDockerVersionHash(arch, version)
	if err != nil {
		return "", "", err
	}

	return dockerAssetUrl, dockerAssetHash, nil
}

func findDockerVersionUrl(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("17.3.0")) {
		return "", fmt.Errorf("unsupported legacy Docker version: %q", version)
	}

	var u string
	switch arch {
	case architectures.ArchitectureAmd64:
		if sv.GTE(semver.MustParse("18.9.0")) {
			u = fmt.Sprintf(dockerVersionUrlAmd64, version)
		} else {
			u = fmt.Sprintf(dockerLegacyUrlAmd64, version)
		}
	case architectures.ArchitectureArm64:
		if sv.GTE(semver.MustParse("18.9.0")) {
			u = fmt.Sprintf(dockerVersionUrlArm64, version)
		} else if sv.GTE(semver.MustParse("17.9.0")) {
			u = fmt.Sprintf(dockerLegacyUrlArm64, version)
		} else {
			u = fmt.Sprintf(dockerLegacyUrlArm64, dockerFallbackVersion)
		}
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	return u, nil
}

func findDockerVersionHash(arch architectures.Architecture, version string) (string, error) {
	sv, err := semver.ParseTolerant(version)
	if err != nil {
		return "", fmt.Errorf("unable to parse version string: %q", version)
	}
	if sv.LT(semver.MustParse("17.3.0")) {
		return "", fmt.Errorf("unsupported legacy Docker version: %q", version)
	}

	var h string
	switch arch {
	case architectures.ArchitectureAmd64:
		h = findAllDockerHashesAmd64()[version]
	case architectures.ArchitectureArm64:
		if sv.GTE(semver.MustParse("17.9.0")) {
			h = findAllDockerHashesArm64()[version]
		} else {
			h = findAllDockerHashesArm64()[dockerFallbackVersion]
		}
	default:
		return "", fmt.Errorf("unknown arch: %q", arch)
	}

	if h == "" {
		return "", fmt.Errorf("unknown hash for Docker version: %s - %s", arch, version)
	}

	return h, nil
}

func findAllDockerHashesAmd64() map[string]string {
	hashes := map[string]string{
		"17.03.0":  "aac08524db82d3fdc8fc092f495e1174f5e1dd774b95a6d081544997d34b4855",
		"17.03.1":  "3e070e7b34e99cf631f44d0ff5cf9a127c0b8af5c53dfc3e1fce4f9615fbf603",
		"17.03.2":  "183b31b001e7480f3c691080486401aa519101a5cfe6e05ad01b9f5521c4112d",
		"17.06.0":  "e582486c9db0f4229deba9f8517145f8af6c5fae7a1243e6b07876bd3e706620",
		"17.06.1":  "e35fe12806eadbb7eb8aa63e3dfb531bda5f901cd2c14ac9cdcd54df6caed697",
		"17.06.2":  "a15f62533e773c40029a61784a5a1c5bc7dd21e0beb5402fda109f80e1f2994d",
		"17.09.0":  "a9e90a73c3cdfbf238f148e1ec0eaff5eb181f92f35bdd938fd7dab18e1c4647",
		"17.09.1":  "77d3eaa72f2b63c94ea827b548f4a8b572b754a431c59258e3f2730411f64be7",
		"17.12.0":  "692e1c72937f6214b1038def84463018d8e320c8eaf8530546c84c2f8f9c767d",
		"17.12.1":  "1270dce1bd7e1838d62ae21d2505d87f16efc1d9074645571daaefdfd0c14054",
		"18.03.0":  "e5dff6245172081dbf14285dafe4dede761f8bc1750310156b89928dbf56a9ee",
		"18.03.1":  "0e245c42de8a21799ab11179a4fce43b494ce173a8a2d6567ea6825d6c5265aa",
		"18.06.0":  "1c2fa625496465c68b856db0ba850eaad7a16221ca153661ca718de4a2217705",
		"18.06.1":  "83be159cf0657df9e1a1a4a127d181725a982714a983b2bdcc0621244df93687",
		"18.06.2":  "a979d9a952fae474886c7588da692ee00684cb2421d2c633c7ed415948cf0b10",
		"18.06.3":  "346f9394393ee8db5f8bd1e229ee9d90e5b36931bdd754308b2ae68884dd6822",
		"18.09.0":  "08795696e852328d66753963249f4396af2295a7fe2847b839f7102e25e47cb9",
		"18.09.1":  "c9959e42b637fb7362899ac1d1aeef2a966fa0ea85631da91f4c4a7a9ec29644",
		"18.09.2":  "183e10448f0c3a0dc82c9d504c5280c29527b89af0fc71cb27115d684b26c8bd",
		"18.09.3":  "8b886106cfc362f1043debfe178c35b6f73ec42380b034a3919a235fe331e053",
		"18.09.4":  "7baf380a9d503286b3745114e0d8e265897edd9642747b1992459e550fc5c827",
		"18.09.5":  "99ca9395e9c7ffbf75537de71aa828761f492491d02bc6e29db2920fa582c6c5",
		"18.09.6":  "1f3f6774117765279fce64ee7f76abbb5f260264548cf80631d68fb2d795bb09",
		"18.09.7":  "e106ccfa2b1f60794faaa6bae57a2dac9dc4cb33e5541fad6a826ea525d01cc4",
		"18.09.8":  "12277eff64363f51ba2f20dd258bdc2c3248022996c0251921193ec6fd179e52",
		"18.09.9":  "82a362af7689038c51573e0fd0554da8703f0d06f4dfe95dd5bda5acf0ae45fb",
		"19.03.0":  "b7bb0c3610b3f6ee87457dfb440968dbcc3537198c3d6e2468fcf90819855d6f",
		"19.03.1":  "6e7d8e24ee46b13d7547d751696d01607d19c8224c1b2c867acc8c779e77734b",
		"19.03.2":  "865038730c79ab48dfed1365ee7627606405c037f46c9ae17c5ec1f487da1375",
		"19.03.3":  "c3c8833e227b61fe6ce0bc5c17f97fa547035bef4ef17cf6601f30b0f20f4ce5",
		"19.03.4":  "efef2ad32d262674501e712351be0df9dd31d6034b175d0020c8f5d5c9c3fd10",
		"19.03.5":  "50cdf38749642ec43d6ac50f4a3f1f7f6ac688e8d8b4e1c5b7be06e1a82f06e9",
		"19.03.6":  "34ff89ce917796594cd81149b1777d07786d297ffd0fef37a796b5897052f7cc",
		"19.03.7":  "033e97ae6b31e21c598fd089ea034c08d75dc744ceb787898d63dfc4e45ead03",
		"19.03.8":  "7f4115dc6a3c19c917f8b9664d7b51c904def1c984e082c4600097433323cf6f",
		"19.03.9":  "1c03c78be198d9085e7dd6806fc5d93264baaf0c7ea17f584d00af48eae508ee",
		"19.03.10": "7c1576a0bc749418d1423d2b78c8920b5d61f849789904612862dd118742e82b",
		"19.03.11": "0f4336378f61ed73ed55a356ac19e46699a995f2aff34323ba5874d131548b9e",
		"19.03.12": "88de1b87b8a2582fe827154899475a72fb707c5793cfb39d2a24813ba1f31197",
		"19.03.13": "ddb13aff1fcdcceb710bf71a210169b9c1abfd7420eeaf42cf7975f8fae2fcc8",
	}

	return hashes
}

func findAllDockerHashesArm64() map[string]string {
	hashes := map[string]string{
		"17.09.0":  "2af5d112ab514d9b0b84d9e7360a5e7633e88b7168d1bbfc16c6532535cb0123",
		"17.09.1":  "a254b0f7d6fb32c786e272f7b042e010e9cae8e168ad34f7f7cca146f07e03e9",
		"17.12.0":  "b740a4475205ba8a0eb74262171be91f5a18f75554d5922d8247bf40e551f013",
		"17.12.1":  "79ec237e1ae7e2194aa13908b37fd8ccddaf2f2039d26a0be1a7bbd5d4ea3dff",
		"18.03.0":  "096522d1c9979dab76458bb443e9266d9f77c7c725fe6cffe3de31aca19c08f9",
		"18.03.1":  "483a25771d859a492ff253070471c75e062c1b43e5c3a4961fe1ac508e1ffe2c",
		"18.06.0":  "3cb454a5a5d999dff2daac0bb5d060c1fb9cf7beab3327a44e446b09f14cca58",
		"18.06.1":  "57582655ee7fe05913ffa347518c82f126321e7d71945bb6440d6f059e21528c",
		"18.06.2":  "6e7875fef7e96146c4f8994fcc24be028ec72f9f8f9ee2a832b3972dbc51d406",
		"18.06.3":  "defb2ccc95c0825833216c8b9e0e15baaa51bcedb3efc1f393f5352d184dead4",
		"18.09.0":  "c5e20dccb8ac02f2da30755ece2f4a29497afc274685835e7a6093f0cb813565",
		"18.09.1":  "3b94859ca5aa735292d31e32d7e5e33710da92de332b68348967565fc6d8d345",
		"18.09.2":  "458968bc8a4d4d3003097924c27fcfff0abdf8e52b7d3e9f6e09072c1dd42095",
		"18.09.3":  "b80971843aed5b0fdc478de6886499cdac7b34df059b7b46606643d1bdb64fc7",
		"18.09.4":  "177582a220a0a674ea5ebf6c770db6527d7f5cfb1c3811c6118bd2aed7fbc826",
		"18.09.5":  "549ea10709d9ed22b6435a072ea2d9dd7fc14950eb141cfbdd4653d0c12a54e2",
		"18.09.6":  "c4857639514471e2d1aa6d567880b7fc226437ede462021ed44157d4dcd11dc8",
		"18.09.7":  "961bbf1f826565e17dbcf5f89e8707ab4300139337920f8db1306dac5a9b6bb7",
		"18.09.8":  "243f74025612ca940f7c4c151f98a87f87a71da7e6fdce92794401906ddbffc8",
		"18.09.9":  "c6f4cfe1bef71c339d5127c6c79169479bcb7830c6fb0185139d32ab726e038e",
		"19.03.0":  "88bcbe5898b999d67cf158d5d57dd8e3d905a6cdbca669e696b6ff7554057d21",
		"19.03.1":  "44158b9fe44e8b5d3c1226a5d880425850d6f8ec383e4cf053f401e1a8fc269d",
		"19.03.2":  "3bd1bbd2e2eebf0083d684f02217e9215c1475f4ffecd28f824efc2e8585d670",
		"19.03.3":  "d6abb961d5c71a9a15b067de796c581f6ae8ee79044a6d98d529912095853ea7",
		"19.03.4":  "03c10ddd2862234d47c434455af7d27979c91448445ed3658cf383455b56e1a2",
		"19.03.5":  "0deddac5ae6f18ff0e6ce6a143c2cfd99c56dfb58be507770d840240fc9c51a9",
		"19.03.6":  "3840f64aad8f4d63851ef2d3401eb08471f8a46fb13382ae0d49913eac196f1f",
		"19.03.7":  "730058a50102dbf9278e5d164c385a437016c8c8201d6d19195d9321d0a70ec9",
		"19.03.8":  "b19da81ca82123aa0b9191c1f19c0c2632cc50d5f8c2cdb04e5b5976e3268b3b",
		"19.03.9":  "5d6ede3368eac8e74ead70489aa7e4e663fe1ccfbb9763a6ac55991d55b70354",
		"19.03.10": "c949aef8c40beea732ec497d27b8d203799ee0f34b0d330c7001d57601e5c34d",
		"19.03.11": "9cd49fe82f6b7ec413b04daef35bc0c87b01d6da67611e5beef36291538d3145",
		"19.03.12": "bc7810d58e32360652abfddc9cb43405feee4ed9592aedc1132fb35eede9fa9e",
		"19.03.13": "bdf080af7d6f383ad80e415e9c1952a63c7038c149dc673b7598bfca4d3311ec",
	}

	return hashes
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
