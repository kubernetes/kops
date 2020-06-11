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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// EtcdManagerTLSBuilder configures TLS support for etcd-manager
type EtcdManagerTLSBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcdManagerTLSBuilder{}

// Build is responsible for TLS configuration for etcd-manager
func (b *EtcdManagerTLSBuilder) Build(ctx *fi.ModelBuilderContext) error {
	if !b.IsMaster || !b.UseEtcdManager() {
		return nil
	}

	for _, k := range []string{"main", "events"} {
		d := "/etc/kubernetes/pki/etcd-manager-" + k

		keys := make(map[string]string)
		keys["etcd-manager-ca"] = "etcd-manager-ca-" + k
		keys["etcd-peers-ca"] = "etcd-peers-ca-" + k
		// Because API server can only have a single client-cert, we need to share a client CA
		keys["etcd-clients-ca"] = "etcd-clients-ca"

		for fileName, keystoreName := range keys {
			cert, err := b.KeyStore.FindCert(keystoreName)
			if err != nil {
				return err
			}
			if cert == nil {
				return fmt.Errorf("keypair %q not found", keystoreName)
			}

			if err := b.BuildCertificateTask(ctx, keystoreName, d+"/"+fileName+".crt"); err != nil {
				return err
			}
			if err := b.BuildPrivateKeyTask(ctx, keystoreName, d+"/"+fileName+".key"); err != nil {
				return err
			}
		}
	}

	// We also dynamically generate the client keypair for apiserver
	if err := b.buildKubeAPIServerKeypair(ctx); err != nil {
		return err
	}
	return nil
}

func (b *EtcdManagerTLSBuilder) buildKubeAPIServerKeypair(c *fi.ModelBuilderContext) error {
	name := "etcd-client"
	issueCert := &nodetasks.IssueCert{
		Name:   name,
		Signer: "etcd-clients-ca",
		Type:   "client",
		Subject: nodetasks.PKIXName{
			CommonName: "kube-apiserver",
		},
	}
	c.AddTask(issueCert)
	issueCert.AddFileTasks(c, "/etc/kubernetes/pki/kube-apiserver", name, "etcd-ca", nil)

	return nil
}
