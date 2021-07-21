/*
Copyright 2021 The Kubernetes Authors.

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
	"bytes"
	"fmt"
	"io"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components/addonmanifests"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

var _ fi.Resource = &ManifestResource{}

type ManifestResource struct {
	template         fi.Resource
	addon            *api.AddonSpec
	kopsModelContext *model.KopsModelContext
	assetBuilder     *assets.AssetBuilder
	content          *string
}

func (m *ManifestResource) Open() (io.Reader, error) {

	if m.content == nil {

		manifestPath := ""
		manifestBytes, err := fi.ResourceAsBytes(m.template)
		if err != nil {
			return nil, fmt.Errorf("error reading manifest %s: %v", manifestPath, err)
		}

		// Go through any transforms that are best expressed as code
		remapped, err := addonmanifests.RemapAddonManifest(m.addon, m.kopsModelContext, m.assetBuilder, manifestBytes)
		if err != nil {
			klog.Infof("invalid manifest: %s", string(manifestBytes))
			return nil, fmt.Errorf("error remapping manifest %s: %v", manifestPath, err)
		}
		manifestBytes = remapped

		// Trim whitespace
		rawManifest := strings.TrimSpace(string(manifestBytes))

		klog.V(4).Infof("Manifest %v", rawManifest)

		m.content = &rawManifest
	}

	return strings.NewReader(*m.content), nil
}

type ChannelResource struct {
	addons        *api.Addons
	manifestTasks map[string]*fitasks.ManagedFile
}

var _ fi.HasDependencies = &ChannelResource{}

func (c *ChannelResource) Open() (io.Reader, error) {
	for _, addon := range c.addons.Spec.Addons {
		task := c.manifestTasks[*addon.Name]
		hash, err := task.GetHash()
		if err != nil {
			return nil, err
		}
		addon.ManifestHash = hash
	}

	addonsYAML, err := utils.YamlMarshal(c.addons)
	if err != nil {
		return nil, fmt.Errorf("error serializing addons yaml: %v", err)
	}

	return bytes.NewReader(addonsYAML), nil
}

// GetDependencies adds CA to the list of dependencies
func (c *ChannelResource) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var dependencies []fi.Task
	for _, task := range c.manifestTasks {
		dependencies = append(dependencies, task)
	}
	return dependencies
}
