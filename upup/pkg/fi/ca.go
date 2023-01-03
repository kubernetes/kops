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
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

const CertificateIDCA = "kubernetes-ca"

const (
	// SecretNameSSHPrimary is the Name for the primary SSH key
	SecretNameSSHPrimary = "admin"
)

const (
	keysetFormatLatest = "v1alpha2"
)

// Keyset is a parsed api.Keyset.
type Keyset struct {
	// LegacyFormat instructs a keypair task to convert a Legacy Keyset to the new Keyset API format.
	LegacyFormat bool
	Items        map[string]*KeysetItem

	// Primary is the KeysetItem that is considered the "active" key.
	// It is guaranteed to be non-nil, if there are any keypairs.
	Primary *KeysetItem
}

// KeysetItem is a certificate/key pair in a Keyset.
type KeysetItem struct {
	// Id is the identifier of this keypair.
	Id string
	// DistrustTimestamp is RFC 3339 date and time at which this keypair was distrusted.
	// If not set, keypair is trusted.
	DistrustTimestamp *time.Time
	// Certificate is the keypair's certificate.
	Certificate *pki.Certificate
	// PrivateKey is a reference to the keypair's private key.
	PrivateKey *pki.PrivateKey
}

// KeystoreReader contains just the functions we need to consume keypairs, not to update them.
type KeystoreReader interface {
	// FindKeyset finds a Keyset.  If the keyset is not found, it returns (nil, nil).
	FindKeyset(ctx context.Context, name string) (*Keyset, error)
}

// Keystore contains just the functions we need to issue keypairs, not to list / manage them

type Keystore interface {
	KeystoreReader

	// StoreKeyset writes a Keyset to the store.
	StoreKeyset(ctx context.Context, name string, keyset *Keyset) error

	// MirrorTo will copy secrets to a vfs.Path, which is often easier for a machine to read
	MirrorTo(ctx context.Context, basedir vfs.Path) error
}

// HasVFSPath is implemented by keystore & other stores that use a VFS path as their backing store
type HasVFSPath interface {
	VFSPath() vfs.Path
}

type CAStore interface {
	Keystore

	// ListKeysets will return all the KeySets.
	ListKeysets() (map[string]*Keyset, error)
}

// SSHCredentialStore holds SSHCredential objects
type SSHCredentialStore interface {
	// DeleteSSHCredential deletes the specified SSH credential.
	DeleteSSHCredential() error

	// AddSSHPublicKey adds an SSH public key.
	AddSSHPublicKey(ctx context.Context, data []byte) error

	// FindSSHPublicKeys retrieves the SSH public keys.
	FindSSHPublicKeys() ([]*kops.SSHCredential, error)
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

func (k *Keyset) ToCertificateBytes() ([]byte, error) {
	keys := make([]string, 0, len(k.Items))
	for k, item := range k.Items {
		if item.DistrustTimestamp == nil {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return KeysetItemIdOlder(k.Items[keys[i]].Id, k.Items[keys[j]].Id)
	})

	buf := new(bytes.Buffer)
	for _, key := range keys {
		item := k.Items[key]
		if item.Certificate != nil {
			certificate, err := item.Certificate.AsBytes()
			if err != nil {
				return nil, fmt.Errorf("public key %s: %v", item.Id, err)
			}
			buf.Write(certificate)
		}
	}
	return buf.Bytes(), nil
}

func (k *Keyset) ToPublicKeys() (string, error) {
	keys := make([]string, 0, len(k.Items))
	for k, item := range k.Items {
		if item.DistrustTimestamp == nil {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return KeysetItemIdOlder(k.Items[keys[i]].Id, k.Items[keys[j]].Id)
	})

	buf := new(strings.Builder)
	for _, key := range keys {
		item := k.Items[key]
		if item.Certificate != nil {
			publicKeyData, err := x509.MarshalPKIXPublicKey(item.Certificate.PublicKey)
			if err != nil {
				return "", fmt.Errorf("marshalling public key %s: %v", item.Id, err)
			}
			if err = pem.Encode(buf, &pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyData}); err != nil {
				return "", fmt.Errorf("encoding public key %s: %v", item.Id, err)
			}
		}
	}
	return buf.String(), nil
}

// NewKeyset creates a Keyset.
func NewKeyset(cert *pki.Certificate, privateKey *pki.PrivateKey) (*Keyset, error) {
	keyset := &Keyset{
		Items: map[string]*KeysetItem{},
	}
	_, err := keyset.AddItem(cert, privateKey, true)
	if err != nil {
		return nil, err
	}

	return keyset, nil
}

// AddItem adds an item to the keyset
func (k *Keyset) AddItem(cert *pki.Certificate, privateKey *pki.PrivateKey, primary bool) (item *KeysetItem, err error) {
	if cert == nil {
		return item, fmt.Errorf("no certificate provided")
	}
	if privateKey == nil && primary {
		return item, fmt.Errorf("private key not provided for primary item")
	}

	if !primary && k.Primary == nil {
		return item, fmt.Errorf("cannot add secondary item when no existing primary item")
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

	return ki, nil
}

type pkiKeystoreAdapter struct {
	reader KeystoreReader
}

func (p pkiKeystoreAdapter) FindPrimaryKeypair(ctx context.Context, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	keyset, err := p.reader.FindKeyset(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	if keyset == nil || keyset.Primary == nil {
		return nil, nil, nil
	}
	return keyset.Primary.Certificate, keyset.Primary.PrivateKey, nil
}

func NewPKIKeystoreAdapter(reader KeystoreReader) pki.Keystore {
	return &pkiKeystoreAdapter{reader: reader}
}
