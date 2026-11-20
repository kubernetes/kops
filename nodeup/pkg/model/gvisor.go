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
	"fmt"
	"path/filepath"
	"regexp"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// GVisorBuilder installs the gVisor (runsc) sandboxed runtime from the gVisor
// release tarball in the node assets, so it works on every distribution.
type GVisorBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &GVisorBuilder{}

// gVisorBinaries are the top-level binaries of the gVisor release tarball.
var gVisorBinaries = []string{"runsc", "containerd-shim-runsc-v1"}

// Build installs runsc, containerd-shim-runsc-v1 and the gvisor-bin/ directory
// of sidecar binaries, which runsc runs from next to its own binary.
func (b *GVisorBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.InstallGVisorRuntime() {
		return nil
	}

	binDir := b.GVisorBinDir()

	for _, name := range gVisorBinaries {
		assets := b.Assets.FindMatches(regexp.MustCompile("^" + regexp.QuoteMeta(name) + "$"))
		if len(assets) != 1 {
			return fmt.Errorf("found %d %q binaries in the gVisor assets, want 1", len(assets), name)
		}
		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(binDir, name),
			Contents: assets[name],
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		})
	}

	sidecars := b.Assets.FindMatches(regexp.MustCompile(`^gvisor-bin/[^/]+$`))
	if len(sidecars) == 0 {
		return fmt.Errorf("found no gvisor-bin/ sidecar binaries in the gVisor assets")
	}
	// The sentry may re-execute itself as an unprivileged user, so the
	// directory and the binaries in it must be readable and executable by all.
	c.AddTask(&nodetasks.File{
		Path: filepath.Join(binDir, "gvisor-bin"),
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	})
	for name, contents := range sidecars {
		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(binDir, "gvisor-bin", name),
			Contents: contents,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		})
	}

	return nil
}
