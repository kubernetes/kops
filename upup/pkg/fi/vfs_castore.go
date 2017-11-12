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
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/sshcredentials"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSCAStore struct {
	basedir vfs.Path
	cluster *kops.Cluster

	mutex     sync.Mutex
	cachedCAs map[string]*cachedEntry
}

type cachedEntry struct {
	certificates *keyset
	privateKeys  *keyset
}

var _ CAStore = &VFSCAStore{}
var _ SSHCredentialStore = &VFSCAStore{}

func NewVFSCAStore(cluster *kops.Cluster, basedir vfs.Path) CAStore {
	c := &VFSCAStore{
		basedir:   basedir,
		cluster:   cluster,
		cachedCAs: make(map[string]*cachedEntry),
	}

	return c
}

// NewVFSSSHCredentialStore creates a SSHCredentialStore backed by VFS
func NewVFSSSHCredentialStore(cluster *kops.Cluster, basedir vfs.Path) SSHCredentialStore {
	// Note currently identical to NewVFSCAStore
	c := &VFSCAStore{
		basedir:   basedir,
		cluster:   cluster,
		cachedCAs: make(map[string]*cachedEntry),
	}

	return c
}

func (s *VFSCAStore) VFSPath() vfs.Path {
	return s.basedir
}

// Retrieves the CA keypair.  No longer generates keypairs if not found.
func (s *VFSCAStore) readCAKeypairs(id string) (*keyset, *keyset, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cached := s.cachedCAs[id]
	if cached != nil {
		return cached.certificates, cached.privateKeys, nil
	}

	caCertificates, err := s.loadCertificates(s.buildCertificatePoolPath(id))
	if err != nil {
		return nil, nil, err
	}

	var caPrivateKeys *keyset

	if caCertificates != nil {
		caPrivateKeys, err = s.loadPrivateKeys(s.buildPrivateKeyPoolPath(id))
		if err != nil {
			return nil, nil, err
		}

		if caPrivateKeys == nil {
			glog.Warningf("CA private key was not found; will generate new key")
			//return nil, fmt.Errorf("error loading CA private key - key not found")
		}
	}

	if caPrivateKeys == nil {
		// We no longer generate CA certificates automatically - too race-prone
		return caCertificates, caPrivateKeys, nil
	}

	cached = &cachedEntry{certificates: caCertificates, privateKeys: caPrivateKeys}
	s.cachedCAs[id] = cached

	return cached.certificates, cached.privateKeys, nil

}

func BuildCAX509Template() *x509.Certificate {
	subject := &pkix.Name{
		CommonName: "kubernetes",
	}

	template := &x509.Certificate{
		Subject:               *subject,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		IsCA: true,
	}
	return template
}

// Creates and stores CA keypair
// Should be called with the mutex held, to prevent concurrent creation of different keys
func (c *VFSCAStore) generateCACertificate(id string) (*keyset, *keyset, error) {
	template := BuildCAX509Template()

	caRsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	caPrivateKey := &pki.PrivateKey{Key: caRsaKey}

	caCertificate, err := pki.SignNewCertificate(caPrivateKey, template, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	t := time.Now().UnixNano()
	serial := pki.BuildPKISerial(t)

	keyPath := c.buildPrivateKeyPath(id, serial)
	err = c.storePrivateKey(caPrivateKey, keyPath)
	if err != nil {
		return nil, nil, err
	}

	// Make double-sure it round-trips
	privateKeys, err := c.loadPrivateKeys(c.buildPrivateKeyPoolPath(id))
	if err != nil {
		return nil, nil, err
	}
	if privateKeys == nil || privateKeys.primary == nil || privateKeys.primary.id != serial.String() {
		return nil, nil, fmt.Errorf("failed to round-trip CA private key")
	}

	certPath := c.buildCertificatePath(id, serial)
	err = c.storeCertificate(caCertificate, certPath)
	if err != nil {
		return nil, nil, err
	}

	// Make double-sure it round-trips
	certificates, err := c.loadCertificates(c.buildCertificatePoolPath(id))
	if err != nil {
		return nil, nil, err
	}

	if certificates == nil || certificates.primary == nil || certificates.primary.id != serial.String() {
		return nil, nil, fmt.Errorf("failed to round-trip CA certifiacate")
	}

	return certificates, privateKeys, nil
}

func (c *VFSCAStore) buildCertificatePoolPath(id string) vfs.Path {
	return c.basedir.Join("issued", id)
}

func (c *VFSCAStore) buildCertificatePath(id string, serial *big.Int) vfs.Path {
	return c.basedir.Join("issued", id, serial.String()+".crt")
}

func (c *VFSCAStore) buildPrivateKeyPoolPath(id string) vfs.Path {
	return c.basedir.Join("private", id)
}

func (c *VFSCAStore) buildPrivateKeyPath(id string, serial *big.Int) vfs.Path {
	return c.basedir.Join("private", id, serial.String()+".key")
}

func (c *VFSCAStore) loadCertificates(p vfs.Path) (*keyset, error) {
	keyset := &keyset{
		items: make(map[string]*keysetItem),
	}

	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, f := range files {
		cert, err := c.loadOneCertificate(f)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate %q: %v", f, err)
		}

		id := f.Base()
		id = strings.TrimSuffix(id, ".crt")
		keyset.items[id] = &keysetItem{
			id:          id,
			certificate: cert,
		}
	}

	if len(keyset.items) == 0 {
		return nil, nil
	}

	keyset.primary = keyset.findPrimary()

	return keyset, nil
}

