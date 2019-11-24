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
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"math/big"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/util/pkg/vfs"
)

// ClientsetCAStore is a CAStore implementation that stores keypairs in Keyset on a API server
type ClientsetCAStore struct {
	cluster   *kops.Cluster
	namespace string
	clientset kopsinternalversion.KopsInterface

	mutex           sync.Mutex
	cachedCaKeysets map[string]*keyset
}

var _ CAStore = &ClientsetCAStore{}
var _ SSHCredentialStore = &ClientsetCAStore{}

// NewClientsetCAStore is the constructor for ClientsetCAStore
func NewClientsetCAStore(cluster *kops.Cluster, clientset kopsinternalversion.KopsInterface, namespace string) CAStore {
	c := &ClientsetCAStore{
		cluster:         cluster,
		clientset:       clientset,
		namespace:       namespace,
		cachedCaKeysets: make(map[string]*keyset),
	}

	return c
}

// NewClientsetSSHCredentialStore creates an SSHCredentialStore backed by an API client
func NewClientsetSSHCredentialStore(cluster *kops.Cluster, clientset kopsinternalversion.KopsInterface, namespace string) SSHCredentialStore {
	// Note: currently identical to NewClientsetCAStore
	c := &ClientsetCAStore{
		cluster:         cluster,
		clientset:       clientset,
		namespace:       namespace,
		cachedCaKeysets: make(map[string]*keyset),
	}

	return c
}

// readCAKeypairs retrieves the CA keypair.
// (No longer generates a keypair if not found.)
func (c *ClientsetCAStore) readCAKeypairs(id string) (*keyset, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cached := c.cachedCaKeysets[id]
	if cached != nil {
		return cached, nil
	}

	keyset, err := c.loadKeyset(id)
	if err != nil {
		return nil, err
	}

	if keyset == nil {
		return nil, nil
	}
	c.cachedCaKeysets[id] = keyset
	return keyset, nil
}

// generateCACertificate creates and stores a CA keypair
// Should be called with the mutex held, to prevent concurrent creation of different keys
func (c *ClientsetCAStore) generateCACertificate(id string) (*keyset, error) {
	template := BuildCAX509Template()

	caRsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	caPrivateKey := &pki.PrivateKey{Key: caRsaKey}

	t := time.Now().UnixNano()
	template.SerialNumber = pki.BuildPKISerial(t)

	caCertificate, err := pki.SignNewCertificate(caPrivateKey, template, nil, nil)
	if err != nil {
		return nil, err
	}

	return c.storeAndVerifyKeypair(id, caCertificate, caPrivateKey)
}

// keyset is a parsed Keyset
type keyset struct {
	format  KeysetFormat
	items   map[string]*keysetItem
	primary *keysetItem
}

// keysetItem is a parsed KeysetItem
type keysetItem struct {
	id          string
	certificate *pki.Certificate
	privateKey  *pki.PrivateKey
}

func parseKeyset(o *kops.Keyset) (*keyset, error) {
	name := o.Name

	keyset := &keyset{
		items: make(map[string]*keysetItem),
	}

	for _, key := range o.Spec.Keys {
		ki := &keysetItem{
			id: key.Id,
		}
		if len(key.PublicMaterial) != 0 {
			cert, err := pki.ParsePEMCertificate(key.PublicMaterial)
			if err != nil {
				klog.Warningf("key public material was %s", key.PublicMaterial)
				return nil, fmt.Errorf("error loading certificate %s/%s: %v", name, key.Id, err)
			}
			ki.certificate = cert
		}

		if len(key.PrivateMaterial) != 0 {
			privateKey, err := pki.ParsePEMPrivateKey(key.PrivateMaterial)
			if err != nil {
				return nil, fmt.Errorf("error loading private key %s/%s: %v", name, key.Id, err)
			}
			ki.privateKey = privateKey
		}

		keyset.items[key.Id] = ki
	}

	keyset.primary = keyset.findPrimary()

	return keyset, nil
}

// loadKeyset gets the named keyset and the format of the Keyset.
func (c *ClientsetCAStore) loadKeyset(name string) (*keyset, error) {
	o, err := c.clientset.Keysets(c.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading keyset %q: %v", name, err)
	}

	keyset, err := parseKeyset(o)
	if err != nil {
		return nil, err
	}
	keyset.format = KeysetFormatV1Alpha2
	return keyset, nil
}

