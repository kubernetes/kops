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

package bootstrapchannelbuilder

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components/addonmanifests"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// +kops:fitask
type AddonManifest struct {
	Name      *string
	Lifecycle fi.Lifecycle
	Location  *string
	Contents  fi.Resource
	PublicACL *bool

	addonRenderer AddonTemplateRenderer
	source        fi.Resource
	// addonSpec is mutated during Normalize and subsequently read by BootstrapChannel,
	// which depends on this task so the mutations are fully visible by the time it runs.
	addonSpec       *channelsapi.AddonSpec
	buildPrune      bool
	skipRemap       bool
	skipRender      bool
	modelContext    *model.KopsModelContext
	assetBuilder    *assets.AssetBuilder
	serviceAccounts map[types.NamespacedName]iam.Subject
}

var (
	_ fi.CloudupTaskNormalize   = (*AddonManifest)(nil)
	_ fi.CloudupHasDependencies = (*AddonManifest)(nil)
)

// GetDependencies makes the addon wait for every non-addon task, so templates that reach into
// the task graph see fully realized state. Intentionally pessimistic: addon rendering is fast
// enough that over-depending is not worth the footgun of under-declaring.
func (a *AddonManifest) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	dependencies := make([]fi.CloudupTask, 0, len(tasks))
	for _, task := range tasks {
		if isAddonTask(task) {
			continue
		}
		dependencies = append(dependencies, task)
	}
	return dependencies
}

func (a *AddonManifest) Normalize(c *fi.CloudupContext) error {
	if a.addonSpec == nil {
		return fmt.Errorf("addon spec is not configured for %q", fi.ValueOf(a.Name))
	}
	if a.source == nil {
		return fmt.Errorf("addon source is not configured for %q", fi.ValueOf(a.Name))
	}

	manifestBytes, err := fi.ResourceAsBytes(a.source)
	if err != nil {
		return fmt.Errorf("error reading addon %q manifest: %v", fi.ValueOf(a.Name), err)
	}

	if !a.skipRender && a.addonRenderer != nil {
		manifestBytes, err = a.addonRenderer.RenderTemplate(fi.ValueOf(a.Location), manifestBytes, tasksVisibleToAddons(c.AllTasks()))
		if err != nil {
			return fmt.Errorf("error rendering addon %q template: %w", fi.ValueOf(a.Name), err)
		}
	}

	if !a.skipRemap {
		manifestBytes, err = addonmanifests.RemapAddonManifest(a.addonSpec, a.modelContext, a.assetBuilder, manifestBytes, a.serviceAccounts)
		if err != nil {
			klog.Infof("invalid manifest: %s", string(manifestBytes))
			return fmt.Errorf("error remapping manifest %s: %v", fi.ValueOf(a.Location), err)
		}
	}

	manifestBytes = []byte(strings.TrimSpace(string(manifestBytes)))

	if a.buildPrune {
		if err := buildPruneDirectives(a.addonSpec, manifestBytes); err != nil {
			return fmt.Errorf("failed to configure pruning for %s: %w", fi.ValueOf(a.addonSpec.Name), err)
		}
	}

	rawManifest := string(manifestBytes)
	manifestHash, err := utils.HashString(rawManifest)
	if err != nil {
		return fmt.Errorf("error hashing manifest: %v", err)
	}
	a.addonSpec.ManifestHash = manifestHash
	a.Contents = fi.NewBytesResource(manifestBytes)

	return nil
}

// Find returns a sparsely-populated AddonManifest reflecting the stored ManagedFile: only the
// fields needed by CheckChanges/Render/RenderTerraform (which delegate to toManagedFile) are set.
// Render-only fields such as addonSpec and source are intentionally left nil since they have no
// meaning for an already-materialized remote file.
func (a *AddonManifest) Find(c *fi.CloudupContext) (*AddonManifest, error) {
	managedFile := a.toManagedFile()
	actual, err := managedFile.Find(c)
	if err != nil || actual == nil {
		return nil, err
	}
	a.PublicACL = managedFile.PublicACL

	return &AddonManifest{
		Name:      a.Name,
		Lifecycle: a.Lifecycle,
		Location:  a.Location,
		Contents:  actual.Contents,
		PublicACL: actual.PublicACL,
	}, nil
}

func (a *AddonManifest) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(a, c)
}

func (a *AddonManifest) CheckChanges(actual, expected, changes *AddonManifest) error {
	return expected.toManagedFile().CheckChanges(actual.toManagedFile(), expected.toManagedFile(), changes.toManagedFile())
}

func (a *AddonManifest) Render(c *fi.CloudupContext, actual, expected, changes *AddonManifest) error {
	return expected.toManagedFile().Render(c, actual.toManagedFile(), expected.toManagedFile(), changes.toManagedFile())
}

func (a *AddonManifest) RenderTerraform(c *fi.CloudupContext, t *terraform.TerraformTarget, actual, expected, changes *AddonManifest) error {
	return expected.toManagedFile().RenderTerraform(c, t, actual.toManagedFile(), expected.toManagedFile(), changes.toManagedFile())
}

func (a *AddonManifest) toManagedFile() *fitasks.ManagedFile {
	if a == nil {
		return nil
	}
	return &fitasks.ManagedFile{
		Name:      a.Name,
		Lifecycle: a.Lifecycle,
		Location:  a.Location,
		Contents:  a.Contents,
		PublicACL: a.PublicACL,
	}
}

// tasksVisibleToAddons returns all tasks that are valid references from an addon template.
func tasksVisibleToAddons(tasks map[string]fi.CloudupTask) map[string]fi.CloudupTask {
	filtered := make(map[string]fi.CloudupTask, len(tasks))
	for key, task := range tasks {
		if isAddonTask(task) {
			continue
		}
		filtered[key] = task
	}
	return filtered
}

// isAddonTask reports whether a task belongs to the addon-rendering pipeline.
func isAddonTask(task fi.CloudupTask) bool {
	switch task.(type) {
	case *AddonManifest, *BootstrapChannel:
		return true
	}
	return false
}
