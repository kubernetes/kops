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
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// SecretBuilder writes secrets
type SecretBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &SecretBuilder{}

// Build is responisble for pulling down the secrets
func (b *SecretBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.KeyStore == nil {
		return fmt.Errorf("KeyStore not set")
	}

	// retrieve the platform ca
	{
		ca, err := b.KeyStore.CertificatePool(fi.CertificateId_CA)
		if err != nil {
			return err
		}

		serialized, err := ca.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "ca.crt"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	if b.SecretStore != nil {
		key := "dockerconfig"
		dockercfg, _ := b.SecretStore.Secret(key)
		if dockercfg != nil {
			contents := string(dockercfg.Data)
			t := &nodetasks.File{
				Path:     filepath.Join("root", ".docker", "config.json"),
				Contents: fi.NewStringResource(contents),
				Type:     nodetasks.FileType_File,
				Mode:     s("0600"),
			}
			c.AddTask(t)
		}
	}

	// if we are not a master we can stop here
	if !b.IsMaster {
		return nil
	}

	{
		cert, err := b.KeyStore.Cert("master")
		if err != nil {
			return err
		}

		serialized, err := cert.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "server.cert"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	{
		k, err := b.KeyStore.PrivateKey("master")
		if err != nil {
			return err
		}

		serialized, err := k.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "server.key"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	if b.SecretStore != nil {
		key := "kube"
		token, err := b.SecretStore.FindSecret(key)
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("token not found: %q", key)
		}
		csv := string(token.Data) + ",admin,admin,system:masters"

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "basic_auth.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	if b.SecretStore != nil {
		allTokens, err := b.allTokens()
		if err != nil {
			return err
		}

		var lines []string
		for id, token := range allTokens {
			if id == "dockerconfig" {
				continue
			}
			lines = append(lines, token+","+id+","+id)
		}
		csv := strings.Join(lines, "\n")

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "known_tokens.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	return nil
}

// allTokens returns a map of all tokens
func (b *SecretBuilder) allTokens() (map[string]string, error) {
	tokens := make(map[string]string)
	ids, err := b.SecretStore.ListSecrets()
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		token, err := b.SecretStore.FindSecret(id)
		if err != nil {
			return nil, err
		}
		tokens[id] = string(token.Data)
	}
	return tokens, nil
}
