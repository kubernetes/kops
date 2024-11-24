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
	"context"
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

// KubernetesFileAssets are the assets for downloading Kubernetes binaries
type KubernetesFileAssets struct {
	// KubernetesFileAssets are the assets for downloading Kubernetes binaries
	KubernetesFileAssets map[architectures.Architecture][]*assets.MirroredAsset
}

// BuildKubernetesFileAssets returns the Kubernetes file assets for the given cluster
func BuildKubernetesFileAssets(cluster *kops.Cluster, assetBuilder *assets.AssetBuilder) (*KubernetesFileAssets, error) {
	var baseURL string
	if components.IsBaseURL(cluster.Spec.KubernetesVersion) {
		baseURL = cluster.Spec.KubernetesVersion
	} else {
		baseURL = "https://dl.k8s.io/release/v" + cluster.Spec.KubernetesVersion
	}

	kubernetesAssets := make(map[architectures.Architecture][]*assets.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		kubernetesAssets[arch] = []*assets.MirroredAsset{}

		k8sAssetsNames := []string{
			fmt.Sprintf("/bin/linux/%s/kubelet", arch),
			fmt.Sprintf("/bin/linux/%s/kubectl", arch),
		}

		if needsMounterAsset(cluster) {
			k8sAssetsNames = append(k8sAssetsNames, fmt.Sprintf("/bin/linux/%s/mounter", arch))
		}

		for _, an := range k8sAssetsNames {
			k, err := url.Parse(baseURL)
			if err != nil {
				return nil, err
			}
			k.Path = path.Join(k.Path, an)

			asset, err := assetBuilder.RemapFile(k, nil)
			if err != nil {
				return nil, err
			}
			kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(asset))
		}

		kubernetesVersion, _ := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)

		cloudProvider := cluster.GetCloudProvider()
		if ok := model.UseExternalKubeletCredentialProvider(*kubernetesVersion, cloudProvider); ok {
			switch cloudProvider {
			case kops.CloudProviderGCE:
				binaryLocation := cluster.Spec.CloudProvider.GCE.BinariesLocation
				if binaryLocation == nil {
					binaryLocation = fi.PtrTo("https://storage.googleapis.com/k8s-staging-cloud-provider-gcp/auth-provider-gcp")
				}
				// VALID FOR 60 DAYS WE REALLY NEED TO MERGE https://github.com/kubernetes/cloud-provider-gcp/pull/601 and CUT A RELEASE
				k, err := url.Parse(fmt.Sprintf("%s/linux-%s/v20231005-providersv0.27.1-65-g8fbe8d27", *binaryLocation, arch))
				if err != nil {
					return nil, err
				}

				// TODO: Move these hashes to assetdata
				hashes := map[architectures.Architecture]string{
					"amd64": "827d558953d861b81a35c3b599191a73f53c1f63bce42c61e7a3fee21a717a89",
					"arm64": "f1617c0ef77f3718e12a3efc6f650375d5b5e96eebdbcbad3e465e89e781bdfa",
				}
				hash, err := hashing.FromString(hashes[arch])
				if err != nil {
					return nil, fmt.Errorf("unable to parse auth-provider-gcp binary asset hash %q: %v", hashes[arch], err)
				}
				asset, err := assetBuilder.RemapFile(k, hash)
				if err != nil {
					return nil, err
				}

				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(asset))
			case kops.CloudProviderAWS:
				binaryLocation := cluster.Spec.CloudProvider.AWS.BinariesLocation
				if binaryLocation == nil {
					binaryLocation = fi.PtrTo("https://artifacts.k8s.io/binaries/cloud-provider-aws/v1.27.1")
				}

				u, err := url.Parse(fmt.Sprintf("%s/linux/%s/ecr-credential-provider-linux-%s", *binaryLocation, arch, arch))
				if err != nil {
					return nil, err
				}
				asset, err := assetBuilder.RemapFile(u, nil)
				if err != nil {
					return nil, err
				}
				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(asset))
			}
		}

		{
			cniAsset, err := wellknownassets.FindCNIAssets(cluster, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(cniAsset))
		}

		if cluster.Spec.Containerd == nil || !cluster.Spec.Containerd.SkipInstall {
			containerdAsset, err := wellknownassets.FindContainerdAsset(cluster, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			if containerdAsset != nil {
				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(containerdAsset))
			}

			runcAsset, err := wellknownassets.FindRuncAsset(cluster, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			if runcAsset != nil {
				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(runcAsset))
			}
			nerdctlAsset, err := wellknownassets.FindNerdctlAsset(cluster, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			if nerdctlAsset != nil {
				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(nerdctlAsset))
			}
		}

		crictlAsset, err := wellknownassets.FindCrictlAsset(cluster, assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		if crictlAsset != nil {
			kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(crictlAsset))
		}

	}

	return &KubernetesFileAssets{
		KubernetesFileAssets: kubernetesAssets,
	}, nil
}

// NodeUpAssets are the assets for downloading nodeup
type NodeUpAssets struct {
	// NodeUpAssets are the assets for downloading nodeup
	NodeUpAssets map[architectures.Architecture]*assets.MirroredAsset
}

func BuildNodeUpAssets(ctx context.Context, assetBuilder *assets.AssetBuilder) (*NodeUpAssets, error) {
	nodeUpAssets := make(map[architectures.Architecture]*assets.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		asset, err := wellknownassets.NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		nodeUpAssets[arch] = asset
	}
	return &NodeUpAssets{
		NodeUpAssets: nodeUpAssets,
	}, nil
}

// needsMounterAsset checks if we need the mounter program
// This is only needed currently on ContainerOS i.e. GCE, but we don't have a nice way to detect it yet
func needsMounterAsset(c *kops.Cluster) bool {
	// TODO: Do real detection of ContainerOS (but this has to work with image names, and maybe even forked images)
	switch c.GetCloudProvider() {
	case kops.CloudProviderGCE:
		return true
	default:
		return false
	}
}
