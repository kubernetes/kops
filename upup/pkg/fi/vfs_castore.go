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
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/sshcredentials"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSCAStore struct {
	basedir   vfs.Path
	cluster   *kops.Cluster
	allowList bool

	mutex     sync.Mutex
	cachedCAs map[string]*cachedEntry

	// SerialGenerator is the function for generating certificate serial numbers
	// It can be replaced for testing purposes.
	SerialGenerator func() *big.Int
}

type cachedEntry struct {
	certificates *keyset
	privateKeys  *keyset
}

var _ CAStore = &VFSCAStore{}
var _ SSHCredentialStore = &VFSCAStore{}

func NewVFSCAStore(cluster *kops.Cluster, basedir vfs.Path, allowList bool) *VFSCAStore {
	c := &VFSCAStore{
		basedir:   basedir,
		cluster:   cluster,
		cachedCAs: make(map[string]*cachedEntry),
		allowList: allowList,
	}

	c.SerialGenerator = func() *big.Int {
		t := time.Now().UnixNano()
		return pki.BuildPKISerial(t)
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

	caCertificates, err := s.loadCertificates(s.buildCertificatePoolPath(id), true)
	if err != nil {
		return nil, nil, err
	}

	var caPrivateKeys *keyset

	if caCertificates != nil {
		caPrivateKeys, err = s.loadPrivateKeys(s.buildPrivateKeyPoolPath(id), true)
		if err != nil {
			return nil, nil, err
		}

		if caPrivateKeys == nil {
			klog.Warningf("CA private key was not found")
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
		IsCA:                  true,
	}
	return template
}

func (c *VFSCAStore) buildCertificatePoolPath(name string) vfs.Path {
	return c.basedir.Join("issued", name)
}

func (c *VFSCAStore) buildCertificatePath(name string, id string) vfs.Path {
	return c.basedir.Join("issued", name, id+".crt")
}

func (c *VFSCAStore) buildPrivateKeyPoolPath(name string) vfs.Path {
	return c.basedir.Join("private", name)
}

func (c *VFSCAStore) buildPrivateKeyPath(name string, id string) vfs.Path {
	return c.basedir.Join("private", name, id+".key")
}

func (c *VFSCAStore) parseKeysetYaml(data []byte) (*kops.Keyset, KeysetFormat, error) {
	defaultReadVersion := v1alpha2.SchemeGroupVersion.WithKind("Keyset")

	object, gvk, err := kopscodecs.Decode(data, &defaultReadVersion)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing keyset: %v", err)
	}

	keyset, ok := object.(*kops.Keyset)
	if !ok {
		return nil, "", fmt.Errorf("object was not a keyset, was a %T", object)
	}

	if gvk == nil {
		return nil, "", fmt.Errorf("object did not have GroupVersionKind: %q", keyset.Name)
	}

	return keyset, KeysetFormat(gvk.Version), nil
}

// loadCertificatesBundle loads a keyset from the path
// Returns (nil, nil) if the file is not found
// Bundles avoid the need for a list-files permission, which can be tricky on e.g. GCE
func (c *VFSCAStore) loadKeysetBundle(p vfs.Path) (*keyset, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read bundle %q: %v", p, err)
	}

	o, format, err := c.parseKeysetYaml(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing bundle %q: %v", p, err)
	}

	keyset, err := parseKeyset(o)
	if err != nil {
		return nil, fmt.Errorf("error mapping bundle %q: %v", p, err)
	}

	keyset.format = format
	return keyset, nil
}

func (k *keyset) ToAPIObject(name string, includePrivateKeyMaterial bool) (*kops.Keyset, error) {
	o := &kops.Keyset{}
	o.Name = name
	o.Spec.Type = kops.SecretTypeKeypair

	for _, ki := range k.items {
		oki := kops.KeysetItem{
			Id: ki.id,
		}

		if ki.certificate != nil {
			var publicMaterial bytes.Buffer
			if _, err := ki.certificate.WriteTo(&publicMaterial); err != nil {
				return nil, err
			}
			oki.PublicMaterial = publicMaterial.Bytes()
		}

		if includePrivateKeyMaterial && ki.privateKey != nil {
			var privateMaterial bytes.Buffer
			if _, err := ki.privateKey.WriteTo(&privateMaterial); err != nil {
				return nil, err
			}

			oki.PrivateMaterial = privateMaterial.Bytes()
		}

		o.Spec.Keys = append(o.Spec.Keys, oki)
	}
	return o, nil
}

