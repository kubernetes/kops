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

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// +kops:fitask
type BootstrapChannel struct {
	Name      *string
	Lifecycle fi.Lifecycle
	Location  *string
	Contents  fi.Resource
	PublicACL *bool

	addonManifests []*AddonManifest
}

var (
	_ fi.CloudupTaskNormalize   = (*BootstrapChannel)(nil)
	_ fi.CloudupHasDependencies = (*BootstrapChannel)(nil)
)

func (a *BootstrapChannel) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	dependencies := make([]fi.CloudupTask, 0, len(a.addonManifests))
	for _, manifest := range a.addonManifests {
		dependencies = append(dependencies, manifest)
	}
	return dependencies
}

func (a *BootstrapChannel) Normalize(c *fi.CloudupContext) error {
	addonsObject := &channelsapi.Addons{}
	addonsObject.Kind = "Addons"
	addonsObject.ObjectMeta.Name = "bootstrap"

	for _, manifest := range a.addonManifests {
		if manifest.addonSpec == nil {
			return fmt.Errorf("addon manifest %q did not have a spec", fi.ValueOf(manifest.Name))
		}
		if manifest.addonSpec.ManifestHash == "" {
			return fmt.Errorf("addon %q manifest hash was not populated", fi.ValueOf(manifest.addonSpec.Name))
		}
		addonsObject.Spec.Addons = append(addonsObject.Spec.Addons, manifest.addonSpec)
	}

	if err := addonsObject.Verify(); err != nil {
		return err
	}

	addonsYAML, err := utils.YamlMarshal(addonsObject)
	if err != nil {
		return fmt.Errorf("error serializing addons yaml: %v", err)
	}

	a.Contents = fi.NewBytesResource(addonsYAML)
	return nil
}

func (a *BootstrapChannel) Find(c *fi.CloudupContext) (*BootstrapChannel, error) {
	managedFile := a.toManagedFile()
	actual, err := managedFile.Find(c)
	if err != nil || actual == nil {
		return nil, err
	}
	a.PublicACL = managedFile.PublicACL

	return &BootstrapChannel{
		Name:      a.Name,
		Lifecycle: a.Lifecycle,
		Location:  a.Location,
		Contents:  actual.Contents,
		PublicACL: actual.PublicACL,
	}, nil
}

func (a *BootstrapChannel) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(a, c)
}

func (a *BootstrapChannel) CheckChanges(actual, expected, changes *BootstrapChannel) error {
	return expected.toManagedFile().CheckChanges(actual.toManagedFile(), expected.toManagedFile(), changes.toManagedFile())
}

func (a *BootstrapChannel) Render(c *fi.CloudupContext, actual, expected, changes *BootstrapChannel) error {
	return expected.toManagedFile().Render(c, actual.toManagedFile(), expected.toManagedFile(), changes.toManagedFile())
}

func (a *BootstrapChannel) RenderTerraform(c *fi.CloudupContext, t *terraform.TerraformTarget, actual, expected, changes *BootstrapChannel) error {
	return expected.toManagedFile().RenderTerraform(c, t, actual.toManagedFile(), expected.toManagedFile(), changes.toManagedFile())
}

func (a *BootstrapChannel) toManagedFile() *fitasks.ManagedFile {
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
