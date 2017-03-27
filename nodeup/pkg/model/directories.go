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
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// DirectoryBuilder creates required directories
type DirectoryBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DirectoryBuilder{}

func (b *DirectoryBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Distribution == distros.DistributionContainerOS {
		dir := "/home/kubernetes/bin"

		t := &nodetasks.File{
			Path: dir,
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),

			OnChangeExecute: [][]string{
				{"/bin/mount", "--bind", "/home/kubernetes/bin", "/home/kubernetes/bin"},
				{"/bin/mount", "-o", "remount,exec", "/home/kubernetes/bin"},
			},
		}
		c.AddTask(t)
	}

	return nil
}
