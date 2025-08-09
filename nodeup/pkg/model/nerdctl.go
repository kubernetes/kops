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
	"path/filepath"
	"regexp"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

type NerdctlBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &NerdctlBuilder{}

func (b *NerdctlBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.skipInstall() {
		klog.V(8).Info("won't install nerdctl")
		return nil
	}

	assets := b.Assets.FindMatches(regexp.MustCompile(`^nerdctl$`))
	if len(assets) == 0 {
		klog.Warning("unable to find any nerdctl binaries in assets")
		return nil
	}
	if len(assets) > 1 {
		klog.Warning("multiple nerdctl binaries are found")
		return nil
	}

	for k, v := range assets {
		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(b.binaryPath(), k),
			Contents: v,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		})
	}

	return nil
}

func (b *NerdctlBuilder) binaryPath() string {
	path := "/usr/local/bin"
	if b.Distribution == distributions.DistributionFlatcar {
		path = "/opt/kops/bin"
	}
	if b.Distribution == distributions.DistributionContainerOS {
		path = "/home/kubernetes/bin"
	}
	return path
}

func (b *NerdctlBuilder) skipInstall() bool {
	containerd := b.NodeupConfig.ContainerdConfig

	if containerd == nil {
		return false
	}

	return containerd.SkipInstall && !containerd.InstallNerdCtl
}
