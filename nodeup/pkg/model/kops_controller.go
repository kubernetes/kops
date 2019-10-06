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
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"path/filepath"

	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/pkiutil"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KopsControllerBuilder installs the keys for a kops-controller
type KopsControllerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KopsControllerBuilder{}

// Build is responsible for configuring keys that will be used by kops-controller (via hostPath)
func (b *KopsControllerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	// Create the directory, even if we aren't going to populate it
	pkiDir := "/etc/kubernetes/pki/kops-controller"
	c.AddTask(&nodetasks.File{
		Path: pkiDir,
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	})

	if !b.UseKopsControllerForKubeletBootstrap() {
		return nil
	}

	// We run kops-controller under an unprivileged user (wellknownusers.KopsControllerID), and then grant specific permissions
	c.AddTask(&nodetasks.UserTask{
		Name:  wellknownusers.KopsControllerName,
		UID:   wellknownusers.KopsControllerID,
		Shell: "/sbin/nologin",
	})

	serverKey, serverCert, err := b.buildServerKeypair()
	if err != nil {
		return err
	}

	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(pkiDir, "server.crt"),
		Contents: fi.NewBytesResource(pkiutil.EncodeCertPEM(serverCert)),
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
		Owner:    s(wellknownusers.KopsControllerName),
	})
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(pkiDir, "server.key"),
		Contents: fi.NewBytesResource(pkiutil.EncodePrivateKeyPEM(serverKey)),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
		Owner:    s(wellknownusers.KopsControllerName),
	})

	return nil
}

func (b *KopsControllerBuilder) buildServerKeypair() (*rsa.PrivateKey, *x509.Certificate, error) {
	commonName := "kops-controller"
	certConfig := &certutil.Config{
		CommonName: commonName,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certConfig.AltNames.DNSNames = []string{
		b.Cluster.Spec.MasterInternalName,
	}

	signerID := fi.CertificateId_CA

	var signerKey *pki.PrivateKey
	{
		k, err := b.KeyStore.FindPrivateKey(signerID)
		if err != nil {
			return nil, nil, err
		}

		if k == nil {
			return nil, nil, fmt.Errorf("private key %q not found", signerID)
		}
		signerKey = k
	}

	var signerCertificate *pki.Certificate
	{
		cert, err := b.KeyStore.FindCert(signerID)
		if err != nil {
			return nil, nil, err
		}

		if cert == nil {
			return nil, nil, fmt.Errorf("certificate %q not found", signerID)
		}

		signerCertificate = cert
	}

	humanName := certConfig.CommonName

	privateKey, err := pkiutil.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	signer, ok := signerKey.Key.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key was not a Signer: %T", signerKey.Key)
	}

	klog.Infof("signing certificate for %q", humanName)
	cert, err := pkiutil.NewSignedCert(certConfig, privateKey, signerCertificate.Certificate, signer)
	if err != nil {
		return nil, nil, fmt.Errorf("error signing certificate for %q: %v", humanName, err)
	}

	return privateKey, cert, nil
}