// findPrimary returns the primary keysetItem in the keyset
func (k *keyset) findPrimary() *keysetItem {
	var primary *keysetItem
	var primaryVersion *big.Int

	for _, item := range k.items {
		version, ok := big.NewInt(0).SetString(item.id, 10)
		if !ok {
			klog.Warningf("Ignoring key item with non-integer version: %q", item.id)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			primary = item
			primaryVersion = version
		}
	}
	return primary
}

// FindPrimary returns the primary KeysetItem in the Keyset
func FindPrimary(keyset *kops.Keyset) *kops.KeysetItem {
	var primary *kops.KeysetItem
	var primaryVersion *big.Int
	for i := range keyset.Spec.Keys {
		item := &keyset.Spec.Keys[i]
		version, ok := big.NewInt(0).SetString(item.Id, 10)
		if !ok {
			klog.Warningf("Ignoring key item with non-integer version: %q", item.Id)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			primary = item
			primaryVersion = version
		}
	}
	return primary
}

// CertificatePool implements CAStore::CertificatePool
func (c *ClientsetCAStore) CertificatePool(id string, createIfMissing bool) (*CertificatePool, error) {
	cert, err := c.FindCertificatePool(id)
	if err == nil && cert == nil {
		if !createIfMissing {
			klog.Warningf("using empty certificate, because running with DryRun")
			return &CertificatePool{}, err
		}
		return nil, fmt.Errorf("cannot find certificate pool %q", id)
	}
	return cert, err

}

// FindKeypair implements CAStore::FindKeypair
func (c *ClientsetCAStore) FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, KeysetFormat, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, nil, "", err
	}

	if keyset != nil && keyset.primary != nil {
		return keyset.primary.certificate, keyset.primary.privateKey, keyset.format, nil
	}

	return nil, nil, "", nil
}

// FindCert implements CAStore::FindCert
func (c *ClientsetCAStore) FindCert(name string) (*pki.Certificate, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, err
	}

	if keyset != nil && keyset.primary != nil {
		return keyset.primary.certificate, nil
	}

	return nil, nil
}

// FindCertificatePool implements CAStore::FindCertificatePool
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

// FindCertificateKeyset implements CAStore::FindCertificateKeyset
func (c *ClientsetCAStore) FindCertificateKeyset(name string) (*kops.Keyset, error) {
	o, err := c.clientset.Keysets(c.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading keyset %q: %v", name, err)
	}
	return o, nil
}

// ListKeysets implements CAStore::ListKeysets
func (c *ClientsetCAStore) ListKeysets() ([]*kops.Keyset, error) {
	var items []*kops.Keyset

	{
		list, err := c.clientset.Keysets(c.namespace).List(v1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error listing Keysets: %v", err)
		}

		for i := range list.Items {
			keyset := &list.Items[i]
			switch keyset.Spec.Type {
			case kops.SecretTypeKeypair:
				items = append(items, &list.Items[i])

			case kops.SecretTypeSecret:
				continue // Ignore - this is handled by ClientsetSecretStore
			default:
				return nil, fmt.Errorf("unhandled secret type %q: %v", keyset.Spec.Type, err)
			}
		}
	}

	return items, nil
}

// ListSSHCredentials implements SSHCredentialStore::ListSSHCredentials
func (c *ClientsetCAStore) ListSSHCredentials() ([]*kops.SSHCredential, error) {
	var items []*kops.SSHCredential

	{
		list, err := c.clientset.SSHCredentials(c.namespace).List(v1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error listing SSHCredentials: %v", err)
		}

		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	}

	return items, nil
}

// IssueCert implements CAStore::IssueCert
func (c *ClientsetCAStore) IssueCert(signer string, name string, serial *big.Int, privateKey *pki.PrivateKey, template *x509.Certificate) (*pki.Certificate, error) {
	klog.Infof("Issuing new certificate: %q", name)

	template.SerialNumber = serial

	caKeyset, err := c.readCAKeypairs(signer)
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
	cert, err := pki.SignNewCertificate(privateKey, template, caKeyset.primary.certificate.Certificate, caKeyset.primary.privateKey)
	if err != nil {
		return nil, err
	}

	if _, err := c.storeAndVerifyKeypair(name, cert, privateKey); err != nil {
		return nil, err
	}

	return cert, nil
}

