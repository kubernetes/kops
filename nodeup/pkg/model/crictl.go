/*
Copyright 2024 The Kubernetes Authors.

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
	"fmt"
	"path/filepath"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

type CrictlBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &CrictlBuilder{}

func (b *CrictlBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	assetName := "crictl"
	assetPath := ""
	asset, err := b.Assets.Find(assetName, assetPath)
	if err != nil {
		return fmt.Errorf("unable to locate asset %q: %w", assetName, err)
	}

	c.AddTask(&nodetasks.File{
		Path:     b.crictlPath(),
		Contents: asset,
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	})

	return nil
}

func (b *CrictlBuilder) binaryPath() string {
	path := "/usr/local/bin"
	if b.Distribution == distributions.DistributionFlatcar {
		path = "/opt/kops/bin"
	}
	if b.Distribution == distributions.DistributionContainerOS {
		path = "/home/kubernetes/bin"
	}
	return path
}

func (b *CrictlBuilder) crictlPath() string {
	return filepath.Join(b.binaryPath(), "crictl")
}
