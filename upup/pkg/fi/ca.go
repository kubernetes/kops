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
	"fmt"
	"math/big"
	"strconv"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

const CertificateIDCA = "ca"

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
	ID   string
	Data []byte
}

// Keyset is a parsed api.Keyset.
type Keyset struct {
	// LegacyFormat instructs a keypair task to convert a Legacy Keyset to the new Keyset API format.
	LegacyFormat bool
	Items        map[string]*KeysetItem
	Primary      *KeysetItem
}

// KeysetItem is a certificate/key pair in a Keyset.
type KeysetItem struct {
	Id          string
	Certificate *pki.Certificate
	PrivateKey  *pki.PrivateKey
}

// Keystore contains just the functions we need to issue keypairs, not to list / manage them
type Keystore interface {
	pki.Keystore

	// FindKeyset finds a Keyset.
	FindKeyset(name string) (*Keyset, error)

	// StoreKeyset writes a Keyset to the store.
	StoreKeyset(name string, keyset *Keyset) error

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

// FindPrimaryKeypair is a common implementation of pki.FindPrimaryKeypair.
func FindPrimaryKeypair(c Keystore, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	keyset, err := c.FindKeyset(name)
	if err != nil {
		return nil, nil, err
	}
	if keyset == nil || keyset.Primary == nil {
		return nil, nil, nil
	}
	return keyset.Primary.Certificate, keyset.Primary.PrivateKey, nil
}

// AddCert adds an alternative certificate to the keyset (primarily useful for CAs)
func AddCert(keyset *Keyset, cert *pki.Certificate) {
	serial := 0

	for keyset.Items[strconv.Itoa(serial)] != nil {
		serial++
	}

	keyset.Items[strconv.Itoa(serial)] = &KeysetItem{
		Id:          strconv.Itoa(serial),
		Certificate: cert,
	}
}

// KeysetItemIdOlder returns whether the KeysetItem Id a is older than b.
func KeysetItemIdOlder(a, b string) bool {
	aVersion, aOk := big.NewInt(0).SetString(a, 10)
	bVersion, bOk := big.NewInt(0).SetString(b, 10)
	if aOk {
		if !bOk {
			return false
		}
		return aVersion.Cmp(bVersion) < 0
	} else {
		if bOk {
			return true
		}
		return a < b
	}
}
