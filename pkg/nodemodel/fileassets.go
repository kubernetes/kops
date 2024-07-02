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

package nodemodel

import (
	"fmt"
	"net/url"
	"path"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/nodemodel/wellknownassets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

type FileAssets struct {
	// Assets is a list of sources for files (primarily when not using everything containerized)
	// Formats:
	//  raw url: http://... or https://...
	//  url with hash: <hex>@http://... or <hex>@https://...
	Assets map[architectures.Architecture][]*assets.MirroredAsset

	// NodeUpAssets are the assets for downloading nodeup
	NodeUpAssets map[architectures.Architecture]*assets.MirroredAsset

	Cluster *kops.Cluster
}

// AddFileAssets adds the file assets within the assetBuilder
func (c *FileAssets) AddFileAssets(assetBuilder *assets.AssetBuilder) error {
	var baseURL string
	if components.IsBaseURL(c.Cluster.Spec.KubernetesVersion) {
		baseURL = c.Cluster.Spec.KubernetesVersion
	} else {
		baseURL = "https://dl.k8s.io/release/v" + c.Cluster.Spec.KubernetesVersion
	}

	c.Assets = make(map[architectures.Architecture][]*assets.MirroredAsset)
	c.NodeUpAssets = make(map[architectures.Architecture]*assets.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		c.Assets[arch] = []*assets.MirroredAsset{}

		k8sAssetsNames := []string{
			fmt.Sprintf("/bin/linux/%s/kubelet", arch),
			fmt.Sprintf("/bin/linux/%s/kubectl", arch),
		}

		if needsMounterAsset(c.Cluster) {
			k8sAssetsNames = append(k8sAssetsNames, fmt.Sprintf("/bin/linux/%s/mounter", arch))
		}

		for _, an := range k8sAssetsNames {
			k, err := url.Parse(baseURL)
			if err != nil {
				return err
			}
			k.Path = path.Join(k.Path, an)

			asset, err := assetBuilder.RemapFile(k, nil)
			if err != nil {
				return err
			}
			c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(asset))
		}

		kubernetesVersion, _ := util.ParseKubernetesVersion(c.Cluster.Spec.KubernetesVersion)

		cloudProvider := c.Cluster.Spec.GetCloudProvider()
		if ok := model.UseExternalKubeletCredentialProvider(*kubernetesVersion, cloudProvider); ok {
			switch cloudProvider {
			case kops.CloudProviderGCE:
				binaryLocation := c.Cluster.Spec.CloudProvider.GCE.BinariesLocation
				if binaryLocation == nil {
					binaryLocation = fi.PtrTo("https://storage.googleapis.com/k8s-staging-cloud-provider-gcp/auth-provider-gcp")
				}
				// VALID FOR 60 DAYS WE REALLY NEED TO MERGE https://github.com/kubernetes/cloud-provider-gcp/pull/601 and CUT A RELEASE
				k, err := url.Parse(fmt.Sprintf("%s/linux-%s/v20231005-providersv0.27.1-65-g8fbe8d27", *binaryLocation, arch))
				if err != nil {
					return err
				}

				// TODO: Move these hashes to assetdata
				hashes := map[architectures.Architecture]string{
					"amd64": "827d558953d861b81a35c3b599191a73f53c1f63bce42c61e7a3fee21a717a89",
					"arm64": "f1617c0ef77f3718e12a3efc6f650375d5b5e96eebdbcbad3e465e89e781bdfa",
				}
				hash, err := hashing.FromString(hashes[arch])
				if err != nil {
					return fmt.Errorf("unable to parse auth-provider-gcp binary asset hash %q: %v", hashes[arch], err)
				}
				asset, err := assetBuilder.RemapFile(k, hash)
				if err != nil {
					return err
				}

				c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(asset))
			case kops.CloudProviderAWS:
				binaryLocation := c.Cluster.Spec.CloudProvider.AWS.BinariesLocation
				if binaryLocation == nil {
					binaryLocation = fi.PtrTo("https://artifacts.k8s.io/binaries/cloud-provider-aws/v1.29.0")
				}

				u, err := url.Parse(fmt.Sprintf("%s/linux/%s/ecr-credential-provider-linux-%s", *binaryLocation, arch, arch))
				if err != nil {
					return err
				}
				asset, err := assetBuilder.RemapFile(u, nil)
				if err != nil {
					return err
				}
				c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(asset))
			}
		}

		{
			cniAsset, err := wellknownassets.FindCNIAssets(c.Cluster, assetBuilder, arch)
			if err != nil {
				return err
			}
			c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(cniAsset))
		}

		if c.Cluster.Spec.Containerd == nil || !c.Cluster.Spec.Containerd.SkipInstall {
			containerdAsset, err := wellknownassets.FindContainerdAsset(c.Cluster, assetBuilder, arch)
			if err != nil {
				return err
			}
			if containerdAsset != nil {
				c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(containerdAsset))
			}

			runcAsset, err := wellknownassets.FindRuncAsset(c.Cluster, assetBuilder, arch)
			if err != nil {
				return err
			}
			if runcAsset != nil {
				c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(runcAsset))
			}
			nerdctlAsset, err := wellknownassets.FindNerdctlAsset(c.Cluster, assetBuilder, arch)
			if err != nil {
				return err
			}
			if nerdctlAsset != nil {
				c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(nerdctlAsset))
			}
		}

		crictlAsset, err := wellknownassets.FindCrictlAsset(c.Cluster, assetBuilder, arch)
		if err != nil {
			return err
		}
		if crictlAsset != nil {
			c.Assets[arch] = append(c.Assets[arch], assets.BuildMirroredAsset(crictlAsset))
		}

		asset, err := wellknownassets.NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return err
		}
		c.NodeUpAssets[arch] = asset
	}

	return nil
}

// needsMounterAsset checks if we need the mounter program
// This is only needed currently on ContainerOS i.e. GCE, but we don't have a nice way to detect it yet
func needsMounterAsset(c *kops.Cluster) bool {
	// TODO: Do real detection of ContainerOS (but this has to work with image names, and maybe even forked images)
	switch c.Spec.GetCloudProvider() {
	case kops.CloudProviderGCE:
		return true
	default:
		return false
	}
}