// storeAndVerifyKeypair writes the keypair, also re-reading it to double-check it
func (c *ClientsetCAStore) storeAndVerifyKeypair(name string, cert *pki.Certificate, privateKey *pki.PrivateKey) (*keyset, error) {
	id := cert.Certificate.SerialNumber.String()
	if err := c.storeKeypair(name, id, cert, privateKey); err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, fmt.Errorf("error fetching stored certificate: %v", err)
	}

	if keyset == nil {
		return nil, fmt.Errorf("stored certificate not found: %v", err)
	}
	if keyset.primary == nil {
		return nil, fmt.Errorf("stored certificate did not have data: %v", err)
	}
	if keyset.primary.id != id {
		return nil, fmt.Errorf("stored certificate changed concurrently (id mismatch)")
	}
	return keyset, nil
}

// StoreKeypair implements CAStore::StoreKeypair
func (c *ClientsetCAStore) StoreKeypair(name string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	return c.storeKeypair(name, cert.Certificate.SerialNumber.String(), cert, privateKey)
}

// AddCert implements CAStore::AddCert
func (c *ClientsetCAStore) AddCert(name string, cert *pki.Certificate) error {
	klog.Infof("Adding TLS certificate: %q", name)

	// We add with a timestamp of zero so this will never be the newest cert
	serial := pki.BuildPKISerial(0)

	err := c.storeKeypair(name, serial.String(), cert, nil)
	if err != nil {
		return err
	}

	return nil
}

// FindPrivateKey implements CAStore::FindPrivateKey
func (c *ClientsetCAStore) FindPrivateKey(name string) (*pki.PrivateKey, error) {
	keyset, err := c.loadKeyset(name)
	if err != nil {
		return nil, err
	}

	if keyset != nil && keyset.primary != nil {
		return keyset.primary.privateKey, nil
	}
	return nil, nil
}

// FindPrivateKeyset implements CAStore::FindPrivateKeyset
func (c *ClientsetCAStore) FindPrivateKeyset(name string) (*kops.Keyset, error) {
	o, err := c.clientset.Keysets(c.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading keyset %q: %v", name, err)
	}
	return o, nil
}

// CreateKeypair implements CAStore::CreateKeypair
func (c *ClientsetCAStore) CreateKeypair(signer string, id string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	serial := c.buildSerial()

	cert, err := c.IssueCert(signer, id, serial, privateKey, template)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// addKey saves the specified key to the registry
func (c *ClientsetCAStore) addKey(name string, keysetType kops.KeysetType, item *kops.KeysetItem) error {
	create := false
	client := c.clientset.Keysets(c.namespace)
	keyset, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			keyset = nil
		} else {
			return fmt.Errorf("error reading keyset %q: %v", name, err)
		}
	}
	if keyset == nil {
		keyset = &kops.Keyset{}
		keyset.Name = name
		keyset.Spec.Type = keysetType
		create = true
	}
	keyset.Spec.Keys = append(keyset.Spec.Keys, *item)
	if create {
		if _, err := client.Create(keyset); err != nil {
			return fmt.Errorf("error creating keyset %q: %v", name, err)
		}
	} else {
		if _, err := client.Update(keyset); err != nil {
			return fmt.Errorf("error updating keyset %q: %v", name, err)
		}
	}
	return nil
}

// DeleteKeysetItem deletes the specified key from the registry; deleting the whole keyset if it was the last one
func DeleteKeysetItem(client kopsinternalversion.KeysetInterface, name string, keysetType kops.KeysetType, id string) error {
	keyset, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error reading Keyset %q: %v", name, err)
	}

	if keyset.Spec.Type != keysetType {
		return fmt.Errorf("mismatch on Keyset type on %q", name)
	}

	var newKeys []kops.KeysetItem
	found := false
	for _, ki := range keyset.Spec.Keys {
		if ki.Id == id {
			found = true
		} else {
			newKeys = append(newKeys, ki)
		}
	}
	if !found {
		return fmt.Errorf("KeysetItem %q not found in Keyset %q", id, name)
	}
	if len(newKeys) == 0 {
		if err := client.Delete(name, &v1.DeleteOptions{}); err != nil {
			return fmt.Errorf("error deleting Keyset %q: %v", name, err)
		}
	} else {
		keyset.Spec.Keys = newKeys
		if _, err := client.Update(keyset); err != nil {
			return fmt.Errorf("error updating Keyset %q: %v", name, err)
		}
	}
	return nil
}

