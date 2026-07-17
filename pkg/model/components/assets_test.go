/*
Copyright 2026 The Kubernetes Authors.

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

package components

import (
	"net/url"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

func TestAssetsOptionsBuilder(t *testing.T) {
	newAzureCluster := func(assets *kops.AssetsSpec) *kops.Cluster {
		return &kops.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "my.k8s"},
			Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Azure: &kops.AzureSpec{
						SubscriptionID: "00000000-0000-0000-0000-000000000001",
					},
				},
				Assets: assets,
			},
		}
	}

	builder := &AssetsOptionsBuilder{}

	// managed defaults both locations
	cluster := newAzureCluster(&kops.AssetsSpec{Managed: new(true)})
	if err := builder.BuildOptions(cluster); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fileRepository := fi.ValueOf(cluster.Spec.Assets.FileRepository)
	if !strings.HasPrefix(fileRepository, "oci://kops") || !strings.HasSuffix(fileRepository, ".azurecr.io/assets") {
		t.Errorf("unexpected fileRepository: %q", fileRepository)
	}
	registryHost := cluster.Spec.OCIAssetRegistryHost()
	if a := fi.ValueOf(cluster.Spec.Assets.ContainerRegistry); a != registryHost {
		t.Errorf("unexpected containerRegistry: expected %q, but got %q", registryHost, a)
	}
	registryName := strings.TrimSuffix(registryHost, ".azurecr.io")
	if len(registryName) < 5 || len(registryName) > 50 {
		t.Errorf("registry name %q length out of the 5-50 range", registryName)
	}

	// the derived name must differ for the same cluster name in another subscription
	otherSubscription := newAzureCluster(&kops.AssetsSpec{Managed: new(true)})
	otherSubscription.Spec.CloudProvider.Azure.SubscriptionID = "00000000-0000-0000-0000-000000000002"
	if err := builder.BuildOptions(otherSubscription); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a := fi.ValueOf(otherSubscription.Spec.Assets.FileRepository); a == fileRepository {
		t.Errorf("expected a different fileRepository for a different subscription, got %q for both", a)
	}

	// user-provided locations are not overwritten
	cluster = newAzureCluster(&kops.AssetsSpec{
		Managed:        new(true),
		FileRepository: new("oci://myregistry.azurecr.io/assets"),
	})
	if err := builder.BuildOptions(cluster); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a, e := fi.ValueOf(cluster.Spec.Assets.FileRepository), "oci://myregistry.azurecr.io/assets"; a != e {
		t.Errorf("unexpected fileRepository: expected %q, but got %q", e, a)
	}
	if a, e := fi.ValueOf(cluster.Spec.Assets.ContainerRegistry), "myregistry.azurecr.io"; a != e {
		t.Errorf("unexpected containerRegistry: expected %q, but got %q", e, a)
	}

	// without managed, nothing is defaulted
	cluster = newAzureCluster(&kops.AssetsSpec{})
	if err := builder.BuildOptions(cluster); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cluster.Spec.Assets.FileRepository != nil || cluster.Spec.Assets.ContainerRegistry != nil {
		t.Errorf("unexpected defaulting without managed: %+v", cluster.Spec.Assets)
	}
}

func TestAssetsOptionsBuilderUpdatesAssetBuilder(t *testing.T) {
	// The AssetBuilder is created from the cluster spec before defaulting has run;
	// assets remapped after defaulting must use the defaulted locations.
	cluster := &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "my.k8s"},
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Azure: &kops.AzureSpec{
					SubscriptionID: "00000000-0000-0000-0000-000000000001",
				},
			},
			Assets: &kops.AssetsSpec{Managed: new(true)},
		},
	}

	assetBuilder := assets.NewAssetBuilder(vfs.NewVFSContext(), cluster.Spec.Assets, false)
	cluster = cluster.DeepCopy() // cluster completion works on a copy of the cluster

	builder := &AssetsOptionsBuilder{
		Context: &OptionsContext{AssetBuilder: assetBuilder},
	}
	if err := builder.BuildOptions(cluster); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	canonicalURL, err := url.Parse("https://artifacts.k8s.io/binaries/kops/1.35.0/linux/amd64/nodeup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash, err := hashing.FromString("833723369ad345a88dd85d61b1e77336d56e61b864557ded71b92b6e34158e6a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	asset, err := assetBuilder.RemapFile(canonicalURL, hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a, e := asset.DownloadURL.String(), fi.ValueOf(cluster.Spec.Assets.FileRepository)+"/binaries/kops/1.35.0/linux/amd64/nodeup"; a != e {
		t.Errorf("unexpected remapped asset URL: expected %q, but got %q", e, a)
	}
}
