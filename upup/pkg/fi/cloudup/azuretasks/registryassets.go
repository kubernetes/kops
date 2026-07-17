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
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// RegistryAssets pushes the cluster's file and image assets to the cluster's
// container registry, after the registry has been created and before instances
// try to pull from it.
// +kops:fitask
type RegistryAssets struct {
	Name      *string
	Lifecycle fi.Lifecycle

	// Registry is the container registry holding the assets; referencing the task
	// orders the push after the registry creation.
	Registry *ContainerRegistry

	// push pushes the assets. It is injected by the kops CLI so that only the CLI
	// links the container registry client libraries. It is unexported so that task
	// reflection (dependency analysis) ignores it.
	push func() error
}

var (
	_ fi.CloudupTask             = &RegistryAssets{}
	_ fi.CloudupHasDependencies  = &RegistryAssets{}
	_ fi.RunsAfterAddonManifests = &RegistryAssets{}
)

// SetPush injects the function that pushes the assets.
func (r *RegistryAssets) SetPush(push func() error) {
	r.push = push
}

// RunsAfterAddonManifests implements fi.RunsAfterAddonManifests.
func (r *RegistryAssets) RunsAfterAddonManifests() {}

// GetDependencies implements fi.HasDependencies. The full list of assets is only
// known once all the bootstrap configs and the addon manifests have been built,
// so the push runs after the BootstrapScript and AddonManifest tasks, in
// addition to the registry creation.
func (r *RegistryAssets) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for key, task := range tasks {
		if strings.HasPrefix(key, "BootstrapScript/") || strings.HasPrefix(key, "AddonManifest/") || strings.HasPrefix(key, "ContainerRegistry/") {
			deps = append(deps, task)
		}
	}
	return deps
}

// Run implements fi.Task.Run. The push is idempotent (assets already present in
// the registry are skipped), so it does not use the Find/changes machinery.
func (r *RegistryAssets) Run(c *fi.CloudupContext) error {
	if r.Lifecycle == fi.LifecycleIgnore {
		return nil
	}

	switch target := c.Target.(type) {
	case *azure.AzureAPITarget:
		if r.push == nil {
			return fmt.Errorf("pushing assets to the registry is not supported in this context")
		}
		// A newly created registry can reject data-plane requests for several
		// minutes while permissions propagate; the push is idempotent, so ride
		// out the propagation window.
		deadline := time.Now().Add(15 * time.Minute)
		for {
			err := r.push()
			if err == nil {
				return nil
			}
			if time.Now().After(deadline) {
				return err
			}
			klog.Warningf("failed to push assets to the registry (will retry): %v", err)
			time.Sleep(30 * time.Second)
		}
	case *fi.CloudupDryRunTarget:
		klog.V(2).Infof("would push assets to registry %q", fi.ValueOf(r.Registry.Name))
		return nil
	default:
		klog.Warningf("assets are not pushed to the registry with target %T; run 'kops get assets --copy' after applying", target)
		return nil
	}
}
