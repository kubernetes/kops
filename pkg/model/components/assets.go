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
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// AssetsOptionsBuilder defaults the assets locations when assets.managed is enabled.
type AssetsOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.ClusterOptionsBuilder = &AssetsOptionsBuilder{}

// BuildOptions fills in the assets locations for a kOps-managed registry.
func (b *AssetsOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	assets := o.Spec.Assets
	if assets == nil || !fi.ValueOf(assets.Managed) {
		return nil
	}

	if o.GetCloudProvider() == kops.CloudProviderAzure {
		if assets.FileRepository == nil {
			registryName := azureManagedRegistryName(o.Spec.CloudProvider.Azure.SubscriptionID, o.ObjectMeta.Name)
			assets.FileRepository = new(fmt.Sprintf("oci://%s.azurecr.io/assets", registryName))
		}
		if assets.ContainerRegistry == nil {
			if host := o.Spec.OCIAssetRegistryHost(); host != "" {
				assets.ContainerRegistry = new(host)
			}
		}
	}

	// The AssetBuilder was created from the cluster spec before the locations
	// were defaulted; asset remapping must see the defaulted locations.
	if b.Context != nil && b.Context.AssetBuilder != nil {
		b.Context.AssetBuilder.SetAssetsLocation(assets)
	}

	return nil
}

// azureManagedRegistryName derives the name of the kOps-managed container registry.
// Registry names are global, and cluster names (e.g. "my.k8s") are not unique across
// users, so the name is derived from both the subscription ID and the cluster name.
func azureManagedRegistryName(subscriptionID, clusterName string) string {
	hash := sha256.Sum256([]byte(subscriptionID + "/" + clusterName))
	return "kops" + hex.EncodeToString(hash[:8])
}
