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
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"k8s.io/kops/util/pkg/vfs"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"
)

type VFSCAStore struct {
	DryRun  bool
	basedir vfs.Path

	mutex               sync.Mutex
	cacheCaCertificates *certificates
	cacheCaPrivateKeys  *privateKeys
}

var _ CAStore = &VFSCAStore{}

func NewVFSCAStore(basedir vfs.Path) CAStore {
	c := &VFSCAStore{
		basedir: basedir,
	}

	return c
}

func (s *VFSCAStore) VFSPath() vfs.Path {
	return s.basedir
}

// Retrieves the CA keypair, generating a new keypair if not found
func (s *VFSCAStore) readCAKeypairs() (*certificates, *privateKeys, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cacheCaPrivateKeys != nil {
		return s.cacheCaCertificates, s.cacheCaPrivateKeys, nil
	}

	caCertificates, err := s.loadCertificates(s.buildCertificatePoolPath(CertificateId_CA))
	if err != nil {
		return nil, nil, err
	}

	var caPrivateKeys *privateKeys

	if caCertificates != nil {
		caPrivateKeys, err = s.loadPrivateKeys(s.buildPrivateKeyPoolPath(CertificateId_CA))
		if err != nil {
			return nil, nil, err
		}

		if caPrivateKeys == nil {
			glog.Warningf("CA private key was not found; will generate new key")
			//return nil, fmt.Errorf("error loading CA private key - key not found")
		}
	}

	if caPrivateKeys == nil {
		caCertificates, caPrivateKeys, err = s.generateCACertificate()
		if err != nil {
			return nil, nil, err
		}

	}
	s.cacheCaCertificates = caCertificates
	s.cacheCaPrivateKeys = caPrivateKeys

	return s.cacheCaCertificates, s.cacheCaPrivateKeys, nil
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
func (c *VFSCAStore) generateCACertificate() (*certificates, *privateKeys, error) {
	template := BuildCAX509Template()

	caRsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	caPrivateKey := &PrivateKey{Key: caRsaKey}

	caCertificate, err := SignNewCertificate(caPrivateKey, template, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	t := time.Now().UnixNano()
	serial := BuildPKISerial(t)

	keyPath := c.buildPrivateKeyPath(CertificateId_CA, serial)
	err = c.storePrivateKey(caPrivateKey, keyPath)
	if err != nil {
		return nil, nil, err
	}

	// Make double-sure it round-trips
	privateKeys, err := c.loadPrivateKeys(c.buildPrivateKeyPoolPath(CertificateId_CA))
	if err != nil {
		return nil, nil, err
	}
	if privateKeys == nil || privateKeys.primary != serial.String() {
		return nil, nil, fmt.Errorf("failed to round-trip CA private key")
	}

	certPath := c.buildCertificatePath(CertificateId_CA, serial)
	err = c.storeCertificate(caCertificate, certPath)
	if err != nil {
		return nil, nil, err
	}

	// Make double-sure it round-trips
	certificates, err := c.loadCertificates(c.buildCertificatePoolPath(CertificateId_CA))
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
	certificates map[string]*Certificate
	primary      string
}

func (p *certificates) Primary() *Certificate {
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
		certificates: make(map[string]*Certificate),
	}

	for _, f := range files {
		cert, err := c.loadOneCertificate(f)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate %q: %v", f, err)
		}
		name := f.Base()
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

func (c *VFSCAStore) loadOneCertificate(p vfs.Path) (*Certificate, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	cert, err := LoadPEMCertificate(data)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}
	return cert, nil
}

func (c *VFSCAStore) Cert(id string) (*Certificate, error) {
	cert, err := c.FindCert(id)
	if err == nil && cert == nil {
		if c.DryRun {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &Certificate{}, err
		}
		return nil, fmt.Errorf("cannot find certificate %q", id)
	}
	return cert, err

}

func (c *VFSCAStore) CertificatePool(id string) (*CertificatePool, error) {
	cert, err := c.FindCertificatePool(id)
	if err == nil && cert == nil {
		if c.DryRun {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &CertificatePool{}, err
		}
		return nil, fmt.Errorf("cannot find certificate pool %q", id)
	}
	return cert, err

}

func (c *VFSCAStore) FindKeypair(id string) (*Certificate, *PrivateKey, error) {
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

func (c *VFSCAStore) FindCert(id string) (*Certificate, error) {
	var certs *certificates

	if id == CertificateId_CA {
		caCertificates, _, err := c.readCAKeypairs()
		if err != nil {
			return nil, err
		}
		certs = caCertificates
	} else {
		var err error
		p := c.buildCertificatePoolPath(id)
		certs, err = c.loadCertificates(p)
		if err != nil {
			return nil, err
		}
	}

	var cert *Certificate
	if certs != nil && certs.primary != "" {
		cert = certs.certificates[certs.primary]
	}

	return cert, nil
}

func (c *VFSCAStore) FindCertificatePool(id string) (*CertificatePool, error) {
	var certs *certificates

	if id == CertificateId_CA {
		caCertificates, _, err := c.readCAKeypairs()
		if err != nil {
			return nil, err
		}
		certs = caCertificates
	} else {
		var err error
		p := c.buildCertificatePoolPath(id)
		certs, err = c.loadCertificates(p)
		if err != nil {
			return nil, err
		}
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

func (c *VFSCAStore) IssueCert(id string, serial *big.Int, privateKey *PrivateKey, template *x509.Certificate) (*Certificate, error) {
	glog.Infof("Issuing new certificate: %q", id)

	template.SerialNumber = serial

	caCertificates, caPrivateKeys, err := c.readCAKeypairs()
	if err != nil {
		return nil, err
	}

	if caPrivateKeys == nil || caPrivateKeys.Primary() == nil {
		return nil, fmt.Errorf("ca.key was not found; cannot issue certificates")
	}
	cert, err := SignNewCertificate(privateKey, template, caCertificates.Primary().Certificate, caPrivateKeys.Primary())
	if err != nil {
		return nil, err
	}

	err = c.StoreKeypair(id, cert, privateKey)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	p := c.buildCertificatePath(id, serial)
	return c.loadOneCertificate(p)
}

func (c *VFSCAStore) StoreKeypair(id string, cert *Certificate, privateKey *PrivateKey) error {
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

func (c *VFSCAStore) AddCert(id string, cert *Certificate) error {
	glog.Infof("Adding TLS certificate: %q", id)

	// We add with a timestamp of zero so this will never be the newest cert
	serial := BuildPKISerial(0)

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
	keys    map[string]*PrivateKey
	primary string
}

func (p *privateKeys) Primary() *PrivateKey {
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
		keys: make(map[string]*PrivateKey),
	}

	for _, f := range files {
		key, err := c.loadOnePrivateKey(f)
		if err != nil {
			return nil, fmt.Errorf("error loading private key %q: %v", f, err)
		}
		name := f.Base()
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

func (c *VFSCAStore) loadOnePrivateKey(p vfs.Path) (*PrivateKey, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	k, err := ParsePEMPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key from %q: %v", p, err)
	}
	return k, err
}

func ParsePEMPrivateKey(data []byte) (*PrivateKey, error) {
	k, err := parsePEMPrivateKey(data)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, nil
	}
	return &PrivateKey{Key: k}, nil
}

func (c *VFSCAStore) FindPrivateKey(id string) (*PrivateKey, error) {
	var keys *privateKeys
	if id == CertificateId_CA {
		_, caPrivateKeys, err := c.readCAKeypairs()
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

	var key *PrivateKey
	if keys != nil && keys.primary != "" {
		key = keys.keys[keys.primary]
	}
	return key, nil
}

func (c *VFSCAStore) PrivateKey(id string) (*PrivateKey, error) {
	key, err := c.FindPrivateKey(id)
	if err == nil && key == nil {
		if c.DryRun {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &PrivateKey{}, err
		}
		return nil, fmt.Errorf("cannot find SSL key %q", id)
	}
	return key, err

}

func (c *VFSCAStore) CreateKeypair(id string, template *x509.Certificate, privateKey *PrivateKey) (*Certificate, error) {
	serial := c.buildSerial()

	cert, err := c.IssueCert(id, serial, privateKey, template)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *VFSCAStore) storePrivateKey(privateKey *PrivateKey, p vfs.Path) error {
	var data bytes.Buffer
	_, err := privateKey.WriteTo(&data)
	if err != nil {
		return err
	}

	return p.WriteFile(data.Bytes())
}

func (c *VFSCAStore) storeCertificate(cert *Certificate, p vfs.Path) error {
	// TODO: replace storePrivateKey & storeCertificate with writeFile(io.WriterTo)?
	var data bytes.Buffer
	_, err := cert.WriteTo(&data)
	if err != nil {
		return err
	}

	return p.WriteFile(data.Bytes())
}

func (c *VFSCAStore) buildSerial() *big.Int {
	t := time.Now().UnixNano()
	return BuildPKISerial(t)
}

// BuildPKISerial produces a serial number for certs that is vanishingly unlikely to collide
// The timestamp should be provided as an input (time.Now().UnixNano()), and then we combine
// that with a 32 bit random crypto-rand integer.
// We also know that a bigger value was created later (modulo clock skew)
func BuildPKISerial(timestamp int64) *big.Int {
	randomLimit := new(big.Int).Lsh(big.NewInt(1), 32)
	randomComponent, err := crypto_rand.Int(crypto_rand.Reader, randomLimit)
	if err != nil {
		glog.Fatalf("error generating random number: %v", err)
	}

	serial := big.NewInt(timestamp)
	serial.Lsh(serial, 32)
	serial.Or(serial, randomComponent)

	return serial
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
	var buf bytes.Buffer

	for {
		if id == "" {
			break
		}
		if buf.Len() != 0 {
			buf.WriteString(":")
		}
		if len(id) < 2 {
			buf.WriteString(id)
		} else {
			buf.WriteString(id[0:2])
			id = id[2:]
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
	return c.storeData(pubkey, p)
}

func (c *VFSCAStore) buildSSHPublicKeyPath(name string, id string) vfs.Path {
	// id is fingerprint with colons, but we store without colons
	id = strings.Replace(id, ":", "", -1)
	return c.basedir.Join("ssh", "public", name, id)
}

func (c *VFSCAStore) storeData(data []byte, p vfs.Path) error {
	return p.WriteFile(data)
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

func (c *VFSCAStore) loadData(p vfs.Path) (*PrivateKey, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	k, err := ParsePEMPrivateKey(data)
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

	default:
		// Primarily because we need to make sure users can recreate them!
		return fmt.Errorf("deletion of keystore items of type %v not (yet) supported", item.Type)
	}
}