// addSshCredential saves the specified SSH Credential to the registry, doing an update or insert
func (c *ClientsetCAStore) addSshCredential(name string, publicKey string) error {
	create := false
	client := c.clientset.SSHCredentials(c.namespace)
	sshCredential, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			sshCredential = nil
		} else {
			return fmt.Errorf("error reading SSHCredential %q: %v", name, err)
		}
	}
	if sshCredential == nil {
		sshCredential = &kops.SSHCredential{}
		sshCredential.Name = name
		create = true
	}
	sshCredential.Spec.PublicKey = publicKey
	if create {
		if _, err := client.Create(sshCredential); err != nil {
			return fmt.Errorf("error creating SSHCredential %q: %v", name, err)
		}
	} else {
		if _, err := client.Update(sshCredential); err != nil {
			return fmt.Errorf("error updating SSHCredential %q: %v", name, err)
		}
	}
	return nil
}

// deleteSSHCredential deletes the specified SSHCredential from the registry
func (c *ClientsetCAStore) deleteSSHCredential(name string) error {
	client := c.clientset.SSHCredentials(c.namespace)
	err := client.Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting SSHCredential %q: %v", name, err)
	}
	return nil
}

// addKey saves the specified keypair to the registry
func (c *ClientsetCAStore) storeKeypair(name string, id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	var publicMaterial bytes.Buffer
	if _, err := cert.WriteTo(&publicMaterial); err != nil {
		return err
	}

	var privateMaterial bytes.Buffer
	if _, err := privateKey.WriteTo(&privateMaterial); err != nil {
		return err
	}

	item := &kops.KeysetItem{
		Id:              id,
		PublicMaterial:  publicMaterial.Bytes(),
		PrivateMaterial: privateMaterial.Bytes(),
	}
	return c.addKey(name, kops.SecretTypeKeypair, item)
}

// buildSerial returns a serial for use when issuing certificates
func (c *ClientsetCAStore) buildSerial() *big.Int {
	t := time.Now().UnixNano()
	return pki.BuildPKISerial(t)
}

// AddSSHPublicKey implements CAStore::AddSSHPublicKey
func (c *ClientsetCAStore) AddSSHPublicKey(name string, pubkey []byte) error {
	_, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
	if err != nil {
		return fmt.Errorf("error parsing SSH public key: %v", err)
	}

	// TODO: Reintroduce or remove
	//// compute fingerprint to serve as id
	//h := md5.New()
	//_, err = h.Write(sshPublicKey.Marshal())
	//if err != nil {
	//	return err
	//}
	//id = formatFingerprint(h.Sum(nil))

	return c.addSshCredential(name, string(pubkey))
}

// FindSSHPublicKeys implements CAStore::FindSSHPublicKeys
func (c *ClientsetCAStore) FindSSHPublicKeys(name string) ([]*kops.SSHCredential, error) {
	o, err := c.clientset.SSHCredentials(c.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading SSHCredential %q: %v", name, err)
	}

	items := []*kops.SSHCredential{o}
	return items, nil
}

// DeleteKeysetItem implements CAStore::DeleteKeysetItem
func (c *ClientsetCAStore) DeleteKeysetItem(item *kops.Keyset, id string) error {
	switch item.Spec.Type {
	case kops.SecretTypeKeypair:
		client := c.clientset.Keysets(c.namespace)
		return DeleteKeysetItem(client, item.Name, kops.SecretTypeKeypair, id)
	default:
		// Primarily because we need to make sure users can recreate them!
		return fmt.Errorf("deletion of keystore items of type %v not (yet) supported", item.Spec.Type)
	}
}

// DeleteSSHCredential implements SSHCredentialStore::DeleteSSHCredential
func (c *ClientsetCAStore) DeleteSSHCredential(item *kops.SSHCredential) error {
	return c.deleteSSHCredential(item.Name)
}

func (c *ClientsetCAStore) MirrorTo(basedir vfs.Path) error {
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
