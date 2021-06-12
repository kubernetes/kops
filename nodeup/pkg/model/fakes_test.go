/*
Copyright 2020 The Kubernetes Authors.

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
	"crypto/x509"
	"fmt"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// fakeCAStore mocks out some of fi.CAStore, for our tests.
type fakeCAStore struct {
	T              *testing.T
	privateKeysets map[string]*kops.Keyset
	certs          map[string]*pki.Certificate
}

var _ fi.CAStore = &fakeCAStore{}

func (k fakeCAStore) FindPrimaryKeypair(name string) (*pki.Certificate, *pki.PrivateKey, error) {
	panic("fakeCAStore does not implement FindPrimaryKeypair")
}

func (k fakeCAStore) FindKeyset(name string) (*fi.Keyset, error) {
	kopsKeyset := k.privateKeysets[name]
	if kopsKeyset == nil {
		return nil, nil
	}

	keyset := &fi.Keyset{
		Items: make(map[string]*fi.KeysetItem),
	}

	for _, key := range kopsKeyset.Spec.Keys {
		ki := &fi.KeysetItem{
			Id: key.Id,
		}
		if len(key.PublicMaterial) != 0 {
			cert, err := pki.ParsePEMCertificate(key.PublicMaterial)
			if err != nil {
				return nil, fmt.Errorf("error loading certificate %s/%s: %v", name, key.Id, err)
			}
			ki.Certificate = cert
		}

		if len(key.PrivateMaterial) != 0 {
			privateKey, err := pki.ParsePEMPrivateKey(key.PrivateMaterial)
			if err != nil {
				return nil, fmt.Errorf("error loading private key %s/%s: %v", name, key.Id, err)
			}
			ki.PrivateKey = privateKey
		}

		keyset.Items[key.Id] = ki
	}

	keyset.Primary = keyset.Items[fi.FindPrimary(kopsKeyset).Id]

	return keyset, nil
}

func (k fakeCAStore) CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	panic("fakeCAStore does not implement CreateKeypair")
}

func (k fakeCAStore) StoreKeyset(name string, keyset *fi.Keyset) error {
	panic("fakeCAStore does not implement StoreKeyset")
}

func (k fakeCAStore) MirrorTo(basedir vfs.Path) error {
	panic("fakeCAStore does not implement MirrorTo")
}

func (k fakeCAStore) FindCertificatePool(name string) (*fi.CertificatePool, error) {
	panic("fakeCAStore does not implement FindCertificatePool")
}

func (k fakeCAStore) FindCertificateKeyset(name string) (*kops.Keyset, error) {
	panic("fakeCAStore does not implement FindCertificateKeyset")
}

func (k fakeCAStore) FindPrivateKey(name string) (*pki.PrivateKey, error) {
	primaryId := k.privateKeysets[name].Spec.PrimaryId
	for _, item := range k.privateKeysets[name].Spec.Keys {
		if item.Id == primaryId {
			return pki.ParsePEMPrivateKey(item.PrivateMaterial)
		}
	}
	return nil, nil
}

func (k fakeCAStore) FindPrivateKeyset(name string) (*kops.Keyset, error) {
	return k.privateKeysets[name], nil
}

func (k fakeCAStore) FindCert(name string) (*pki.Certificate, error) {
	return k.certs[name], nil
}

func (k fakeCAStore) ListKeysets() ([]*kops.Keyset, error) {
	panic("fakeCAStore does not implement ListKeysets")
}

func (k fakeCAStore) DeleteKeysetItem(item *kops.Keyset, id string) error {
	panic("fakeCAStore does not implement DeleteKeysetItem")
}
