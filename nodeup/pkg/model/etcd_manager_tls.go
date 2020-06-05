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
	"crypto/x509/pkix"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

// EtcdManagerTLSBuilder configures TLS support for etcd-manager
type EtcdManagerTLSBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcdManagerTLSBuilder{}

// Build is responsible for TLS configuration for etcd-manager
func (b *EtcdManagerTLSBuilder) Build(ctx *fi.ModelBuilderContext) error {
	if !b.IsMaster {
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
				klog.Warningf("keypair %q not found, won't configure", keystoreName)
				continue
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
	if err := b.buildKubeAPIServerKeypair(); err != nil {
		return err
	}
	return nil
}

func (b *EtcdManagerTLSBuilder) buildKubeAPIServerKeypair() error {
	req := &pki.IssueCertRequest{
		Signer: "etcd-clients-ca",
		Type:   "client",
		Subject: pkix.Name{
			CommonName: "kube-apiserver",
		},
		MinValidDays: 455,
	}
	dir := "/etc/kubernetes/pki/kube-apiserver"
	name := "etcd-client"
	humanName := dir + "/" + name
	klog.Infof("signing certificate for %q", humanName)
	cert, privateKey, etcdClientsCACertificate, err := pki.IssueCert(req, b.KeyStore)
	if err != nil {
		return fmt.Errorf("error signing certificate for %q: %v", humanName, err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories %q: %v", dir, err)
	}

	{
		p := filepath.Join(dir, "etcd-ca.crt")
		if err := etcdClientsCACertificate.WriteToFile(p, 0644); err != nil {
			return fmt.Errorf("error writing certificate key file %q: %v", p, err)
		}
	}

	p := filepath.Join(dir, name)
	if err := cert.WriteToFile(p+".crt", 0644); err != nil {
		return fmt.Errorf("error writing certificate key file %q: %v", p+".crt", err)
	}
	if err := privateKey.WriteToFile(p+".key", 0600); err != nil {
		return fmt.Errorf("error writing private key file %q: %v", p+".key", err)
	}

	return nil
}
