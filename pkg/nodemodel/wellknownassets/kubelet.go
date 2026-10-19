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
	"net/url"
	"path"

	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
)

func FindKubeletAsset(ig model.InstanceGroup, assetBuilder *assets.AssetBuilder, arch architectures.Architecture, baseURL string) (*assets.FileAsset, error) {
	canonicalURL := ""
	knownHash := ""

	kubelet := ig.KubeletSpec()
	if kubelet != nil && kubelet.Packages != nil {
		if arch == architectures.ArchitectureAmd64 && kubelet.Packages.UrlAmd64 != nil && kubelet.Packages.HashAmd64 != nil {
			canonicalURL = fi.ValueOf(kubelet.Packages.UrlAmd64)
			knownHash = fi.ValueOf(kubelet.Packages.HashAmd64)
		}
		if arch == architectures.ArchitectureArm64 && kubelet.Packages.UrlArm64 != nil && kubelet.Packages.HashArm64 != nil {
			canonicalURL = fi.ValueOf(kubelet.Packages.UrlArm64)
			knownHash = fi.ValueOf(kubelet.Packages.HashArm64)
		}
	}

	if canonicalURL == "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		u.Path = path.Join(u.Path, fmt.Sprintf("/bin/linux/%s/kubelet", arch))
		canonicalURL = u.String()
	}

	return buildFileAsset(assetBuilder, canonicalURL, knownHash)
}
