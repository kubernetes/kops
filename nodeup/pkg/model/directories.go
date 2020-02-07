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
	"path/filepath"

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// DirectoryBuilder creates required directories
type DirectoryBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DirectoryBuilder{}

// Build is responsible for specific directories are created - os dependent
func (b *DirectoryBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Distribution == distros.DistributionContainerOS {
		dirname := "/home/kubernetes/bin"

		c.AddTask(&nodetasks.File{
			Path: dirname,
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		c.AddTask(&nodetasks.BindMount{
			Source:     dirname,
			Mountpoint: dirname,
			Options:    []string{"exec"},
		})
	}

	// We try to put things into /opt/kops
	// On some OSes though, /opt/ is not writeable, and we can't even create the mountpoint
	if b.Distribution == distros.DistributionContainerOS {
		src := "/mnt/stateful_partition/opt/"

		c.AddTask(&nodetasks.File{
			Path: src,
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		// Rebuild things we are masking
		c.AddTask(&nodetasks.File{
			Path: filepath.Join(src, "google"),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})
		c.AddTask(&nodetasks.File{
			Path: filepath.Join(src, "google", "crash-reporter"),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})
		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(src, "google", "crash-reporter", "filter"),
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
			Contents: fi.NewStringResource(cosCrashFilter),
		})

		// Precreate the directory that will be /opt/kops, so we can bind remount it
		c.AddTask(&nodetasks.File{
			Path: filepath.Join(src, "kops"),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})
		c.AddTask(&nodetasks.File{
			Path: filepath.Join(src, "kops", "bin"),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		c.AddTask(&nodetasks.BindMount{
			Source:     src,
			Mountpoint: "/opt",
			Options:    []string{"ro"},
		})

		c.AddTask(&nodetasks.BindMount{
			Source:     filepath.Join(src, "kops", "bin"),
			Mountpoint: "/opt/kops/bin",
			Options:    []string{"exec", "nosuid", "nodev"},
		})

		// /opt/cni and /opt/cni/bin
		c.AddTask(&nodetasks.File{
			Path: filepath.Join(src, "cni"),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})
		c.AddTask(&nodetasks.File{
			Path: filepath.Join(src, "cni", "bin"),
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		c.AddTask(&nodetasks.BindMount{
			Source:     filepath.Join(src, "cni", "bin"),
			Mountpoint: "/opt/cni/bin",
			Options:    []string{"exec", "nosuid", "nodev"},
		})
	}

	return nil
}

// cosCrashFilter is used on COS to prevent userspace crash-reporting
// This is the one thing we need from /opt
const cosCrashFilter = `#!/bin/bash
# Copyright 2016 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Do no collect any userland crash.
exit 1
`
