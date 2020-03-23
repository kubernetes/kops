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

	"k8s.io/kops/pkg/tokens"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// SecretBuilder writes secrets
type SecretBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &SecretBuilder{}

const (
	adminUser  = "admin"
	adminGroup = "system:masters"
)

// Build is responsible for pulling down the secrets
func (b *SecretBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.KeyStore == nil {
		return fmt.Errorf("KeyStore not set")
	}

	// @step: retrieve the platform ca
	if err := b.BuildCertificateTask(c, fi.CertificateId_CA, "ca.crt"); err != nil {
		return err
	}

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

	// if we are not a master we can stop here
	if !b.IsMaster {
		return nil
	}

	{
		name := "master"
		if err := b.BuildCertificateTask(c, name, "server.cert"); err != nil {
			return err
		}
		if err := b.BuildPrivateKeyTask(c, name, "server.key"); err != nil {
			return err
		}
	}

	{
		// TODO: Remove - we use the apiserver-aggregator keypair instead (which is signed by a different CA)
		if err := b.BuildCertificateTask(c, "apiserver-proxy-client", "proxy-client.cert"); err != nil {
			return err
		}
		if err := b.BuildPrivateKeyTask(c, "apiserver-proxy-client", "proxy-client.key"); err != nil {
			return err
		}
	}

	{
		if err := b.BuildCertificateTask(c, "apiserver-aggregator", "apiserver-aggregator.cert"); err != nil {
			return err
		}
		if err := b.BuildPrivateKeyTask(c, "apiserver-aggregator", "apiserver-aggregator.key"); err != nil {
			return err
		}
	}

	{
		if err := b.BuildCertificateTask(c, "apiserver-aggregator-ca", "apiserver-aggregator-ca.cert"); err != nil {
			return err
		}
	}

	// Support for basic auth was deprecated 1.16 and removed in 1.19
	// https://github.com/kubernetes/kubernetes/pull/89069
	if b.IsKubernetesLT("1.19") && b.SecretStore != nil {
		key := "kube"
		token, err := b.SecretStore.FindSecret(key)
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("token not found: %q", key)
		}
		csv := string(token.Data) + "," + adminUser + "," + adminUser + "," + adminGroup

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "basic_auth.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	if b.SecretStore != nil {
		allTokens, err := b.allAuthTokens()
		if err != nil {
			return err
		}

		var lines []string
		for id, token := range allTokens {
			if id == adminUser {
				lines = append(lines, token+","+id+","+id+","+adminGroup)
			} else {
				lines = append(lines, token+","+id+","+id)
			}
		}
		csv := strings.Join(lines, "\n")

		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "known_tokens.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		})
	}

	return nil
}

// allTokens returns a map of all auth tokens that are present
func (b *SecretBuilder) allAuthTokens() (map[string]string, error) {
	possibleTokens := tokens.GetKubernetesAuthTokens_Deprecated()

	tokens := make(map[string]string)
	for _, id := range possibleTokens {
		token, err := b.SecretStore.FindSecret(id)
		if err != nil {
			return nil, err
		}
		if token != nil {
			tokens[id] = string(token.Data)
		}
	}
	return tokens, nil
}
