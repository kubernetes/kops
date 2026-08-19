/*
Copyright 2017 The Kubernetes Authors.

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

package fitasks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func TestKeypairDeps(t *testing.T) {
	ca := &Keypair{
		Name: new("ca"),
	}
	cert := &Keypair{
		Name:   new("cert"),
		Signer: ca,
	}

	tasks := make(map[string]fi.CloudupTask)
	tasks["ca"] = ca
	tasks["cert"] = cert

	deps := fi.FindTaskDependencies(tasks)

	if strings.Join(deps["ca"], ",") != "" {
		t.Errorf("unexpected dependencies for ca: %v", deps["ca"])
	}

	if strings.Join(deps["cert"], ",") != "ca" {
		t.Errorf("unexpected dependencies for cert: %v", deps["cert"])
	}
}

// fakeKeystore is a minimal in-memory fi.Keystore for exercising CreateKeyset
// without going through a real VFS-backed store.
type fakeKeystore struct {
	keysets map[string]*fi.Keyset
	stored  bool
}

func (f *fakeKeystore) FindKeyset(ctx context.Context, name string) (*fi.Keyset, error) {
	return f.keysets[name], nil
}

func (f *fakeKeystore) StoreKeyset(ctx context.Context, name string, keyset *fi.Keyset) error {
	f.stored = true
	if f.keysets == nil {
		f.keysets = map[string]*fi.Keyset{}
	}
	f.keysets[name] = keyset
	return nil
}

func (f *fakeKeystore) MirrorTo(ctx context.Context, basedir vfs.Path) error {
	return nil
}

var _ fi.Keystore = &fakeKeystore{}

// selfSignedTestCA builds a minimal self-signed CA certificate/key pair for tests.
func selfSignedTestCA(t *testing.T, commonName string) (*pki.Certificate, *pki.PrivateKey) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating test certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsing test certificate: %v", err)
	}

	return &pki.Certificate{Certificate: cert}, &pki.PrivateKey{Key: key}
}

func TestCreateKeysetRefusesToReissueExistingCA(t *testing.T) {
	ctx := context.Background()

	existingCert, existingKey := selfSignedTestCA(t, "kubernetes-ca")

	store := &fakeKeystore{
		keysets: map[string]*fi.Keyset{
			"kubernetes-ca": {
				Primary: &fi.KeysetItem{Id: "1", Certificate: existingCert, PrivateKey: existingKey},
				Items: map[string]*fi.KeysetItem{
					"1": {Id: "1", Certificate: existingCert, PrivateKey: existingKey},
				},
			},
		},
	}

	_, err := CreateKeyset(ctx, store, "kubernetes-ca", pki.IssueCertRequest{
		Type:    "ca",
		Subject: pkix.Name{CommonName: "kubernetes-ca"},
	})

	if err == nil {
		t.Fatal("expected CreateKeyset to refuse reissuing an existing CA, got nil error")
	}
	if !strings.Contains(err.Error(), "refusing to create keypair") {
		t.Errorf("unexpected error message: %v", err)
	}
	if store.stored {
		t.Error("CreateKeyset must not store a new keyset when refusing to reissue an existing CA")
	}
}

func TestCreateKeysetCreatesNewCAWhenNoneExists(t *testing.T) {
	ctx := context.Background()
	store := &fakeKeystore{}

	keyset, err := CreateKeyset(ctx, store, "kubernetes-ca", pki.IssueCertRequest{
		Type:    "ca",
		Subject: pkix.Name{CommonName: "kubernetes-ca"},
	})
	if err != nil {
		t.Fatalf("unexpected error creating a brand-new CA: %v", err)
	}
	if keyset == nil || keyset.Primary == nil || keyset.Primary.Certificate == nil {
		t.Fatal("expected a populated keyset for a brand-new CA")
	}
	if !store.stored {
		t.Error("expected CreateKeyset to store the newly created CA")
	}
}
