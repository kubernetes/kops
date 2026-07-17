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

package azuremodel

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
	"k8s.io/kops/util/pkg/vfs"
)

// PushAssetsFunc pushes the cluster's image and file assets to a private registry.
type PushAssetsFunc func(imageAssets []*assets.ImageAsset, fileAssets []*assets.FileAsset, vfsContext *vfs.VFSContext, cluster *kops.Cluster) error

// ContainerRegistryModelBuilder configures the Azure Container Registry that holds
// the cluster's assets when assets.managed is enabled.
type ContainerRegistryModelBuilder struct {
	*AzureModelContext
	Lifecycle fi.Lifecycle

	// AssetBuilder collects the cluster's image and file assets as the model is built.
	AssetBuilder *assets.AssetBuilder
	// PushAssets is injected by the kops CLI so that only the CLI links the
	// container registry client libraries.
	PushAssets PushAssetsFunc
}

var _ fi.CloudupModelBuilder = &ContainerRegistryModelBuilder{}

// Build is responsible for constructing the Container Registry from the kops spec.
func (b *ContainerRegistryModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	registryName := b.Cluster.AzureManagedContainerRegistryName()
	if registryName == "" {
		return nil
	}

	registry := &azuretasks.ContainerRegistry{
		Name:          to.Ptr(registryName),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
	}
	c.AddTask(registry)

	registryAssets := &azuretasks.RegistryAssets{
		Name:      to.Ptr(registryName + "-assets"),
		Lifecycle: b.Lifecycle,
		Registry:  registry,
	}
	if b.PushAssets != nil {
		// Snapshot the assets lazily; the full asset list is only known once all
		// model builders have run.
		registryAssets.SetPush(func() error {
			return b.PushAssets(b.AssetBuilder.ImageAssets(), b.AssetBuilder.FileAssets(), b.AssetBuilder.VFSContext(), b.Cluster)
		})
	}
	c.AddTask(registryAssets)

	return nil
}