func (c *VFSCAStore) loadOneCertificate(p vfs.Path) (*pki.Certificate, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	cert, err := pki.LoadPEMCertificate(data)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}
	return cert, nil
}

func (c *VFSCAStore) CertificatePool(id string, createIfMissing bool) (*CertificatePool, error) {
	cert, err := c.FindCertificatePool(id)
	if err == nil && cert == nil {
		if !createIfMissing {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &CertificatePool{}, err
		}
		return nil, fmt.Errorf("cannot find certificate pool %q", id)
	}
	return cert, err

}

func (c *VFSCAStore) FindKeypair(id string) (*pki.Certificate, *pki.PrivateKey, error) {
	cert, err := c.FindCert(id)
	if err != nil {
		return nil, nil, err
	}

	key, err := c.FindPrivateKey(id)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func (c *VFSCAStore) FindCert(id string) (*pki.Certificate, error) {
	var certs *keyset

	var err error
	p := c.buildCertificatePoolPath(id)
	certs, err = c.loadCertificates(p)
	if err != nil {
		return nil, fmt.Errorf("error in 'FindCert' attempting to load cert %q: %v", id, err)
	}

	var cert *pki.Certificate
	if certs != nil && certs.primary != nil {
		cert = certs.primary.certificate
	}

	return cert, nil
}

func (c *VFSCAStore) FindCertificatePool(id string) (*CertificatePool, error) {
	var certs *keyset

	var err error
	p := c.buildCertificatePoolPath(id)
	certs, err = c.loadCertificates(p)
	if err != nil {
		return nil, fmt.Errorf("error in 'FindCertificatePool' attempting to load cert %q: %v", id, err)
	}

	pool := &CertificatePool{}

	if certs != nil {
		if certs.primary != nil {
			pool.Primary = certs.primary.certificate
		}

		for k, cert := range certs.items {
			if certs.primary != nil && k == certs.primary.id {
				continue
			}
			if cert.certificate == nil {
				continue
			}
			pool.Secondary = append(pool.Secondary, cert.certificate)
		}
	}
	return pool, nil
}

// ListKeysets implements CAStore::ListKeysets
func (c *VFSCAStore) ListKeysets() ([]*kops.Keyset, error) {
	keysets := make(map[string]*kops.Keyset)

	{
		baseDir := c.basedir.Join("issued")
		files, err := baseDir.ReadTree()
		if err != nil {
			return nil, fmt.Errorf("error reading directory %q: %v", baseDir, err)
		}

		for _, f := range files {
			relativePath, err := vfs.RelativePath(baseDir, f)
			if err != nil {
				return nil, err
			}

			tokens := strings.Split(relativePath, "/")
			if len(tokens) != 2 {
				glog.V(2).Infof("ignoring unexpected file in keystore: %q", f)
				continue
			}

			name := tokens[0]
			keyset := keysets[name]
			if keyset == nil {
				keyset = &kops.Keyset{}
				keyset.Name = tokens[0]
				keyset.Spec.Type = kops.SecretTypeKeypair
				keysets[name] = keyset
			}

			keyset.Spec.Keys = append(keyset.Spec.Keys, kops.KeysetItem{
				Id: strings.TrimSuffix(tokens[1], ".crt"),
			})
		}
	}

	var items []*kops.Keyset
	for _, v := range keysets {
		items = append(items, v)
	}
	return items, nil
}

// ListSSHCredentials implements SSHCredentialStore::ListSSHCredentials
func (c *VFSCAStore) ListSSHCredentials() ([]*kops.SSHCredential, error) {
	var items []*kops.SSHCredential

	{
		baseDir := c.basedir.Join("ssh", "public")
		files, err := baseDir.ReadTree()
		if err != nil {
			return nil, fmt.Errorf("error reading directory %q: %v", baseDir, err)
		}

		for _, f := range files {
			relativePath, err := vfs.RelativePath(baseDir, f)
			if err != nil {
				return nil, err
			}

			tokens := strings.Split(relativePath, "/")
			if len(tokens) != 2 {
				glog.V(2).Infof("ignoring unexpected file in keystore: %q", f)
				continue
			}

			pubkey, err := f.ReadFile()
			if err != nil {
				return nil, fmt.Errorf("error reading SSH credential %q: %v", f, err)
			}

			item := &kops.SSHCredential{}
			item.Name = tokens[0]
			item.Spec.PublicKey = string(pubkey)
			items = append(items, item)
		}
	}

	return items, nil
}

// MirrorTo will copy keys to a vfs.Path, which is often easier for a machine to read
func (c *VFSCAStore) MirrorTo(basedir vfs.Path) error {
	if basedir.Path() == c.basedir.Path() {
		return nil
	}
	glog.V(2).Infof("Mirroring key store from %q to %q", c.basedir, basedir)

	aclOracle := func(p vfs.Path) (vfs.ACL, error) {
		return acls.GetACL(p, c.cluster)
	}
	return vfs.CopyTree(c.basedir, basedir, aclOracle)
}

func (c *VFSCAStore) IssueCert(signer string, id string, serial *big.Int, privateKey *pki.PrivateKey, template *x509.Certificate) (*pki.Certificate, error) {
	glog.Infof("Issuing new certificate: %q", id)

	template.SerialNumber = serial

	var cert *pki.Certificate
	if template.IsCA {
		var err error
		cert, err = pki.SignNewCertificate(privateKey, template, nil, nil)
		if err != nil {
			return nil, err
		}
	} else {
		caCertificates, caPrivateKeys, err := c.readCAKeypairs(signer)
		if err != nil {
			return nil, err
		}

		if caPrivateKeys == nil || caPrivateKeys.primary == nil || caPrivateKeys.primary.privateKey == nil {
			return nil, fmt.Errorf("ca key for %q was not found; cannot issue certificates", signer)
		}
		if caCertificates == nil || caCertificates.primary == nil || caCertificates.primary.certificate == nil {
			return nil, fmt.Errorf("ca certificate for %q was not found; cannot issue certificates", signer)
		}
		cert, err = pki.SignNewCertificate(privateKey, template, caCertificates.primary.certificate.Certificate, caPrivateKeys.primary.privateKey)
		if err != nil {
			return nil, err
		}
	}

	err := c.StoreKeypair(id, cert, privateKey)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	p := c.buildCertificatePath(id, serial)
	return c.loadOneCertificate(p)
}

func (c *VFSCAStore) StoreKeypair(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	serial := cert.Certificate.SerialNumber

	{
		p := c.buildPrivateKeyPath(id, serial)
		err := c.storePrivateKey(privateKey, p)
		if err != nil {
			return err
		}
	}

	{
		p := c.buildCertificatePath(id, serial)
		err := c.storeCertificate(cert, p)
		if err != nil {
			// TODO: Delete private key?
			return err
		}
	}

	return nil
}

func (c *VFSCAStore) AddCert(id string, cert *pki.Certificate) error {
	glog.Infof("Adding TLS certificate: %q", id)

	// We add with a timestamp of zero so this will never be the newest cert
	serial := pki.BuildPKISerial(0)

	p := c.buildCertificatePath(id, serial)

	err := c.storeCertificate(cert, p)
	if err != nil {
		return err
	}

	// Make double-sure it round-trips
	_, err = c.loadOneCertificate(p)
	return err
}

func (c *VFSCAStore) loadPrivateKeys(p vfs.Path) (*keyset, error) {
	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	keys := &keyset{
		items: make(map[string]*keysetItem),
	}

	for _, f := range files {
		privateKey, err := c.loadOnePrivateKey(f)
		if err != nil {
			return nil, fmt.Errorf("error loading private key %q: %v", f, err)
		}
		id := f.Base()
		id = strings.TrimSuffix(id, ".key")
		keys.items[id] = &keysetItem{
			id:         id,
			privateKey: privateKey,
		}
	}

	if len(keys.items) == 0 {
		return nil, nil
	}

	keys.primary = keys.findPrimary()

	return keys, nil
}

func (c *VFSCAStore) loadOnePrivateKey(p vfs.Path) (*pki.PrivateKey, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	k, err := pki.ParsePEMPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key from %q: %v", p, err)
	}
	return k, err
}

