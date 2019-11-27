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
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
)

// ManifestsBuilder copies manifests from the store (e.g. etcdmanager)
type ManifestsBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ManifestsBuilder{}

// Build creates tasks for copying the manifests
func (b *ManifestsBuilder) Build(c *fi.ModelBuilderContext) error {
	// Write etcd manifests (currently etcd <=> master)
	if b.IsMaster {
		for _, manifest := range b.NodeupConfig.EtcdManifests {
			p, err := vfs.Context.BuildVfsPath(manifest)
			if err != nil {
				return fmt.Errorf("error parsing path for etcd manifest %s: %v", manifest, err)
			}
			data, err := p.ReadFile()
			if err != nil {
				return fmt.Errorf("error reading etcd manifest %s: %v", manifest, err)
			}

			name := p.Base()
			name = strings.TrimSuffix(name, filepath.Ext(name))

			key := "etcd-" + name

			manifestPath := "/etc/kubernetes/manifests/" + key + ".manifest"

			c.AddTask(&nodetasks.File{
				Contents: fi.NewBytesResource(data),
				Mode:     s("0440"),
				Path:     manifestPath,
				Type:     nodetasks.FileType_File,
			})
		}
	}

	return nil
}
