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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"path/filepath"
	"strings"
)

// SecretBuilder writes secrets
type SecretBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &SecretBuilder{}

func (b *SecretBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.KeyStore == nil {
		return fmt.Errorf("KeyStore not set")
	}

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
			lines = append(lines, token+","+id+","+id)
		}
		csv := strings.Join(lines, "\n")

		// TODO: If we want to use tokens with RBAC, we need to add the roles
		// cluster/gce/gci/configure-helper.sh has this:
		//replace_prefixed_line "${known_tokens_csv}" "${KUBE_BEARER_TOKEN},"             "admin,admin,system:masters"
		//replace_prefixed_line "${known_tokens_csv}" "${KUBE_CONTROLLER_MANAGER_TOKEN}," "system:kube-controller-manager,uid:system:kube-controller-manager"
		//replace_prefixed_line "${known_tokens_csv}" "${KUBE_SCHEDULER_TOKEN},"          "system:kube-scheduler,uid:system:kube-scheduler"
		//replace_prefixed_line "${known_tokens_csv}" "${KUBELET_TOKEN},"                 "kubelet,uid:kubelet,system:nodes"
		//replace_prefixed_line "${known_tokens_csv}" "${KUBE_PROXY_TOKEN},"              "system:kube-proxy,uid:kube_proxy"
		//replace_prefixed_line "${known_tokens_csv}" "${NODE_PROBLEM_DETECTOR_TOKEN},"   "system:node-problem-detector,uid:node-problem-detector"

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