func (c *VFSCAStore) FindPrivateKey(id string) (*pki.PrivateKey, error) {
	var keys *keyset
	if id == CertificateId_CA {
		_, caPrivateKeys, err := c.readCAKeypairs(id)
		if err != nil {
			return nil, err
		}
		keys = caPrivateKeys
	} else {
		var err error
		p := c.buildPrivateKeyPoolPath(id)
		keys, err = c.loadPrivateKeys(p)
		if err != nil {
			return nil, err
		}

	}

	var key *pki.PrivateKey
	if keys != nil && keys.primary != nil {
		key = keys.primary.privateKey
	}
	return key, nil
}

func (c *VFSCAStore) CreateKeypair(signer string, id string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	serial := c.buildSerial()

	cert, err := c.IssueCert(signer, id, serial, privateKey, template)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *VFSCAStore) storePrivateKey(privateKey *pki.PrivateKey, p vfs.Path) error {
	var data bytes.Buffer
	_, err := privateKey.WriteTo(&data)
	if err != nil {
		return err
	}

	acl, err := acls.GetACL(p, c.cluster)
	if err != nil {
		return err
	}
	return p.WriteFile(data.Bytes(), acl)
}

func (c *VFSCAStore) storeCertificate(cert *pki.Certificate, p vfs.Path) error {
	// TODO: replace storePrivateKey & storeCertificate with writeFile(io.WriterTo)?
	var data bytes.Buffer
	_, err := cert.WriteTo(&data)
	if err != nil {
		return err
	}

	acl, err := acls.GetACL(p, c.cluster)
	if err != nil {
		return err
	}
	return p.WriteFile(data.Bytes(), acl)
}

