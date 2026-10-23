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
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/nodemodel/wellknownassets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

// KubernetesFileAssets are the assets for downloading Kubernetes binaries
type KubernetesFileAssets struct {
	// KubernetesFileAssets are the assets for downloading Kubernetes binaries
	KubernetesFileAssets map[architectures.Architecture][]*assets.MirroredAsset
}

// BuildKubernetesFileAssets returns the Kubernetes file assets for the given cluster
func BuildKubernetesFileAssets(ig model.InstanceGroup, assetBuilder *assets.AssetBuilder) (*KubernetesFileAssets, error) {
	kubernetesVersion := ig.KubernetesVersion()

	var baseURL string
	if kubernetesVersion.IsBaseURL() {
		baseURL = kubernetesVersion.String()
	} else {
		baseURL = "https://dl.k8s.io/release/v" + kubernetesVersion.String()
	}

	kubernetesAssets := make(map[architectures.Architecture][]*assets.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		kubernetesAssets[arch] = []*assets.MirroredAsset{}

		kubeletAsset, err := buildKubeletAsset(ig, assetBuilder, baseURL, arch)
		if err != nil {
			return nil, err
		}
		if kubeletAsset != nil {
			kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(kubeletAsset))
		}

		kubectlURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		kubectlURL.Path = path.Join(kubectlURL.Path, fmt.Sprintf("/bin/linux/%s/kubectl", arch))
		kubectlAsset, err := assetBuilder.RemapFile(kubectlURL, nil)
		if err != nil {
			return nil, err
		}
		kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(kubectlAsset))

		cloudProvider := ig.GetCloudProvider()
		switch cloudProvider {
		case kops.CloudProviderGCE:
			binaryLocation := ig.RawClusterSpec().CloudProvider.GCE.BinariesLocation
			if binaryLocation == nil {
				binaryLocation = new("https://artifacts.k8s.io/binaries/cloud-provider-gcp/v35.0.0")
			}

			u, err := url.Parse(fmt.Sprintf("%s/auth-provider-gcp/linux/%s/auth-provider-gcp", *binaryLocation, arch))
			if err != nil {
				return nil, err
			}
			asset, err := assetBuilder.RemapFile(u, nil)
			if err != nil {
				return nil, err
			}

			kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(asset))
		case kops.CloudProviderAWS:
			binaryLocation := ig.RawClusterSpec().CloudProvider.AWS.BinariesLocation
			if binaryLocation == nil {
				binaryLocation = new("https://artifacts.k8s.io/binaries/cloud-provider-aws/v1.37.0")
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

		if ig.InstallCNIAssets() {
			cniAsset, err := wellknownassets.FindCNIAssets(ig, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(cniAsset))
		}

		if ig.RawClusterSpec().Containerd == nil || !ig.RawClusterSpec().Containerd.SkipInstall {
			containerdAsset, err := wellknownassets.FindContainerdAsset(ig, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			if containerdAsset != nil {
				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(containerdAsset))
			}

			runcAsset, err := wellknownassets.FindRuncAsset(ig, assetBuilder, arch)
			if err != nil {
				return nil, err
			}
			if runcAsset != nil {
				kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(runcAsset))
			}
			if ig.RawClusterSpec().Containerd.InstallNerdCtl {
				nerdctlAsset, err := wellknownassets.FindNerdctlAsset(ig, assetBuilder, arch)
				if err != nil {
					return nil, err
				}
				if nerdctlAsset != nil {
					kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(nerdctlAsset))
				}
			}
			if ig.RawClusterSpec().Containerd.InstallCriCtl {
				crictlAsset, err := wellknownassets.FindCrictlAsset(ig, assetBuilder, arch)
				if err != nil {
					return nil, err
				}
				if crictlAsset != nil {
					kubernetesAssets[arch] = append(kubernetesAssets[arch], assets.BuildMirroredAsset(crictlAsset))
				}
			}
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

func buildKubeletAsset(ig model.InstanceGroup, assetBuilder *assets.AssetBuilder, baseURL string, arch architectures.Architecture) (*assets.FileAsset, error) {
	kubeletSpec := ig.KubeletSpec()
	if kubeletSpec != nil && kubeletSpec.Packages != nil {
		var canonicalURL string
		var knownHashString string
		switch arch {
		case architectures.ArchitectureAmd64:
			if kubeletSpec.Packages.UrlAmd64 != nil && kubeletSpec.Packages.HashAmd64 != nil {
				canonicalURL = *kubeletSpec.Packages.UrlAmd64
				knownHashString = *kubeletSpec.Packages.HashAmd64
			}
		case architectures.ArchitectureArm64:
			if kubeletSpec.Packages.UrlArm64 != nil && kubeletSpec.Packages.HashArm64 != nil {
				canonicalURL = *kubeletSpec.Packages.UrlArm64
				knownHashString = *kubeletSpec.Packages.HashArm64
			}
		}
		if canonicalURL != "" {
			u, err := url.Parse(canonicalURL)
			if err != nil {
				return nil, fmt.Errorf("unable to parse kubelet package URL %q: %w", canonicalURL, err)
			}
			knownHash, err := hashing.FromString(knownHashString)
			if err != nil {
				return nil, fmt.Errorf("unable to parse kubelet package hash %q: %w", knownHashString, err)
			}
			return assetBuilder.RemapFile(u, knownHash)
		}
	}

	k, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	k.Path = path.Join(k.Path, fmt.Sprintf("/bin/linux/%s/kubelet", arch))
	return assetBuilder.RemapFile(k, nil)
}
