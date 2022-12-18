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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
)

// SecretBuilder writes secrets
type SecretBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &SecretBuilder{}

const (
	adminUser  = "admin"
	adminGroup = "system:masters"
)

// Build is responsible for pulling down the secrets
func (b *SecretBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	// @step: write out the platform ca
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(b.PathSrvKubernetes(), "ca.crt"),
		Contents: fi.NewStringResource(b.NodeupConfig.CAs[fi.CertificateIDCA]),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	})

	// Write out docker auth secret, if exists
	if b.SecretStore != nil {
		key := "dockerconfig"
		dockercfg, _ := b.SecretStore.Secret(key)
		if dockercfg != nil {
			contents := string(dockercfg.Data)
			c.AddTask(&nodetasks.File{
				Path:     filepath.Join("root", ".docker", "config.json"),
				Contents: fi.NewStringResource(contents),
				Type:     nodetasks.FileType_File,
				Mode:     s("0600"),
			})
		}
	}

	return nil
}

func getInstanceAddress() (string, error) {
	addrBytes, err := vfs.Context.ReadFile("metadata://openstack/local-ipv4")
	if err != nil {
		return "", nil
	}
	return string(addrBytes), nil
}
