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

package nodemodel

import (
	"slices"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

func buildTestClusterAndIG(kubeletSpec *kops.KubeletConfigSpec) (*kops.Cluster, *kops.InstanceGroup) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			KubernetesVersion: "1.32.0",
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
			Containerd: &kops.ContainerdConfig{
				SkipInstall: true,
			},
			Networking: kops.NetworkingSpec{
				Kubenet: &kops.KubenetNetworkingSpec{},
			},
		},
	}
	ig := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			Role:    kops.InstanceGroupRoleNode,
			Kubelet: kubeletSpec,
		},
	}
	return cluster, ig
}

func TestBuildKubernetesFileAssets_DefaultKubelet(t *testing.T) {
	cluster, ig := buildTestClusterAndIG(nil)
	igModel, err := model.ForInstanceGroup(cluster, ig)
	if err != nil {
		t.Fatalf("ForInstanceGroup failed: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(vfs.Context, nil, false)
	fileAssets, err := BuildKubernetesFileAssets(igModel, assetBuilder)
	if err != nil {
		t.Fatalf("BuildKubernetesFileAssets failed: %v", err)
	}

	var amd64URLs []string
	for _, a := range fileAssets.KubernetesFileAssets[architectures.ArchitectureAmd64] {
		amd64URLs = append(amd64URLs, a.Locations...)
	}

	wantKubelet := "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubelet"
	wantKubectl := "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubectl"
	if !slices.Contains(amd64URLs, wantKubelet) {
		t.Errorf("expected amd64 assets %v to contain %q", amd64URLs, wantKubelet)
	}
	if !slices.Contains(amd64URLs, wantKubectl) {
		t.Errorf("expected amd64 assets %v to contain %q", amd64URLs, wantKubectl)
	}
}

func TestBuildKubernetesFileAssets_CustomPackages(t *testing.T) {
	customAmd64URL := "https://example.com/releases/v1.32.0/amd64/kubelet"
	customAmd64Hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	customArm64URL := "https://example.com/releases/v1.32.0/arm64/kubelet"
	customArm64Hash := "01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b"

	cluster, ig := buildTestClusterAndIG(&kops.KubeletConfigSpec{
		Packages: &kops.PackagesConfig{
			UrlAmd64:  &customAmd64URL,
			HashAmd64: &customAmd64Hash,
			UrlArm64:  &customArm64URL,
			HashArm64: &customArm64Hash,
		},
	})
	igModel, err := model.ForInstanceGroup(cluster, ig)
	if err != nil {
		t.Fatalf("ForInstanceGroup failed: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(vfs.Context, nil, false)
	fileAssets, err := BuildKubernetesFileAssets(igModel, assetBuilder)
	if err != nil {
		t.Fatalf("BuildKubernetesFileAssets failed: %v", err)
	}

	var amd64URLs []string
	for _, a := range fileAssets.KubernetesFileAssets[architectures.ArchitectureAmd64] {
		amd64URLs = append(amd64URLs, a.Locations...)
	}
	if !slices.Contains(amd64URLs, customAmd64URL) {
		t.Errorf("expected amd64 assets %v to contain %q", amd64URLs, customAmd64URL)
	}

	var arm64URLs []string
	for _, a := range fileAssets.KubernetesFileAssets[architectures.ArchitectureArm64] {
		arm64URLs = append(arm64URLs, a.Locations...)
	}
	if !slices.Contains(arm64URLs, customArm64URL) {
		t.Errorf("expected arm64 assets %v to contain %q", arm64URLs, customArm64URL)
	}
}
