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
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/util/pkg/vfs"
	"math/big"
	"sync"
	"time"
)

type ClientsetCAStore struct {
	namespace string
	clientset kopsinternalversion.KopsInterface

	mutex         sync.Mutex
	cacheCaKeyset *keyset
}

var _ CAStore = &ClientsetCAStore{}

func NewClientsetCAStore(clientset kopsinternalversion.KopsInterface, namespace string) CAStore {
	c := &ClientsetCAStore{
		clientset: clientset,
		namespace: namespace,
	}

	return c
}

// Retrieves the CA keypair, generating a new keypair if not found
func (s *ClientsetCAStore) readCAKeypairs() (*keyset, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cacheCaKeyset != nil {
		return s.cacheCaKeyset, nil
	}

	keyset, err := s.loadKeyset(CertificateId_CA)
	if err != nil {
		return nil, err
	}

	if keyset == nil {
		keyset, err = s.generateCACertificate()
		if err != nil {
			return nil, err
		}

	}
	s.cacheCaKeyset = keyset

	return keyset, nil
}

// Creates and stores CA keypair
// Should be called with the mutex held, to prevent concurrent creation of different keys
func (c *ClientsetCAStore) generateCACertificate() (*keyset, error) {
	template := BuildCAX509Template()

	caRsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	caPrivateKey := &PrivateKey{Key: caRsaKey}

	caCertificate, err := SignNewCertificate(caPrivateKey, template, nil, nil)
	if err != nil {
		return nil, err
	}

	t := time.Now().UnixNano()
	serial := BuildPKISerial(t)
	id := serial.String()

	err = c.storeKeypair(CertificateId_CA, id, caCertificate, caPrivateKey)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	keyset, err := c.loadKeyset(CertificateId_CA)
	if err != nil {
		return nil, err
	}
	if keyset == nil || keyset.primary == nil || keyset.primary.id != id {
		return nil, fmt.Errorf("failed to round-trip CA keyset")
	}

	return keyset, nil
}

//func (c *VFSCAStore) loadCertificates(p vfs.Path) (*certificates, error) {
//	files, err := p.ReadDir()
//	if err != nil {
//		if os.IsNotExist(err) {
//			return nil, nil
//		}
//		return nil, err
//	}
//
//	certs := &certificates{
//		certificates: make(map[string]*Certificate),
//	}
//
//	for _, f := range files {
//		cert, err := c.loadOneCertificate(f)
//		if err != nil {
//			return nil, fmt.Errorf("error loading certificate %q: %v", f, err)
//		}
//		name := f.Base()
//		name = strings.TrimSuffix(name, ".crt")
//		certs.certificates[name] = cert
//	}
//
//	if len(certs.certificates) == 0 {
//		return nil, nil
//	}
//
//	var primaryVersion *big.Int
//	for k := range certs.certificates {
//		version, ok := big.NewInt(0).SetString(k, 10)
//		if !ok {
//			glog.Warningf("Ignoring certificate with non-integer version: %q", k)
//			continue
//		}
//
//		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
//			certs.primary = k
//			primaryVersion = version
//		}
//	}
//
//	return certs, nil
//}

type keyset struct {
	items   map[string]*keysetItem
	primary *keysetItem
}

type keysetItem struct {
	id          string
	certificate *Certificate
	privateKey  *PrivateKey
}

func (c *ClientsetCAStore) loadKeyset(name string) (*keyset, error) {
	o, err := c.clientset.Keysets(c.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading keyset %q: %v", name, err)
	}

	keyset := &keyset{
		items: make(map[string]*keysetItem),
	}

	for _, key := range o.Spec.Keys {
		cert, err := LoadPEMCertificate(key.PublicMaterial)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate %s/%s: %v", name, key.Id, err)
		}
		privateKey, err := ParsePEMPrivateKey(key.PrivateMaterial)
		if err != nil {
			return nil, fmt.Errorf("error loading private key %s/%s: %v", name, key.Id, err)
		}
		keyset.items[key.Id] = &keysetItem{
			id:          key.Id,
			certificate: cert,
			privateKey:  privateKey,
		}
	}

	//if len(certs.certificates) == 0 {
	//	return nil, nil
	//}

	primary := FindPrimary(o)
	if primary != nil {
		keyset.primary = keyset.items[primary.Id]
	}

	return keyset, nil
}

func FindPrimary(keyset *kops.Keyset) *kops.KeyItem {
	var primary *kops.KeyItem
	var primaryVersion *big.Int
	for i := range keyset.Spec.Keys {
		item := &keyset.Spec.Keys[i]
		version, ok := big.NewInt(0).SetString(item.Id, 10)
		if !ok {
			glog.Warningf("Ignoring key item with non-integer version: %q", item.Id)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			primary = item
			primaryVersion = version
		}
	}
	return primary
}
func (c *ClientsetCAStore) Cert(name string, createIfMissing bool) (*Certificate, error) {
	cert, err := c.FindCert(name)
	if err == nil && cert == nil {
		if !createIfMissing {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &Certificate{}, err
		}
		return nil, fmt.Errorf("cannot find certificate %q", name)
	}
	return cert, err

}

