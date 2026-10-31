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
	"os"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kubemanifest"
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

	// instanceGroupKinds are the kinds of the objects that belong to the instance group of the same
	// name, see Addon.InstanceGroupKinds.
	instanceGroupKinds []schema.GroupKind
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

	if len(a.instanceGroupKinds) != 0 {
		manifestBytes, err = a.keepInstanceGroupObjects(c, manifestBytes)
		if err != nil {
			return fmt.Errorf("error keeping the instance group objects of addon %q: %w", fi.ValueOf(a.Name), err)
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

// keepInstanceGroupObjects adds to the manifest the objects of the instance groups not being
// updated, from the manifest in the state store, so that channels does not prune them. The objects
// of instance groups go after the others, in the order of AllInstanceGroups, as the template does.
func (a *AddonManifest) keepInstanceGroupObjects(c *fi.CloudupContext, manifest []byte) ([]byte, error) {
	updating := sets.New[string]()
	for _, ig := range a.modelContext.InstanceGroups {
		updating.Insert(ig.ObjectMeta.Name)
	}
	// An instance group that is no longer managed by Karpenter, such as one deleted and created
	// again with the same name as a cloud group, has no objects to keep.
	notUpdating := sets.New[string]()
	for _, ig := range a.modelContext.AllInstanceGroups {
		if !updating.Has(ig.ObjectMeta.Name) && ig.IsKarpenterManaged() {
			notUpdating.Insert(ig.ObjectMeta.Name)
		}
	}
	if notUpdating.Len() == 0 {
		return manifest, nil
	}

	previousManifest, err := c.T.ClusterConfigBase.Join(fi.ValueOf(a.Location)).ReadFile(c.Context())
	if err != nil {
		if os.IsNotExist(err) {
			// There is no previous manifest, as on a new cluster.
			return manifest, nil
		}
		return nil, fmt.Errorf("reading the previous manifest: %w", err)
	}
	previousObjects, err := kubemanifest.LoadObjectsFrom(previousManifest)
	if err != nil {
		return nil, fmt.Errorf("parsing the previous manifest: %w", err)
	}

	objectsByInstanceGroup := make(map[string]kubemanifest.ObjectList)
	for _, object := range previousObjects {
		if a.isInstanceGroupObject(object) && notUpdating.Has(object.GetName()) {
			objectsByInstanceGroup[object.GetName()] = append(objectsByInstanceGroup[object.GetName()], object)
		}
	}
	if len(objectsByInstanceGroup) == 0 {
		return manifest, nil
	}

	objects, err := kubemanifest.LoadObjectsFrom(manifest)
	if err != nil {
		return nil, fmt.Errorf("parsing the rendered manifest: %w", err)
	}
	var merged kubemanifest.ObjectList
	for _, object := range objects {
		if a.isInstanceGroupObject(object) && updating.Has(object.GetName()) {
			objectsByInstanceGroup[object.GetName()] = append(objectsByInstanceGroup[object.GetName()], object)
		} else {
			merged = append(merged, object)
		}
	}
	for _, ig := range a.modelContext.AllInstanceGroups {
		merged = append(merged, objectsByInstanceGroup[ig.ObjectMeta.Name]...)
	}
	return merged.ToYAML()
}

// isInstanceGroupObject reports whether the object belongs to the instance group named after it.
func (a *AddonManifest) isInstanceGroupObject(object *kubemanifest.Object) bool {
	return slices.Contains(a.instanceGroupKinds, object.GroupVersionKind().GroupKind())
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
