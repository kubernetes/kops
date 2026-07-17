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
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	containerregistry "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func newTestContainerRegistry() *ContainerRegistry {
	return &ContainerRegistry{
		Name:      to.Ptr("registry"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.Ptr("rg"),
		},
		Tags: map[string]*string{
			testTagKey: to.Ptr(testTagValue),
		},
	}
}

func TestContainerRegistryRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	registry := &ContainerRegistry{}
	expected := newTestContainerRegistry()
	if err := registry.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.ContainerRegistriesClient.Registries[*expected.Name]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Errorf("unexpected location: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.SKU.Name, containerregistry.SKUNameBasic; a != e {
		t.Errorf("unexpected SKU: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, expected.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}
}

func TestContainerRegistryFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	registry := newTestContainerRegistry()
	// Find will return nothing if there is no Container Registry created.
	actual, err := registry.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected container registry found: %+v", actual)
	}

	// Create a Container Registry.
	parameters := containerregistry.Registry{
		Location: to.Ptr(cloud.Location),
		SKU: &containerregistry.SKU{
			Name: to.Ptr(containerregistry.SKUNameBasic),
		},
		Tags: map[string]*string{
			"key": to.Ptr("value"),
		},
	}
	if _, err := cloud.ContainerRegistry().CreateOrUpdate(context.Background(), *registry.ResourceGroup.Name, *registry.Name, parameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = registry.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *registry.Name; a != e {
		t.Errorf("unexpected Container Registry name: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, parameters.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}

	// Call Find with an invalid registry name.
	registry = newTestContainerRegistry()
	registry.Name = to.Ptr("invalid")
	actual, err = registry.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected container registry found: %+v", actual)
	}
}

func TestContainerRegistryRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
		Target: azure.NewAzureAPITarget(cloud),
	}

	registry := newTestContainerRegistry()
	err := registry.Normalize(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = registry.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expectedTags := map[string]*string{
		azure.TagClusterName: to.Ptr(testClusterName),
		testTagKey:           to.Ptr(testTagValue),
	}
	if a := registry.Tags; !reflect.DeepEqual(a, expectedTags) {
		t.Errorf("unexpected tags: expected %+v, but got %+v", expectedTags, a)
	}
}

func TestContainerRegistryCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *ContainerRegistry
		success       bool
	}{
		{
			a:       nil,
			e:       &ContainerRegistry{Name: to.Ptr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &ContainerRegistry{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &ContainerRegistry{Name: to.Ptr("name")},
			changes: &ContainerRegistry{Name: nil},
			success: true,
		},
		{
			a:       &ContainerRegistry{Name: to.Ptr("name")},
			changes: &ContainerRegistry{Name: to.Ptr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			registry := ContainerRegistry{}
			err := registry.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
