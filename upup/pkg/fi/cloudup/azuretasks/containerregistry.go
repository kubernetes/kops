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

package azuretasks

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	containerregistry "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// ContainerRegistry is an Azure Container Registry.
// +kops:fitask
type ContainerRegistry struct {
	Name          *string
	Lifecycle     fi.Lifecycle
	ResourceGroup *ResourceGroup
	Tags          map[string]*string
}

var (
	_ fi.CloudupTask          = &ContainerRegistry{}
	_ fi.CompareWithID        = &ContainerRegistry{}
	_ fi.CloudupTaskNormalize = &ContainerRegistry{}
)

// CompareWithID returns the Name of the Container Registry.
func (r *ContainerRegistry) CompareWithID() *string {
	return r.Name
}

// Find discovers the ContainerRegistry in the cloud provider.
func (r *ContainerRegistry) Find(c *fi.CloudupContext) (*ContainerRegistry, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	l, err := cloud.ContainerRegistry().List(context.TODO(), *r.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *containerregistry.Registry
	for _, registry := range l {
		if *registry.Name == *r.Name {
			found = registry
			break
		}
	}
	if found == nil {
		return nil, nil
	}
	return &ContainerRegistry{
		Name:      to.Ptr(*found.Name),
		Lifecycle: r.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: r.ResourceGroup.Name,
		},
		Tags: found.Tags,
	}, nil
}

func (r *ContainerRegistry) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(r.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (r *ContainerRegistry) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(r, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*ContainerRegistry) CheckChanges(a, e, changes *ContainerRegistry) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchangeable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a Container Registry.
func (*ContainerRegistry) RenderAzure(t *azure.AzureAPITarget, a, e, changes *ContainerRegistry) error {
	if a == nil {
		klog.Infof("Creating a new Container Registry with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Container Registry with name: %s", fi.ValueOf(e.Name))
	}
	_, err := t.Cloud.ContainerRegistry().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		containerregistry.Registry{
			Location: to.Ptr(t.Cloud.Region()),
			SKU: &containerregistry.SKU{
				Name: to.Ptr(containerregistry.SKUNameBasic),
			},
			Properties: &containerregistry.RegistryProperties{
				AdminUserEnabled: to.Ptr(false),
			},
			Tags: e.Tags,
		})
	return err
}