// writeKeysetBundle writes a keyset bundle to VFS
func (c *VFSCAStore) writeKeysetBundle(p vfs.Path, name string, keyset *keyset, includePrivateKeyMaterial bool) error {
	p = p.Join("keyset.yaml")

	o, err := keyset.ToAPIObject(name, includePrivateKeyMaterial)
	if err != nil {
		return err
	}

	objectData, err := serializeKeysetBundle(o)
	if err != nil {
		return err
	}

	acl, err := acls.GetACL(p, c.cluster)
	if err != nil {
		return err
	}
	return p.WriteFile(bytes.NewReader(objectData), acl)
}

// serializeKeysetBundle converts a keyset bundle to yaml, for writing to VFS
func serializeKeysetBundle(o *kops.Keyset) ([]byte, error) {
	var objectData bytes.Buffer
	codecs := kopscodecs.Codecs
	yaml, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		klog.Fatalf("no YAML serializer registered")
	}
	encoder := codecs.EncoderForVersion(yaml.Serializer, v1alpha2.SchemeGroupVersion)

	if err := encoder.Encode(o, &objectData); err != nil {
		return nil, fmt.Errorf("error serializing keyset: %v", err)
	}
	return objectData.Bytes(), nil
}

// removePrivateKeyMaterial returns a copy of the Keyset with the private key data removed
func removePrivateKeyMaterial(o *kops.Keyset) *kops.Keyset {
	copy := o.DeepCopy()

	for i := range copy.Spec.Keys {
		copy.Spec.Keys[i].PrivateMaterial = nil
	}

	return copy
}

func SerializeKeyset(o *kops.Keyset) ([]byte, error) {
	var objectData bytes.Buffer
	{
		codecs := kopscodecs.Codecs
		yaml, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), "application/yaml")
		if !ok {
			klog.Fatalf("no YAML serializer registered")
		}
		encoder := codecs.EncoderForVersion(yaml.Serializer, v1alpha2.SchemeGroupVersion)

		if err := encoder.Encode(o, &objectData); err != nil {
			return nil, fmt.Errorf("error serializing keyset: %v", err)
		}
	}

	return objectData.Bytes(), nil
}

func (c *VFSCAStore) loadCertificates(p vfs.Path, useBundle bool) (*keyset, error) {
	// Attempt to load prebuilt bundle, which avoids having to list files, which is a permission that can be hard to
	// give on GCE / other clouds
	if useBundle {
		bundlePath := p.Join("keyset.yaml")
		bundle, err := c.loadKeysetBundle(bundlePath)
		if !c.allowList {
			return bundle, err
		}

		if err != nil {
			klog.Warningf("unable to read bundle %q, falling back to directory-list method: %v", bundlePath, err)
		} else if bundle == nil {
			klog.V(2).Infof("no certificate bundle %q, falling back to directory-list method", bundlePath)
		} else {
			return bundle, nil
		}
	}

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
		id := f.Base()
		if strings.HasSuffix(id, ".yaml") {
			// ignore bundle
			continue
		}
		id = strings.TrimSuffix(id, ".crt")

		cert, err := c.loadOneCertificate(f)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate %q: %v", f, err)
		}

		keyset.items[id] = &keysetItem{
			id:          id,
			certificate: cert,
		}
	}

	if len(keyset.items) == 0 {
		return nil, nil
	}

	keyset.format = KeysetFormatLegacy
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
	cert, err := pki.ParsePEMCertificate(data)
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
			klog.Warningf("using empty certificate pool for %q, because createIfMissing=false", id)
			return &CertificatePool{}, err
		}
		return nil, fmt.Errorf("cannot find certificate pool %q", id)
	}
	return cert, err

}

func (c *VFSCAStore) FindKeypair(id string) (*pki.Certificate, *pki.PrivateKey, KeysetFormat, error) {
	cert, certFormat, err := c.findCert(id)
	if err != nil {
		return nil, nil, "", err
	}

	key, err := c.FindPrivateKey(id)
	if err != nil {
		return nil, nil, "", err
	}

	return cert, key, certFormat, nil
}

