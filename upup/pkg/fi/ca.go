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
	"encoding/pem"
	"fmt"
	"math/big"
	"sort"
	"time"

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

func (k *Keyset) ToPublicKeyBytes() ([]byte, error) {
	keys := make([]string, 0, len(k.Items))
	for k := range k.Items {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return KeysetItemIdOlder(k.Items[keys[i]].Id, k.Items[keys[j]].Id)
	})

	buf := new(bytes.Buffer)
	for _, key := range keys {
		item := k.Items[key]
		publicKeyData, err := x509.MarshalPKIXPublicKey(item.Certificate.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("marshalling public key %s: %v", item.Id, err)
		}
		if err = pem.Encode(buf, &pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyData}); err != nil {
			return nil, fmt.Errorf("encoding public key %s: %v", item.Id, err)
		}
	}
	return buf.Bytes(), nil
}

// AddItem adds an item to the keyset
func (k *Keyset) AddItem(cert *pki.Certificate, privateKey *pki.PrivateKey, primary bool) error {
	if cert == nil {
		return fmt.Errorf("no certificate provided")
	}
	if privateKey == nil && primary {
		return fmt.Errorf("private key not provided for primary item")
	}

	if !primary && k.Primary == nil {
		return fmt.Errorf("cannot add secondary item when no existing primary item")
	}

	highestId := big.NewInt(0)
	for id := range k.Items {
		itemId, ok := big.NewInt(0).SetString(id, 10)
		if ok && highestId.Cmp(itemId) < 0 {
			highestId = itemId
		}
	}

	// Make sure any subsequently created items will have ids that compare higher.
	// If setting a primary, make sure its id doesn't compare lower than existing items.
	idNumber := pki.BuildPKISerial(time.Now().UnixNano())
	if cert.Certificate.SerialNumber.Cmp(idNumber) <= 0 &&
		(!primary || cert.Certificate.SerialNumber.Cmp(highestId) > 0) {
		idNumber = cert.Certificate.SerialNumber
	}

	// If certificate only, ensure the ID comes before the primary.
	if privateKey == nil && k.Primary.Certificate.Certificate.SerialNumber.Cmp(idNumber) <= 0 {
		idNumber = big.NewInt(0)
		for k.Items[idNumber.String()] != nil {
			idNumber.Add(idNumber, big.NewInt(1))
		}
	}

	ki := &KeysetItem{
		Id:          idNumber.String(),
		Certificate: cert,
		PrivateKey:  privateKey,
	}
	k.Items[ki.Id] = ki
	if primary {
		k.Primary = ki
	}

	return nil
}
