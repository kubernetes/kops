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
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
	"k8s.io/kops/pkg/pkiutil"
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
	etcdClientsCACertificate, err := b.KeyStore.FindCert("etcd-clients-ca")
	if err != nil {
		return err
	}

	etcdClientsCAPrivateKey, err := b.KeyStore.FindPrivateKey("etcd-clients-ca")
	if err != nil {
		return err
	}

	if etcdClientsCACertificate == nil {
		klog.Errorf("unable to find etcd-clients-ca certificate, won't build key for apiserver")
		return nil
	}
	if etcdClientsCAPrivateKey == nil {
		klog.Errorf("unable to find etcd-clients-ca private key, won't build key for apiserver")
		return nil
	}

	dir := "/etc/kubernetes/pki/kube-apiserver"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories %q: %v", dir, err)
	}

	{
		p := filepath.Join(dir, "etcd-ca.crt")
		certBytes := pkiutil.EncodeCertPEM(etcdClientsCACertificate.Certificate)
		if err := ioutil.WriteFile(p, certBytes, 0644); err != nil {
			return fmt.Errorf("error writing certificate key file %q: %v", p, err)
		}
	}

	name := "etcd-client"

	humanName := dir + "/" + name
	privateKey, err := pkiutil.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("unable to create private key %q: %v", humanName, err)
	}
	privateKeyBytes := pkiutil.EncodePrivateKeyPEM(privateKey)

	certConfig := &certutil.Config{
		CommonName: "kube-apiserver",
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	signingKey, ok := etcdClientsCAPrivateKey.Key.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("etcd-clients-ca private key had unexpected type %T", etcdClientsCAPrivateKey.Key)
	}

	klog.Infof("signing certificate for %q", humanName)
	cert, err := pkiutil.NewSignedCert(certConfig, privateKey, etcdClientsCACertificate.Certificate, signingKey)
	if err != nil {
		return fmt.Errorf("error signing certificate for %q: %v", humanName, err)
	}

	certBytes := pkiutil.EncodeCertPEM(cert)

	p := filepath.Join(dir, name)
	{
		if err := ioutil.WriteFile(p+".crt", certBytes, 0644); err != nil {
			return fmt.Errorf("error writing certificate key file %q: %v", p+".crt", err)
		}
	}

	{
		if err := ioutil.WriteFile(p+".key", privateKeyBytes, 0600); err != nil {
			return fmt.Errorf("error writing private key file %q: %v", p+".key", err)
		}
	}

	return nil
}
