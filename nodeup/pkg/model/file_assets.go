/*
Copyright 2019 The Kubernetes Authors.

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

package model

import (
	"encoding/base64"
	"fmt"
	"path/filepath"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// FileAssetsBuilder configures the hooks
type FileAssetsBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &FileAssetsBuilder{}

// Build is responsible for writing out the file assets from cluster and instanceGroup
func (f *FileAssetsBuilder) Build(c *fi.ModelBuilderContext) error {
	// used to keep track of previous file, so a instanceGroup can override a cluster wide one
	tracker := make(map[string]bool)

	// ensure the default path exists
	c.EnsureTask(&nodetasks.File{
		Path: f.FileAssetsDefaultPath(),
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	})

	// do we have any instanceGroup file assets
	if f.InstanceGroup.Spec.FileAssets != nil {
		if err := f.buildFileAssets(c, f.InstanceGroup.Spec.FileAssets, tracker); err != nil {
			return err
		}
	}
	if f.Cluster.Spec.FileAssets != nil {
		if err := f.buildFileAssets(c, f.Cluster.Spec.FileAssets, tracker); err != nil {
			return err
		}
	}

	return nil
}

// buildFileAssets is responsible for rendering the file assets to disk
func (f *FileAssetsBuilder) buildFileAssets(c *fi.ModelBuilderContext, assets []kops.FileAssetSpec, tracker map[string]bool) error {
	for _, asset := range assets {
		// @check if the file asset applies to us. If no roles applied we assume its applied to all roles
		if len(asset.Roles) > 0 && !containsRole(f.InstanceGroup.Spec.Role, asset.Roles) {
			continue
		}
		// @check if e have a path and if not use the default path
		assetPath := asset.Path
		if assetPath == "" {
			assetPath = filepath.Join(f.FileAssetsDefaultPath(), asset.Name)
		}
		// @check if the file has already been done and skip
		if _, found := tracker[assetPath]; found {
			continue
		}
		tracker[assetPath] = true // update the tracker

		// @check is the contents requires decoding
		content := asset.Content
		if asset.IsBase64 {
			decoded, err := base64.RawStdEncoding.DecodeString(content)
			if err != nil {
				return fmt.Errorf("failed on file asset: %s is invalid, unable to decode base64, error: %q", asset.Name, err)
			}
			content = string(decoded)
		}

		// We use EnsureTask so that we don't have to check if the asset directories have already been done
		c.EnsureTask(&nodetasks.File{
			Path: filepath.Dir(assetPath),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		c.AddTask(&nodetasks.File{
			Contents: fi.NewStringResource(content),
			Mode:     s("0440"),
			Path:     assetPath,
			Type:     nodetasks.FileType_File,
		})
	}

	return nil
}