func (c *VFSCAStore) buildSerial() *big.Int {
	t := time.Now().UnixNano()
	return pki.BuildPKISerial(t)
}

// AddSSHPublicKey stores an SSH public key
func (c *VFSCAStore) AddSSHPublicKey(name string, pubkey []byte) error {
	id, err := sshcredentials.Fingerprint(string(pubkey))
	if err != nil {
		return fmt.Errorf("error fingerprinting SSH public key %q: %v", name, err)
	}

	p := c.buildSSHPublicKeyPath(name, id)

	acl, err := acls.GetACL(p, c.cluster)
	if err != nil {
		return err
	}

	return p.WriteFile(pubkey, acl)
}

func (c *VFSCAStore) buildSSHPublicKeyPath(name string, id string) vfs.Path {
	// id is fingerprint with colons, but we store without colons
	id = strings.Replace(id, ":", "", -1)
	return c.basedir.Join("ssh", "public", name, id)
}

func (c *VFSCAStore) FindSSHPublicKeys(name string) ([]*kops.SSHCredential, error) {
	p := c.basedir.Join("ssh", "public", name)

	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []*kops.SSHCredential

	for _, f := range files {
		data, err := f.ReadFile()
		if err != nil {
			if os.IsNotExist(err) {
				glog.V(2).Infof("Ignoring not-found issue reading %q", f)
				continue
			}
			return nil, fmt.Errorf("error loading SSH item %q: %v", f, err)
		}

		item := &kops.SSHCredential{}
		item.Name = name
		item.Spec.PublicKey = string(data)
		items = append(items, item)
	}

	return items, nil
}

func (c *VFSCAStore) loadData(p vfs.Path) (*pki.PrivateKey, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	k, err := pki.ParsePEMPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key from %q: %v", p, err)
	}
	return k, err
}

// DeleteKeysetItem implements CAStore::DeleteKeysetItem
func (c *VFSCAStore) DeleteKeysetItem(item *kops.Keyset, id string) error {
	switch item.Spec.Type {
	case kops.SecretTypeKeypair:
		version, ok := big.NewInt(0).SetString(id, 10)
		if !ok {
			return fmt.Errorf("keypair had non-integer version: %q", id)
		}
		p := c.buildCertificatePath(item.Name, version)
		if err := p.Remove(); err != nil {
			return fmt.Errorf("error deleting certificate: %v", err)
		}
		p = c.buildPrivateKeyPath(item.Name, version)
		if err := p.Remove(); err != nil {
			return fmt.Errorf("error deleting private key: %v", err)
		}
		return nil

	default:
		// Primarily because we need to make sure users can recreate them!
		return fmt.Errorf("deletion of keystore items of type %v not (yet) supported", item.Spec.Type)
	}
}

func (c *VFSCAStore) DeleteSSHCredential(item *kops.SSHCredential) error {
	if item.Spec.PublicKey == "" {
		return fmt.Errorf("must specific public key to delete SSHCredential")
	}
	id, err := sshcredentials.Fingerprint(item.Spec.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid PublicKey when deleting SSHCredential: %v", err)
	}
	p := c.buildSSHPublicKeyPath(item.Name, id)
	return p.Remove()
}
