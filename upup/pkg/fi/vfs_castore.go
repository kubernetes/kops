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
	"crypto/md5"
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
	"golang.org/x/crypto/ssh"

	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSCAStore struct {
	basedir vfs.Path
	cluster *kops.Cluster

	mutex     sync.Mutex
	cachedCAs map[string]*cachedEntry
}

type cachedEntry struct {
	certificates *certificates
	privateKeys  *privateKeys
}

var _ CAStore = &VFSCAStore{}

func NewVFSCAStore(cluster *kops.Cluster, basedir vfs.Path) CAStore {
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
func (s *VFSCAStore) readCAKeypairs(id string) (*certificates, *privateKeys, error) {
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

	var caPrivateKeys *privateKeys

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
func (c *VFSCAStore) generateCACertificate(id string) (*certificates, *privateKeys, error) {
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
	if privateKeys == nil || privateKeys.primary != serial.String() {
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

	if certificates == nil || certificates.primary != serial.String() {
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

type certificates struct {
	certificates map[string]*pki.Certificate
	primary      string
}

func (p *certificates) Primary() *pki.Certificate {
	if p.primary == "" {
		return nil
	}
	return p.certificates[p.primary]
}

func (c *VFSCAStore) loadCertificates(p vfs.Path) (*certificates, error) {
	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	certs := &certificates{
		certificates: make(map[string]*pki.Certificate),
	}

	for _, f := range files {
		name := f.Base()
		if strings.HasSuffix(name, ".yaml") {
			// ignore bundle
			continue
		}

		cert, err := c.loadOneCertificate(f)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate %q: %v", f, err)
		}
		name = strings.TrimSuffix(name, ".crt")
		certs.certificates[name] = cert
	}

	if len(certs.certificates) == 0 {
		return nil, nil
	}

	var primaryVersion *big.Int
	for k := range certs.certificates {
		version, ok := big.NewInt(0).SetString(k, 10)
		if !ok {
			glog.Warningf("Ignoring certificate with non-integer version: %q", k)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			certs.primary = k
			primaryVersion = version
		}
	}

	return certs, nil
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

func (c *VFSCAStore) Cert(id string, createIfMissing bool) (*pki.Certificate, error) {
	cert, err := c.FindCert(id)
	if err == nil && cert == nil {
		if !createIfMissing {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &pki.Certificate{}, err
		}
		return nil, fmt.Errorf("cannot find certificate %q", id)
	}
	return cert, err

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
	var certs *certificates

	var err error
	p := c.buildCertificatePoolPath(id)
	certs, err = c.loadCertificates(p)
	if err != nil {
		return nil, fmt.Errorf("error in 'FindCert' attempting to load cert %q: %v", id, err)
	}

	var cert *pki.Certificate
	if certs != nil && certs.primary != "" {
		cert = certs.certificates[certs.primary]
	}

	return cert, nil
}

func (c *VFSCAStore) FindCertificatePool(id string) (*CertificatePool, error) {
	var certs *certificates

	var err error
	p := c.buildCertificatePoolPath(id)
	certs, err = c.loadCertificates(p)
	if err != nil {
		return nil, fmt.Errorf("error in 'FindCertificatePool' attempting to load cert %q: %v", id, err)
	}

	pool := &CertificatePool{}

	if certs != nil {
		pool.Primary = certs.Primary()

		for k, cert := range certs.certificates {
			if k == certs.primary {
				continue
			}
			pool.Secondary = append(pool.Secondary, cert)
		}
	}
	return pool, nil
}

func (c *VFSCAStore) List() ([]*KeystoreItem, error) {
	var items []*KeystoreItem

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

			item := &KeystoreItem{
				Name: tokens[0],
				Id:   strings.TrimSuffix(tokens[1], ".crt"),
				Type: SecretTypeKeypair,
			}
			items = append(items, item)
		}
	}

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

			item := &KeystoreItem{
				Name: tokens[0],
				Id:   insertFingerprintColons(tokens[1]),
				Type: SecretTypeSSHPublicKey,
			}
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

		if caPrivateKeys == nil || caPrivateKeys.Primary() == nil {
			return nil, fmt.Errorf("ca key for %q was not found; cannot issue certificates", signer)
		}
		cert, err = pki.SignNewCertificate(privateKey, template, caCertificates.Primary().Certificate, caPrivateKeys.Primary())
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

type privateKeys struct {
	keys    map[string]*pki.PrivateKey
	primary string
}

func (p *privateKeys) Primary() *pki.PrivateKey {
	if p.primary == "" {
		return nil
	}
	return p.keys[p.primary]
}

func (c *VFSCAStore) loadPrivateKeys(p vfs.Path) (*privateKeys, error) {
	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	keys := &privateKeys{
		keys: make(map[string]*pki.PrivateKey),
	}

	for _, f := range files {
		name := f.Base()
		if strings.HasSuffix(name, ".yaml") {
			// ignore bundle
			continue
		}

		key, err := c.loadOnePrivateKey(f)
		if err != nil {
			return nil, fmt.Errorf("error loading private key %q: %v", f, err)
		}
		name = strings.TrimSuffix(name, ".key")
		keys.keys[name] = key
	}

	if len(keys.keys) == 0 {
		return nil, nil
	}

	var primaryVersion *big.Int
	for k := range keys.keys {
		version, ok := big.NewInt(0).SetString(k, 10)
		if !ok {
			glog.Warningf("Ignoring private key with non-integer version: %q", k)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			keys.primary = k
			primaryVersion = version
		}
	}

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
	var keys *privateKeys
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
	if keys != nil && keys.primary != "" {
		key = keys.keys[keys.primary]
	}
	return key, nil
}

func (c *VFSCAStore) PrivateKey(id string, createIfMissing bool) (*pki.PrivateKey, error) {
	key, err := c.FindPrivateKey(id)
	if err == nil && key == nil {
		if !createIfMissing {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &pki.PrivateKey{}, err
		}
		return nil, fmt.Errorf("cannot find SSL key %q", id)
	}
	return key, err

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

func formatFingerprint(data []byte) string {
	var buf bytes.Buffer

	for i, b := range data {
		s := fmt.Sprintf("%0.2x", b)
		if i != 0 {
			buf.WriteString(":")
		}
		buf.WriteString(s)
	}
	return buf.String()
}

func insertFingerprintColons(id string) string {
	remaining := id

	var buf bytes.Buffer
	for {
		if remaining == "" {
			break
		}
		if buf.Len() != 0 {
			buf.WriteString(":")
		}
		if len(remaining) < 2 {
			glog.Warningf("unexpected format for SSH public key id: %q", id)
			buf.WriteString(remaining)
			break
		} else {
			buf.WriteString(remaining[0:2])
			remaining = remaining[2:]
		}
	}
	return buf.String()
}

// AddSSHPublicKey stores an SSH public key
func (c *VFSCAStore) AddSSHPublicKey(name string, pubkey []byte) error {
	var id string
	{
		sshPublicKey, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
		if err != nil {
			return fmt.Errorf("error parsing public key: %v", err)
		}

		// compute fingerprint to serve as id
		h := md5.New()
		_, err = h.Write(sshPublicKey.Marshal())
		if err != nil {
			return err
		}
		id = formatFingerprint(h.Sum(nil))
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

func (c *VFSCAStore) FindSSHPublicKeys(name string) ([]*KeystoreItem, error) {
	p := c.basedir.Join("ssh", "public", name)

	items, err := c.loadPath(p)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		// Fill in the missing fields
		item.Type = SecretTypeSSHPublicKey
		item.Name = name

		item.Id = insertFingerprintColons(item.Id)
	}
	return items, nil
}

func (c *VFSCAStore) loadPath(p vfs.Path) ([]*KeystoreItem, error) {
	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var keystoreItems []*KeystoreItem

	for _, f := range files {
		data, err := f.ReadFile()
		if err != nil {
			if os.IsNotExist(err) {
				glog.V(2).Infof("Ignoring not-found issue reading %q", f)
				continue
			}
			return nil, fmt.Errorf("error loading keystore item %q: %v", f, err)
		}
		name := f.Base()
		keystoreItem := &KeystoreItem{
			Id:   name,
			Data: data,
		}
		keystoreItems = append(keystoreItems, keystoreItem)
	}

	return keystoreItems, nil
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

func (c *VFSCAStore) DeleteSecret(item *KeystoreItem) error {
	switch item.Type {
	case SecretTypeSSHPublicKey:
		p := c.buildSSHPublicKeyPath(item.Name, item.Id)
		return p.Remove()

	case SecretTypeKeypair:
		version, ok := big.NewInt(0).SetString(item.Id, 10)
		if !ok {
			return fmt.Errorf("keypair had non-integer version: %q", item.Id)
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
		return fmt.Errorf("deletion of keystore items of type %v not (yet) supported", item.Type)
	}
}