func (c *VFSCAStore) findCert(name string) (*pki.Certificate, KeysetFormat, error) {
	p := c.buildCertificatePoolPath(name)
	certs, err := c.loadCertificates(p, true)
	if err != nil {
		return nil, "", fmt.Errorf("error in 'FindCert' attempting to load cert %q: %v", name, err)
	}

	if certs != nil && certs.primary != nil {
		return certs.primary.certificate, certs.format, nil
	}

	return nil, "", nil
}

func (c *VFSCAStore) FindCert(name string) (*pki.Certificate, error) {
	cert, _, err := c.findCert(name)
	return cert, err
}

func (c *VFSCAStore) FindCertificatePool(name string) (*CertificatePool, error) {
	var certs *keyset

	var err error
	p := c.buildCertificatePoolPath(name)
	certs, err = c.loadCertificates(p, true)
	if err != nil {
		return nil, fmt.Errorf("error in 'FindCertificatePool' attempting to load cert %q: %v", name, err)
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

func (c *VFSCAStore) FindCertificateKeyset(name string) (*kops.Keyset, error) {
	p := c.buildCertificatePoolPath(name)
	certs, err := c.loadCertificates(p, true)
	if err != nil {
		return nil, fmt.Errorf("error in 'FindCertificatePool' attempting to load cert %q: %v", name, err)
	}

	if certs == nil {
		return nil, nil
	}

	o, err := certs.ToAPIObject(name, false)
	if err != nil {
		return nil, err
	}

	return o, nil
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
				klog.V(2).Infof("ignoring unexpected file in keystore: %q", f)
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

			if tokens[1] == "keyset.yaml" {
				// TODO: Should we load the keyset to get the actual ids?
			} else {
				keyset.Spec.Keys = append(keyset.Spec.Keys, kops.KeysetItem{
					Id: strings.TrimSuffix(tokens[1], ".crt"),
				})
			}
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
				klog.V(2).Infof("ignoring unexpected file in keystore: %q", f)
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
		klog.V(2).Infof("Skipping key store mirror from %q to %q (same paths)", c.basedir, basedir)
		return nil
	}
	klog.V(2).Infof("Mirroring key store from %q to %q", c.basedir, basedir)

	keysets, err := c.ListKeysets()
	if err != nil {
		return err
	}

	for _, keyset := range keysets {
		if err := mirrorKeyset(c.cluster, basedir, keyset); err != nil {
			return err
		}
	}

	sshCredentials, err := c.ListSSHCredentials()
	if err != nil {
		return fmt.Errorf("error listing SSHCredentials: %v", err)
	}

	for _, sshCredential := range sshCredentials {
		if err := mirrorSSHCredential(c.cluster, basedir, sshCredential); err != nil {
			return err
		}
	}

	return nil
}

// mirrorKeyset writes keyset bundles for the certificates & privatekeys
func mirrorKeyset(cluster *kops.Cluster, basedir vfs.Path, keyset *kops.Keyset) error {
	primary := FindPrimary(keyset)
	if primary == nil {
		return fmt.Errorf("found keyset with no primary data: %s", keyset.Name)
	}

	switch keyset.Spec.Type {
	case kops.SecretTypeKeypair:
		{
			data, err := serializeKeysetBundle(removePrivateKeyMaterial(keyset))
			if err != nil {
				return err
			}
			p := basedir.Join("issued", keyset.Name, "keyset.yaml")
			acl, err := acls.GetACL(p, cluster)
			if err != nil {
				return err
			}

			err = p.WriteFile(bytes.NewReader(data), acl)
			if err != nil {
				return fmt.Errorf("error writing %q: %v", p, err)
			}
		}

		{
			data, err := serializeKeysetBundle(keyset)
			if err != nil {
				return err
			}
			p := basedir.Join("private", keyset.Name, "keyset.yaml")
			acl, err := acls.GetACL(p, cluster)
			if err != nil {
				return err
			}

			err = p.WriteFile(bytes.NewReader(data), acl)
			if err != nil {
				return fmt.Errorf("error writing %q: %v", p, err)
			}
		}

	default:
		return fmt.Errorf("unknown secret type: %q", keyset.Spec.Type)
	}

	return nil
}

// mirrorSSHCredential writes the SSH credential file to the mirror location
func mirrorSSHCredential(cluster *kops.Cluster, basedir vfs.Path, sshCredential *kops.SSHCredential) error {
	id, err := sshcredentials.Fingerprint(sshCredential.Spec.PublicKey)
	if err != nil {
		return fmt.Errorf("error fingerprinting SSH public key %q: %v", sshCredential.Name, err)
	}

	p := basedir.Join("ssh", "public", sshCredential.Name, id)
	acl, err := acls.GetACL(p, cluster)
	if err != nil {
		return err
	}

	err = p.WriteFile(bytes.NewReader([]byte(sshCredential.Spec.PublicKey)), acl)
	if err != nil {
		return fmt.Errorf("error writing %q: %v", p, err)
	}

	return nil
}

func (c *VFSCAStore) IssueCert(signer string, id string, serial *big.Int, privateKey *pki.PrivateKey, template *x509.Certificate) (*pki.Certificate, error) {
	klog.Infof("Issuing new certificate: %q", id)

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
	p := c.buildCertificatePath(id, serial.String())
	return c.loadOneCertificate(p)
}

func (c *VFSCAStore) StoreKeypair(name string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	serial := cert.Certificate.SerialNumber.String()

	ki := &keysetItem{
		id:          serial,
		certificate: cert,
		privateKey:  privateKey,
	}

	{
		err := c.storePrivateKey(name, ki)
		if err != nil {
			return err
		}
	}

	{
		err := c.storeCertificate(name, ki)
		if err != nil {
			// TODO: Delete private key?
			return err
		}
	}

	return nil
}

func (c *VFSCAStore) AddCert(name string, cert *pki.Certificate) error {
	klog.Infof("Adding TLS certificate: %q", name)

	// We add with a timestamp of zero so this will never be the newest cert
	serial := pki.BuildPKISerial(0).String()

	p := c.buildCertificatePath(name, serial)

	ki := &keysetItem{
		id:          serial,
		certificate: cert,
	}
	err := c.storeCertificate(name, ki)
	if err != nil {
		return err
	}

	// Make double-sure it round-trips
	_, err = c.loadOneCertificate(p)
	return err
}

func (c *VFSCAStore) loadPrivateKeys(p vfs.Path, useBundle bool) (*keyset, error) {
	// Attempt to load prebuilt bundle, which avoids having to list files, which is a permission that can be hard to
	// give on GCE / other clouds
	if useBundle {
		bundlePath := p.Join("keyset.yaml")
		bundle, err := c.loadKeysetBundle(bundlePath)

		if !c.allowList {
			return bundle, err
		}

		if err != nil {
			klog.Warningf("unable to read bundle %q, falling back to directory-list method: %v", bundlePath, err)
		} else if bundle == nil {
			klog.V(2).Infof("no private key bundle %q, falling back to directory-list method", bundlePath)
		} else {
			return bundle, nil
		}
	}

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
		id := f.Base()
		if strings.HasSuffix(id, ".yaml") {
			// ignore bundle
			continue
		}
		id = strings.TrimSuffix(id, ".key")

		privateKey, err := c.loadOnePrivateKey(f)
		if err != nil {
			return nil, fmt.Errorf("error loading private key %q: %v", f, err)
		}
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

func (c *VFSCAStore) findPrivateKeyset(id string) (*keyset, error) {
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
		keys, err = c.loadPrivateKeys(p, true)
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}

func (c *VFSCAStore) FindPrivateKey(id string) (*pki.PrivateKey, error) {
	keys, err := c.findPrivateKeyset(id)
	if err != nil {
		return nil, err
	}

	var key *pki.PrivateKey
	if keys != nil && keys.primary != nil {
		key = keys.primary.privateKey
	}
	return key, nil
}

func (c *VFSCAStore) FindPrivateKeyset(name string) (*kops.Keyset, error) {
	keys, err := c.findPrivateKeyset(name)
	if err != nil {
		return nil, err
	}

	o, err := keys.ToAPIObject(name, true)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (c *VFSCAStore) CreateKeypair(signer string, id string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	serial := c.SerialGenerator()

	cert, err := c.IssueCert(signer, id, serial, privateKey, template)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *VFSCAStore) storePrivateKey(name string, ki *keysetItem) error {
	if ki.privateKey == nil {
		return fmt.Errorf("privateKey not provided to storeCertificate")
	}

	// Write the bundle
	{
		p := c.buildPrivateKeyPoolPath(name)
		ks, err := c.loadPrivateKeys(p, false)
		if err != nil {
			return err
		}

		if ks == nil {
			ks = &keyset{}
		}
		if ks.items == nil {
			ks.items = make(map[string]*keysetItem)
		}
		ks.items[ki.id] = ki

		if err := c.writeKeysetBundle(p, name, ks, true); err != nil {
			return fmt.Errorf("error writing bundle: %v", err)
		}
	}

	// Write the data
	{
		var data bytes.Buffer
		if _, err := ki.privateKey.WriteTo(&data); err != nil {
			return err
		}

		p := c.buildPrivateKeyPath(name, ki.id)
		acl, err := acls.GetACL(p, c.cluster)
		if err != nil {
			return err
		}
		return p.WriteFile(bytes.NewReader(data.Bytes()), acl)
	}
}

func (c *VFSCAStore) storeCertificate(name string, ki *keysetItem) error {
	if ki.certificate == nil {
		return fmt.Errorf("certificate not provided to storeCertificate")
	}

	// Write the bundle
	{
		p := c.buildCertificatePoolPath(name)
		ks, err := c.loadCertificates(p, false)
		if err != nil {
			return err
		}

		if ks == nil {
			ks = &keyset{}
		}
		if ks.items == nil {
			ks.items = make(map[string]*keysetItem)
		}
		ks.items[ki.id] = ki

		if err := c.writeKeysetBundle(p, name, ks, false); err != nil {
			return fmt.Errorf("error writing bundle: %v", err)
		}
	}

	// Write the data
	{
		var data bytes.Buffer
		if _, err := ki.certificate.WriteTo(&data); err != nil {
			return err
		}

		p := c.buildCertificatePath(name, ki.id)
		acl, err := acls.GetACL(p, c.cluster)
		if err != nil {
			return err
		}
		return p.WriteFile(bytes.NewReader(data.Bytes()), acl)
	}
}

func (c *VFSCAStore) deletePrivateKey(name string, id string) (bool, error) {
	// Update the bundle
	{
		p := c.buildPrivateKeyPoolPath(name)
		ks, err := c.loadPrivateKeys(p, false)
		if err != nil {
			return false, err
		}

		if ks == nil || ks.items[id] == nil {
			return false, nil
		}
		delete(ks.items, id)

		if err := c.writeKeysetBundle(p, name, ks, true); err != nil {
			return false, fmt.Errorf("error writing bundle: %v", err)
		}
	}

	// Delete the file itself
	{

		p := c.buildPrivateKeyPath(name, id)
		if err := p.Remove(); err != nil {
			return false, err
		}
		return true, nil
	}
}

func (c *VFSCAStore) deleteCertificate(name string, id string) (bool, error) {
	// Update the bundle
	{
		p := c.buildPrivateKeyPoolPath(name)
		ks, err := c.loadCertificates(p, false)
		if err != nil {
			return false, err
		}

		if ks == nil || ks.items[id] == nil {
			return false, nil
		}
		delete(ks.items, id)

		if err := c.writeKeysetBundle(p, name, ks, false); err != nil {
			return false, fmt.Errorf("error writing bundle: %v", err)
		}
	}

	// Delete the file itself
	{
		p := c.buildCertificatePath(name, id)
		if err := p.Remove(); err != nil {
			return false, err
		}
		return true, nil
	}
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

	return p.WriteFile(bytes.NewReader(pubkey), acl)
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
				klog.V(2).Infof("Ignoring not-found issue reading %q", f)
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

// DeleteKeysetItem implements CAStore::DeleteKeysetItem
func (c *VFSCAStore) DeleteKeysetItem(item *kops.Keyset, id string) error {
	switch item.Spec.Type {
	case kops.SecretTypeKeypair:
		_, ok := big.NewInt(0).SetString(id, 10)
		if !ok {
			return fmt.Errorf("keypair had non-integer version: %q", id)
		}
		removed, err := c.deleteCertificate(item.Name, id)
		if err != nil {
			return fmt.Errorf("error deleting certificate: %v", err)
		}
		if !removed {
			klog.Warningf("certificate %s:%s was not found", item.Name, id)
		}
		removed, err = c.deletePrivateKey(item.Name, id)
		if err != nil {
			return fmt.Errorf("error deleting private key: %v", err)
		}
		if !removed {
			klog.Warningf("private key %s:%s was not found", item.Name, id)
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
