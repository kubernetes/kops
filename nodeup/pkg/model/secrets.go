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

	"github.com/golang/glog"
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

	// retrieve the platform ca
	{
		ca, err := b.KeyStore.FindCertificatePool(fi.CertificateId_CA)
		if err != nil {
			return err
		}
		if ca == nil {
			return fmt.Errorf("certificate %q not found", fi.CertificateId_CA)
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
		cert, err := b.KeyStore.FindCert("master")
		if err != nil {
			return err
		}
		if cert == nil {
			return fmt.Errorf("certificate %q not found", "master")
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
		k, err := b.KeyStore.FindPrivateKey("master")
		if err != nil {
			return err
		}
		if k == nil {
			return fmt.Errorf("private key %q not found", "master")
		}
		serialized, err := k.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "server.key"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	if b.IsKubernetesGTE("1.7") {
		// TODO: Remove - we use the apiserver-aggregator keypair instead (which is signed by a different CA)
		cert, err := b.KeyStore.FindCert("apiserver-proxy-client")
		if err != nil {
			return fmt.Errorf("apiserver proxy client cert lookup failed: %v", err.Error())
		}
		if cert == nil {
			return fmt.Errorf("certificate %q not found", "apiserver-proxy-client")
		}

		serialized, err := cert.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "proxy-client.cert"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)

		key, err := b.KeyStore.FindPrivateKey("apiserver-proxy-client")
		if err != nil {
			return fmt.Errorf("apiserver proxy client private key lookup failed: %v", err.Error())
		}
		if key == nil {
			return fmt.Errorf("private key %q not found", "apiserver-proxy-client")
		}

		serialized, err = key.AsString()
		if err != nil {
			return err
		}

		t = &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "proxy-client.key"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	if b.IsKubernetesGTE("1.7") {
		if err := b.writeCertificate(c, "apiserver-aggregator"); err != nil {
			return err
		}

		if err := b.writePrivateKey(c, "apiserver-aggregator"); err != nil {
			return err
		}
	}

	if b.IsKubernetesGTE("1.7") {
		if err := b.writeCertificate(c, "apiserver-aggregator-ca"); err != nil {
			return err
		}
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

// writeCertificate writes the specified certificate to the local filesystem, under PathSrvKubernetes()
func (b *SecretBuilder) writeCertificate(c *fi.ModelBuilderContext, id string) error {
	cert, err := b.KeyStore.FindCert(id)
	if err != nil {
		return fmt.Errorf("cert lookup failed for %q: %v", id, err)
	}

	if cert != nil {
		serialized, err := cert.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), id+".cert"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	} else {
		// TODO: Make this an error?
		glog.Warningf("certificate %q not found", id)
	}

	return nil
}

// writePrivateKey writes the specified private key to the local filesystem, under PathSrvKubernetes()
func (b *SecretBuilder) writePrivateKey(c *fi.ModelBuilderContext, id string) error {
	key, err := b.KeyStore.FindPrivateKey(id)
	if err != nil {
		return fmt.Errorf("private key lookup failed for %q: %v", id, err)
	}
	if key == nil {
		return fmt.Errorf("private key %q not found", id)
	}

	serialized, err := key.AsString()
	if err != nil {
		return err
	}

	t := &nodetasks.File{
		Path:     filepath.Join(b.PathSrvKubernetes(), id+".key"),
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	}
	c.AddTask(t)

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
