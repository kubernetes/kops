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
	"context"
	"crypto/x509"
	"fmt"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// fakeKeystore mocks out some of fi.Keystore, for our tests.
type fakeKeystore struct {
	T              *testing.T
	privateKeysets map[string]*kops.Keyset
}

var _ fi.Keystore = &fakeKeystore{}

func (k fakeKeystore) FindPrimaryKeypair(name string) (*pki.Certificate, *pki.PrivateKey, error) {
	keyset, err := k.FindKeyset(name)
	if err != nil {
		return nil, nil, err
	}

	return keyset.Primary.Certificate, keyset.Primary.PrivateKey, nil
}

func (k fakeKeystore) FindKeyset(name string) (*fi.Keyset, error) {
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

func (k fakeKeystore) CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	panic("fakeKeystore does not implement CreateKeypair")
}

func (k fakeKeystore) StoreKeyset(ctx context.Context, name string, keyset *fi.Keyset) error {
	panic("fakeKeystore does not implement StoreKeyset")
}

func (k fakeKeystore) MirrorTo(ctx context.Context, basedir vfs.Path) error {
	panic("fakeKeystore does not implement MirrorTo")
}
