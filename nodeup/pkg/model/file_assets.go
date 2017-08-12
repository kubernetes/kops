/*
Copyright 2016 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
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
	tracker := make(map[string]bool, 0)
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
func (f *FileAssetsBuilder) buildFileAssets(c *fi.ModelBuilderContext, assets []*kops.FileAssetSpec, tracker map[string]bool) error {
	for _, asset := range assets {
		if err := validateFileAsset(asset); err != nil {
			return fmt.Errorf("The file asset is invalid, name: %s, error: %q", asset.Name, err)
		}
		// @check if the file asset applys to us. If no roles applied we assume its applied to all roles
		// @todo: use the containsRole when the hooks PR is merged
		if len(asset.Roles) > 0 {
			var found bool
			for _, x := range asset.Roles {
				if f.InstanceGroup.Spec.Role == x {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// @check if the file has already been done and skip
		if _, found := tracker[asset.Path]; found {
			continue
		}
		tracker[asset.Path] = true // update the tracker

		// fill in the defaults for the file perms
		if asset.Mode == "" {
			asset.Mode = "0400"
		}
		// @check is the contents requires decoding
		content := asset.Content
		if asset.IsBase64 {
			decoded, err := base64.RawStdEncoding.DecodeString(content)
			if err != nil {
				return fmt.Errorf("Failed on file asset: %s is invalid, unable to decode base64, error: %q", asset.Name, err)
			}
			content = string(decoded)
		}

		// @check if the directory structure exist or create it
		c.AddTask(&nodetasks.File{
			Path: filepath.Dir(asset.Path),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		var resource fi.Resource
		var err error
		switch asset.Templated {
		case true:
			resource, err = f.getRenderedResource(content)
			if err != nil {
				return fmt.Errorf("Failed on file assets: %s, build rendered resource, error: %q", asset.Name, err)
			}
		default:
			resource = fi.NewStringResource(content)
		}

		// @check the file permissions
		perms := asset.Mode
		if !strings.HasPrefix(perms, "0") {
			perms = fmt.Sprintf("%d%s", 0, perms)
		}

		c.AddTask(&nodetasks.File{
			Contents: resource,
			Mode:     s(perms),
			Path:     asset.Path,
			Type:     nodetasks.FileType_File,
		})
	}

	return nil
}

// @perhaps a path finder?
var templateFuncs = template.FuncMap{
	"split": strings.Split,
	"join":  strings.Join,
}

// getRenderedResource is responsible for rendering the content if templated
func (f *FileAssetsBuilder) getRenderedResource(content string) (fi.Resource, error) {
	context := map[string]interface{}{
		"Cluster":       f.Cluster.Spec,
		"InstanceGroup": f.InstanceGroup.Spec,
		"Master":        fmt.Sprintf("%t", f.IsMaster),
		"Name":          f.InstanceGroup.Name,
	}

	resource, err := model.NewTemplateResource("FileAsset", content, templateFuncs, context)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// validateFileAsset performs some basic validation on the asset
func validateFileAsset(asset *kops.FileAssetSpec) error {
	if asset.Path == "" {
		return errors.New("does not have a path")
	}
	if asset.Content == "" {
		return errors.New("does not have any contents")
	}

	return nil
}
