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
	"context"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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
}

var (
	_ CAStore            = &ClientsetCAStore{}
	_ SSHCredentialStore = &ClientsetCAStore{}
)

// NewClientsetCAStore is the constructor for ClientsetCAStore
func NewClientsetCAStore(cluster *kops.Cluster, clientset kopsinternalversion.KopsInterface, namespace string) *ClientsetCAStore {
	c := &ClientsetCAStore{
		cluster:   cluster,
		clientset: clientset,
		namespace: namespace,
	}

	return c
}

// NewClientsetSSHCredentialStore creates an SSHCredentialStore backed by an API client
func NewClientsetSSHCredentialStore(cluster *kops.Cluster, clientset kopsinternalversion.KopsInterface, namespace string) SSHCredentialStore {
	// Note: currently identical to NewClientsetCAStore
	c := &ClientsetCAStore{
		cluster:   cluster,
		clientset: clientset,
		namespace: namespace,
	}

	return c
}

func parseKeyset(o *kops.Keyset) (*Keyset, error) {
	name := o.Name

	keyset := &Keyset{
		Items: make(map[string]*KeysetItem),
	}

	for _, key := range o.Spec.Keys {
		ki := &KeysetItem{
			Id: key.Id,
		}
		if key.DistrustTimestamp != nil {
			distrustTimestamp := key.DistrustTimestamp.Time
			ki.DistrustTimestamp = &distrustTimestamp
		}
		if len(key.PublicMaterial) != 0 {
			cert, err := pki.ParsePEMCertificate(key.PublicMaterial)
			if err != nil {
				klog.Warningf("key public material was %s", key.PublicMaterial)
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

	keyset.Primary = keyset.Items[FindPrimary(o).Id]

	return keyset, nil
}

// loadKeyset gets the named Keyset and the format of the Keyset.
func (c *ClientsetCAStore) loadKeyset(ctx context.Context, name string) (*Keyset, error) {
	o, err := c.clientset.Keysets(c.namespace).Get(ctx, name, metav1.GetOptions{})
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
	return keyset, nil
}

// FindPrimary returns the primary KeysetItem in the Keyset
func FindPrimary(keyset *kops.Keyset) *kops.KeysetItem {
	var primary *kops.KeysetItem
	var primaryVersion *big.Int

	primaryId := keyset.Spec.PrimaryID

	for i := range keyset.Spec.Keys {
		item := &keyset.Spec.Keys[i]
		if item.DistrustTimestamp != nil {
			continue
		}

		version, ok := big.NewInt(0).SetString(item.Id, 10)
		if !ok {
			klog.Warningf("Ignoring key item with non-integer version: %q", item.Id)
			continue
		}

		if item.Id == primaryId {
			return item
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			primary = item
			primaryVersion = version
		}
	}
	return primary
}

// FindKeyset implements KeystoreReader.
func (c *ClientsetCAStore) FindKeyset(ctx context.Context, name string) (*Keyset, error) {
	return c.loadKeyset(ctx, name)
}

// FindPrimaryKeypair implements pki.Keystore
func (c *ClientsetCAStore) FindPrimaryKeypair(ctx context.Context, name string) (*pki.Certificate, *pki.PrivateKey, error) {
	keyset, err := c.FindKeyset(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	if keyset == nil {
		return nil, nil, nil
	}
	if keyset.Primary == nil {
		return nil, nil, nil
	}
	if keyset.Primary.Certificate == nil {
		return nil, nil, nil
	}
	return keyset.Primary.Certificate, keyset.Primary.PrivateKey, nil
}

// ListKeysets implements CAStore::ListKeysets
func (c *ClientsetCAStore) ListKeysets() (map[string]*Keyset, error) {
	ctx := context.TODO()
	items := map[string]*Keyset{}

	{
		list, err := c.clientset.Keysets(c.namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error listing Keysets: %v", err)
		}

		for i := range list.Items {
			keyset := &list.Items[i]
			switch keyset.Spec.Type {
			case kops.SecretTypeKeypair:
				item, err := parseKeyset(keyset)
				if err != nil {
					return nil, fmt.Errorf("parsing keyset %q: %w", keyset.Name, err)
				}

				items[keyset.Name] = item

			case kops.SecretTypeSecret:
				continue // Ignore - this is handled by ClientsetSecretStore
			default:
				return nil, fmt.Errorf("unhandled secret type %q: %v", keyset.Spec.Type, err)
			}
		}
	}

	return items, nil
}

// StoreKeyset implements CAStore::StoreKeyset
func (c *ClientsetCAStore) StoreKeyset(ctx context.Context, name string, keyset *Keyset) error {
	return c.storeKeyset(ctx, name, keyset)
}

// storeKeyset saves the specified keyset to the registry.
func (c *ClientsetCAStore) storeKeyset(ctx context.Context, name string, keyset *Keyset) error {
	create := false
	client := c.clientset.Keysets(c.namespace)

	kopsKeyset, err := keyset.ToAPIObject(name)
	if err != nil {
		return err
	}

	oldKeyset, err := client.Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		oldKeyset = nil
		err = nil
	}
	if err == nil {
		if oldKeyset == nil {
			create = true
		} else {
			kopsKeyset.ObjectMeta = oldKeyset.ObjectMeta
		}
	} else {
		return fmt.Errorf("error reading keyset %q: %v", name, err)
	}

	if create {
		if _, err := client.Create(ctx, kopsKeyset, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("error creating keyset %q: %v", name, err)
		}
	} else {
		if _, err := client.Update(ctx, kopsKeyset, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating keyset %q: %v", name, err)
		}
	}
	return nil
}

// addSSHCredential saves the specified SSH Credential to the registry, doing an update or insert
func (c *ClientsetCAStore) addSSHCredential(ctx context.Context, publicKey string) error {
	create := false
	client := c.clientset.SSHCredentials(c.namespace)
	sshCredential, err := client.Get(ctx, "admin", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			sshCredential = nil
		} else {
			return fmt.Errorf("error reading SSHCredential: %v", err)
		}
	}
	if sshCredential == nil {
		sshCredential = &kops.SSHCredential{}
		sshCredential.Name = "admin"
		create = true
	}
	sshCredential.Spec.PublicKey = publicKey
	if create {
		if _, err := client.Create(ctx, sshCredential, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("error creating SSHCredential: %v", err)
		}
	} else {
		if _, err := client.Update(ctx, sshCredential, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating SSHCredential: %v", err)
		}
	}
	return nil
}

// deleteSSHCredential deletes the SSHCredential from the registry.
func (c *ClientsetCAStore) deleteSSHCredential(ctx context.Context) error {
	client := c.clientset.SSHCredentials(c.namespace)
	err := client.Delete(ctx, "admin", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting SSHCredential: %v", err)
	}
	return nil
}

// AddSSHPublicKey implements CAStore::AddSSHPublicKey
func (c *ClientsetCAStore) AddSSHPublicKey(ctx context.Context, pubkey []byte) error {
	_, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
	if err != nil {
		return fmt.Errorf("error parsing SSH public key: %v", err)
	}

	return c.addSSHCredential(ctx, strings.TrimSpace(string(pubkey)))
}

// FindSSHPublicKeys implements CAStore::FindSSHPublicKeys
func (c *ClientsetCAStore) FindSSHPublicKeys() ([]*kops.SSHCredential, error) {
	ctx := context.TODO()

	o, err := c.clientset.SSHCredentials(c.namespace).Get(ctx, "admin", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading SSHCredential: %v", err)
	}
	o.Spec.PublicKey = strings.TrimSpace(o.Spec.PublicKey)

	items := []*kops.SSHCredential{o}
	return items, nil
}

// DeleteSSHCredential implements SSHCredentialStore::DeleteSSHCredential
func (c *ClientsetCAStore) DeleteSSHCredential() error {
	ctx := context.TODO()

	return c.deleteSSHCredential(ctx)
}

func (c *ClientsetCAStore) MirrorTo(ctx context.Context, basedir vfs.Path) error {
	keysets, err := c.ListKeysets()
	if err != nil {
		return err
	}

	for name, keyset := range keysets {
		if err := mirrorKeyset(ctx, c.cluster, basedir, name, keyset); err != nil {
			return err
		}
	}

	sshCredentials, err := c.FindSSHPublicKeys()
	if err != nil {
		return fmt.Errorf("error listing SSHCredentials: %v", err)
	}

	for _, sshCredential := range sshCredentials {
		if err := mirrorSSHCredential(ctx, c.cluster, basedir, sshCredential); err != nil {
			return err
		}
	}

	return nil
}
