/*
Copyright The Kubernetes Authors.

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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// GVisorBuilder installs the gVisor (runsc) sandboxed runtime.
// Only supported on Debian-family distributions.
type GVisorBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &GVisorBuilder{}

// Build installs gVisor packages via the upstream apt repository.
func (b *GVisorBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.InstallGVisorRuntime() {
		return nil
	}

	// gVisor packages are only published in the upstream apt repository,
	// so installation is limited to Debian-family distributions (Debian, Ubuntu).
	if !b.Distribution.IsDebianFamily() {
		return nil
	}

	c.AddTask(&nodetasks.AptSource{
		Name:    "gvisor",
		Keyring: "https://gvisor.dev/archive.key",
		Sources: []string{
			"deb [arch=$(ARCH)] https://storage.googleapis.com/gvisor/releases release main",
		},
	})
	// The runsc package bundles both runsc and containerd-shim-runsc-v1.
	c.AddTask(&nodetasks.Package{Name: "runsc"})

	return nil
}
