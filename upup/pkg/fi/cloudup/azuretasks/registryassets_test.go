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
	"reflect"
	"sort"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestRegistryAssetsDependencies(t *testing.T) {
	registry := newTestContainerRegistry()
	registryAssets := &RegistryAssets{
		Name:      to.Ptr("registry-assets"),
		Lifecycle: fi.LifecycleSync,
		Registry:  registry,
	}
	registryAssets.SetPush(func() error { return nil })

	// The role assignment granting pull access can only be created once the registry
	// exists. Reference the registry task directly, as the loader would have resolved
	// the reference to the primary task.
	roleAssignment := &RoleAssignment{
		Name:              to.Ptr("acrpull"),
		Lifecycle:         fi.LifecycleSync,
		Scope:             to.Ptr("scope"),
		RoleDefID:         to.Ptr("role"),
		ContainerRegistry: registry,
	}

	// Dependency analysis reflects over task fields; it must handle the tasks,
	// including the injected push function, and order them after the registry.
	// The push must also run after the bootstrap configs have been built, which
	// is when the full list of assets is known.
	bootstrapScript := &ResourceGroup{Name: to.Ptr("bootstrap-script-stand-in")}
	managedFile := &ResourceGroup{Name: to.Ptr("addon-manifest-stand-in")}
	tasks := map[string]fi.CloudupTask{
		"ResourceGroup/rg":               registry.ResourceGroup,
		"ContainerRegistry/registry":     registry,
		"RegistryAssets/registry-assets": registryAssets,
		"RoleAssignment/acrpull":         roleAssignment,
		"BootstrapScript/nodes":          bootstrapScript,
		"AddonManifest/addons":           managedFile,
	}
	dependencies := fi.FindTaskDependencies(tasks)
	registryAssetsDependencies := dependencies["RegistryAssets/registry-assets"]
	sort.Strings(registryAssetsDependencies)
	if a, e := registryAssetsDependencies, []string{"AddonManifest/addons", "BootstrapScript/nodes", "ContainerRegistry/registry"}; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected dependencies: expected %v, but got %v", e, a)
	}
	if a, e := dependencies["RoleAssignment/acrpull"], []string{"ContainerRegistry/registry"}; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected role assignment dependencies: expected %v, but got %v", e, a)
	}
}

func TestRegistryAssetsRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
		Target: azure.NewAzureAPITarget(cloud),
	}

	pushed := false
	registryAssets := &RegistryAssets{
		Name:      to.Ptr("registry-assets"),
		Lifecycle: fi.LifecycleSync,
		Registry:  newTestContainerRegistry(),
	}
	registryAssets.SetPush(func() error {
		pushed = true
		return nil
	})

	if err := registryAssets.Run(ctx); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !pushed {
		t.Error("expected the assets to be pushed")
	}
}
