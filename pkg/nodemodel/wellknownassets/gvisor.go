/*
Copyright The Kubernetes Authors.

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

	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
)

// The gVisor release tarball holds runsc, containerd-shim-runsc-v1 and the
// gvisor-bin/ directory of sidecar binaries that runsc runs from next to itself.
// gVisor publishes it only as .tar.zstd (or .tar.bz2), with a sha512 checksum
// next to it: <url>.sha512.
const (
	gVisorAssetUrlAmd64  = "https://storage.googleapis.com/gvisor/releases/release/20261109.0/x86_64/gvisor.tar.zstd"
	gVisorAssetUrlArm64  = "https://storage.googleapis.com/gvisor/releases/release/20261109.0/aarch64/gvisor.tar.zstd"
	gVisorAssetHashAmd64 = "0a0041fd862f4ea3042a9db8d41c742f48a01f909699adc2a0256e46d85e53f97b82b9fd154b3dce636f62b8d557677639f91c77d13f3f875ff9fa535275393c"
	gVisorAssetHashArm64 = "a0b5d911512ef5fe70e833821e653a90c4d06c7027246ee49423f1d2fe2b2a509389a637797248958b90e49a069bb079ecde860efe1871b1d980ec0fe17e18ac"
)

// FindGVisorAsset returns the gVisor release tarball for the given architecture.
func FindGVisorAsset(ig model.InstanceGroup, assetBuilder *assets.AssetBuilder, arch architectures.Architecture) (*assets.FileAsset, error) {
	var assetURL, assetHash string
	switch arch {
	case architectures.ArchitectureAmd64:
		assetURL = gVisorAssetUrlAmd64
		assetHash = gVisorAssetHashAmd64
	case architectures.ArchitectureArm64:
		assetURL = gVisorAssetUrlArm64
		assetHash = gVisorAssetHashArm64
	default:
		return nil, fmt.Errorf("unknown arch for gVisor asset: %s", arch)
	}

	return buildFileAsset(assetBuilder, assetURL, assetHash)
}
