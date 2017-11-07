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

package fi

import (
	"bytes"
	"crypto/x509"
	"fmt"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

const CertificateId_CA = "ca"

const (
	SecretTypeSSHPublicKey = "SSHPublicKey"
	SecretTypeKeypair      = "Keypair"
	SecretTypeSecret       = "Secret"

	// Name for the primary SSH key
	SecretNameSSHPrimary = "admin"
)

type KeystoreItem struct {
	Type string
	Name string
	Id   string
	Data []byte
}

// Keystore contains just the functions we need to issue keypairs, not to list / manage them
type Keystore interface {
	// FindKeypair finds a cert & private key, returning nil where either is not found
	// (if the certificate is found but not keypair, that is not an error: only the cert will be returned)
	FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, error)

	CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error)

	// StoreKeypair writes the keypair to the store
	StoreKeypair(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error

	// MirrorTo will copy secrets to a vfs.Path, which is often easier for a machine to read
	MirrorTo(basedir vfs.Path) error
}

// HasVFSPath is implemented by keystore & other stores that use a VFS path as their backing store
type HasVFSPath interface {
	VFSPath() vfs.Path
}

type CAStore interface {
	Keystore

	// Cert returns the primary specified certificate
	// For createIfMissing=false, using FindCert is preferred
	Cert(name string, createIfMissing bool) (*pki.Certificate, error)
	// CertificatePool returns all active certificates with the specified id
	CertificatePool(name string, createIfMissing bool) (*CertificatePool, error)
	PrivateKey(name string, createIfMissing bool) (*pki.PrivateKey, error)

	// FindCert returns the specified certificate, if it exists, or nil if not found
	FindCert(name string) (*pki.Certificate, error)
	FindPrivateKey(name string) (*pki.PrivateKey, error)

	// List will list all the items, but will not fetch the data
	List() ([]*KeystoreItem, error)

	// AddCert adds an alternative certificate to the pool (primarily useful for CAs)
	AddCert(name string, cert *pki.Certificate) error

	// AddSSHPublicKey adds an SSH public key
	AddSSHPublicKey(name string, data []byte) error

	// FindSSHPublicKeys retrieves the SSH public keys with the specific name
	FindSSHPublicKeys(name string) ([]*KeystoreItem, error)

	// DeleteSecret will delete the specified item
	DeleteSecret(item *KeystoreItem) error
}

type CertificatePool struct {
	Secondary []*pki.Certificate
	Primary   *pki.Certificate
}

func (c *CertificatePool) AsString() (string, error) {
	// Nicer behaviour because this is called from templates
	if c == nil {
		return "", fmt.Errorf("AsString called on nil CertificatePool")
	}

	var data bytes.Buffer
	if c.Primary != nil {
		_, err := c.Primary.WriteTo(&data)
		if err != nil {
			return "", fmt.Errorf("error writing SSL certificate: %v", err)
		}
	}
	for _, cert := range c.Secondary {
		_, err := cert.WriteTo(&data)
		if err != nil {
			return "", fmt.Errorf("error writing SSL certificate: %v", err)
		}
	}
	return data.String(), nil
}