func (c *ClientsetCAStore) CertificatePool(id string, createIfMissing bool) (*CertificatePool, error) {
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

func (c *ClientsetCAStore) FindKeypair(name string) (*Certificate, *PrivateKey, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, nil, err
	}

	if keyset != nil && keyset.primary != nil {
		return keyset.primary.certificate, keyset.primary.privateKey, nil
	}

	return nil, nil, nil
}

func (c *ClientsetCAStore) FindCert(name string) (*Certificate, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, err
	}

	var cert *Certificate
	if keyset != nil && keyset.primary != nil {
		cert = keyset.primary.certificate
	}

	return cert, nil
}

func (c *ClientsetCAStore) FindCertificatePool(name string) (*CertificatePool, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, err
	}

	pool := &CertificatePool{}

	if keyset != nil {
		if keyset.primary != nil {
			pool.Primary = keyset.primary.certificate
		}

		for id, item := range keyset.items {
			if id == keyset.primary.id {
				continue
			}
			pool.Secondary = append(pool.Secondary, item.certificate)
		}
	}
	return pool, nil
}

func (c *ClientsetCAStore) List() ([]*KeystoreItem, error) {
	list, err := c.clientset.Keysets(c.namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing keysets: %v", err)
	}

	var items []*KeystoreItem

	for _, keyset := range list.Items {
		for _, item := range keyset.Spec.Keys {
			ki := &KeystoreItem{
				Name: keyset.Name,
				Id:   item.Id,
			}

			switch keyset.Spec.Type {
			case kops.SecretTypeSSHPublicKey:
				ki.Type = SecretTypeSSHPublicKey
			case kops.SecretTypeKeypair:
				ki.Type = SecretTypeKeypair
			case kops.SecretTypeSecret:
				//ki.Type = SecretTypeSecret
				continue // Ignore
			default:
				return nil, fmt.Errorf("unhandled secret type %q: %v", ki.Type, err)
			}
			items = append(items, ki)
		}
	}

	return items, nil
}

func (c *ClientsetCAStore) IssueCert(name string, serial *big.Int, privateKey *PrivateKey, template *x509.Certificate) (*Certificate, error) {
	glog.Infof("Issuing new certificate: %q", name)

	template.SerialNumber = serial

	caKeyset, err := c.readCAKeypairs()
	if err != nil {
		return nil, err
	}

	if caKeyset == nil {
		return nil, fmt.Errorf("ca keyset was not found; cannot issue certificates")
	}
	if caKeyset.primary == nil {
		return nil, fmt.Errorf("ca keyset did not have any key data; cannot issue certificates")
	}
	if caKeyset.primary.certificate == nil {
		return nil, fmt.Errorf("ca certificate was not found; cannot issue certificates")
	}
	if caKeyset.primary.privateKey == nil {
		return nil, fmt.Errorf("ca privateKey was not found; cannot issue certificates")
	}
	cert, err := SignNewCertificate(privateKey, template, caKeyset.primary.certificate.Certificate, caKeyset.primary.privateKey)
	if err != nil {
		return nil, err
	}

	err = c.StoreKeypair(name, cert, privateKey)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, fmt.Errorf("error fetching issued certificate: %v", err)
	}

	if keyset == nil {
		return nil, fmt.Errorf("issued certificate not found: %v", err)
	}
	if keyset.primary == nil {
		return nil, fmt.Errorf("issued certificate did not have data: %v", err)
	}
	if keyset.primary.id != serial.String() {
		return nil, fmt.Errorf("issued certificate changed concurrently (id mismatch)")
	}
	return keyset.primary.certificate, nil
}

func (c *ClientsetCAStore) StoreKeypair(name string, cert *Certificate, privateKey *PrivateKey) error {
	serial := cert.Certificate.SerialNumber

	return c.storeKeypair(name, serial.String(), cert, privateKey)
}

func (c *ClientsetCAStore) AddCert(name string, cert *Certificate) error {
	glog.Infof("Adding TLS certificate: %q", name)

	// We add with a timestamp of zero so this will never be the newest cert
	serial := BuildPKISerial(0)

	err := c.storeKeypair(name, serial.String(), cert, nil)
	if err != nil {
		return err
	}

	//// Make double-sure it round-trips
	//_, err = c.loadKeyset(name)
	//return err

	return nil
}

func (c *ClientsetCAStore) FindPrivateKey(name string) (*PrivateKey, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, err
	}

	if keyset != nil && keyset.primary != nil {
		return keyset.primary.privateKey, nil
	}
	return nil, nil
}

