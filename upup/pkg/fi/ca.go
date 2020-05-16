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

package fi

import (
	"bytes"
	"crypto/x509"
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

const CertificateId_CA = "ca"

const (
	// SecretNameSSHPrimary is the Name for the primary SSH key
	SecretNameSSHPrimary = "admin"
)

const (
	keysetFormatLatest = "v1alpha2"
)

type KeystoreItem struct {
	Type kops.KeysetType
	Name string
	Id   string
	Data []byte
}

// Keystore contains just the functions we need to issue keypairs, not to list / manage them
type Keystore interface {
	// FindKeypair finds a cert & private key, returning nil where either is not found
	// (if the certificate is found but not keypair, that is not an error: only the cert will be returned).
	// This func returns a cert, private key and a bool.  The bool value is whether the keypair is stored
	// in a legacy format. This bool is used by a keypair
	// task to convert a Legacy Keypair to the new Keypair API format.
	FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, bool, error)

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

	// FindCertificatePool returns the named CertificatePool, or (nil,nil) if not found
	FindCertificatePool(name string) (*CertificatePool, error)

	// FindCertificateKeyset will return the keyset for a certificate
	FindCertificateKeyset(name string) (*kops.Keyset, error)

	// FindPrivateKey returns the named private key, or (nil,nil) if not found
	FindPrivateKey(name string) (*pki.PrivateKey, error)

	// FindPrivateKeyset will return the keyset for a private key
	FindPrivateKeyset(name string) (*kops.Keyset, error)

	// FindCert returns the specified certificate, if it exists, or nil if not found
	FindCert(name string) (*pki.Certificate, error)

	// ListKeysets will return all the KeySets
	// The key material is not guaranteed to be populated - metadata like the name will be.
	ListKeysets() ([]*kops.Keyset, error)

	// AddCert adds an alternative certificate to the pool (primarily useful for CAs)
	AddCert(name string, cert *pki.Certificate) error

	// DeleteKeysetItem will delete the specified item from the Keyset
	DeleteKeysetItem(item *kops.Keyset, id string) error
}

// SSHCredentialStore holds SSHCredential objects
type SSHCredentialStore interface {
	// DeleteSSHCredential deletes the specified SSH credential
	DeleteSSHCredential(item *kops.SSHCredential) error

	// ListSSHCredentials will list all the SSH credentials
	ListSSHCredentials() ([]*kops.SSHCredential, error)

	// AddSSHPublicKey adds an SSH public key
	AddSSHPublicKey(name string, data []byte) error

	// FindSSHPublicKeys retrieves the SSH public keys with the specific name
	FindSSHPublicKeys(name string) ([]*kops.SSHCredential, error)
}

type CertificatePool struct {
	Secondary []*pki.Certificate
	Primary   *pki.Certificate
}

func (c *CertificatePool) All() []*pki.Certificate {
	var certs []*pki.Certificate
	if c.Primary != nil {
		certs = append(certs, c.Primary)
	}
	if len(c.Secondary) != 0 {
		certs = append(certs, c.Secondary...)
	}
	return certs
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