func (c *ClientsetCAStore) PrivateKey(name string, createIfMissing bool) (*PrivateKey, error) {
	key, err := c.FindPrivateKey(name)
	if err == nil && key == nil {
		if !createIfMissing {
			glog.Warningf("using empty certificate, because running with DryRun")
			return &PrivateKey{}, err
		}
		return nil, fmt.Errorf("cannot find SSL key %q", name)
	}
	return key, err

}

func (c *ClientsetCAStore) CreateKeypair(id string, template *x509.Certificate, privateKey *PrivateKey) (*Certificate, error) {
	serial := c.buildSerial()

	cert, err := c.IssueCert(id, serial, privateKey, template)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *ClientsetCAStore) storeKeypair(name string, id string, cert *Certificate, privateKey *PrivateKey) error {
	var publicMaterial bytes.Buffer
	if _, err := cert.WriteTo(&publicMaterial); err != nil {
		return err
	}

	var privateMaterial bytes.Buffer
	if _, err := privateKey.WriteTo(&privateMaterial); err != nil {
		return err
	}

	client := c.clientset.Keysets(c.namespace)
	o, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading keyset %q: %v", name, err)
	}
	key := kops.KeyItem{
		Id:              id,
		PublicMaterial:  publicMaterial.Bytes(),
		PrivateMaterial: privateMaterial.Bytes(),
	}
	o.Spec.Keys = append(o.Spec.Keys, key)
	if _, err := client.Update(o); err != nil {
		return fmt.Errorf("error updating keyset %q: %v", name, err)
	}

	return nil
}

func (c *ClientsetCAStore) buildSerial() *big.Int {
	t := time.Now().UnixNano()
	return BuildPKISerial(t)
}

// AddSSHPublicKey stores an SSH public key
func (c *ClientsetCAStore) AddSSHPublicKey(name string, pubkey []byte) error {
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

	name = "ssh-" + name

	client := c.clientset.Keysets(c.namespace)
	o, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading keyset %q: %v", name, err)
	}
	key := kops.KeyItem{
		Id:             id,
		PublicMaterial: pubkey,
	}
	o.Spec.Keys = append(o.Spec.Keys, key)
	if _, err := client.Update(o); err != nil {
		return fmt.Errorf("error updating keyset %q: %v", name, err)
	}

	return nil
}

func (c *ClientsetCAStore) FindSSHPublicKeys(name string) ([]*KeystoreItem, error) {
	itemName := "ssh-" + name

	client := c.clientset.Keysets(c.namespace)
	o, err := client.Get(itemName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error reading keyset %q: %v", itemName, err)
	}

	if o.Spec.Type != kops.SecretTypeSSHPublicKey {
		return nil, fmt.Errorf("expecting type %s for %q, was type %s", kops.SecretTypeSSHPublicKey, itemName, o.Spec.Type)
	}

	var items []*KeystoreItem
	for _, k := range o.Spec.Keys {
		item := &KeystoreItem{
			Type: SecretTypeSSHPublicKey,
			Name: name,
			Id:   insertFingerprintColons(k.Id),
			Data: k.PublicMaterial,
		}
		items = append(items, item)
	}

	return items, nil
}

func (c *ClientsetCAStore) DeleteSecret(item *KeystoreItem) error {
	switch item.Type {
	//case SecretTypeSSHPublicKey:
	//	p := c.buildSSHPublicKeyPath(item.Name, item.Id)
	//	return p.Remove()

	default:
		// Primarily because we need to make sure users can recreate them!
		return fmt.Errorf("deletion of keystore items of type %v not (yet) supported", item.Type)
	}
}

func (c *ClientsetCAStore) MirrorTo(basedir vfs.Path) error {
	list, err := c.clientset.Keysets(c.namespace).List(v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing keysets: %v", err)
	}

	for i := range list.Items {
		keyset := &list.Items[i]

		if keyset.Spec.Type == kops.SecretTypeSecret {
			continue
		}

		primary := FindPrimary(keyset)
		if primary == nil {
			glog.Warningf("skipping keyset with no primary data: %s", keyset.Name)
			continue
		}

		switch keyset.Spec.Type {
		case kops.SecretTypeSSHPublicKey:
			for i := range keyset.Spec.Keys {
				item := &keyset.Spec.Keys[i]
				p := basedir.Join("ssh", "public", keyset.Name, item.Id)
				err = p.WriteFile(item.PublicMaterial)
				if err != nil {
					return fmt.Errorf("error writing %q: %v", p, err)
				}
			}

		case kops.SecretTypeKeypair:
			for i := range keyset.Spec.Keys {
				item := &keyset.Spec.Keys[i]
				{
					p := basedir.Join("issued", keyset.Name, item.Id+".crt")
					err = p.WriteFile(item.PublicMaterial)
					if err != nil {
						return fmt.Errorf("error writing %q: %v", p, err)
					}
				}
				{
					p := basedir.Join("private", keyset.Name, item.Id+".key")
					err = p.WriteFile(item.PrivateMaterial)
					if err != nil {
						return fmt.Errorf("error writing %q: %v", p, err)
					}
				}
			}

		}
	}

	return nil
}
