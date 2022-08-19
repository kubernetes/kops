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
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/sshcredentials"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSCAStore struct {
	basedir vfs.Path
	cluster *kops.Cluster

	mutex    sync.Mutex
	cachedCA *Keyset
}

var (
	_ CAStore            = &VFSCAStore{}
	_ SSHCredentialStore = &VFSCAStore{}
)

func NewVFSCAStore(cluster *kops.Cluster, basedir vfs.Path) *VFSCAStore {
	c := &VFSCAStore{
		basedir: basedir,
		cluster: cluster,
	}

	return c
}

// NewVFSSSHCredentialStore creates a SSHCredentialStore backed by VFS
func NewVFSSSHCredentialStore(cluster *kops.Cluster, basedir vfs.Path) SSHCredentialStore {
	// Note currently identical to NewVFSCAStore
	c := &VFSCAStore{
		basedir: basedir,
		cluster: cluster,
	}

	return c
}

func (c *VFSCAStore) VFSPath() vfs.Path {
	return c.basedir
}

func (c *VFSCAStore) buildPrivateKeyPoolPath(name string) vfs.Path {
	return c.basedir.Join("private", name)
}

func (c *VFSCAStore) parseKeysetYaml(data []byte) (*kops.Keyset, bool, error) {
	defaultReadVersion := v1alpha2.SchemeGroupVersion.WithKind("Keyset")

	object, gvk, err := kopscodecs.Decode(data, &defaultReadVersion)
	if err != nil {
		return nil, false, fmt.Errorf("error parsing keyset: %v", err)
	}

	keyset, ok := object.(*kops.Keyset)
	if !ok {
		return nil, false, fmt.Errorf("object was not a keyset, was a %T", object)
	}

	if gvk == nil {
		return nil, false, fmt.Errorf("object did not have GroupVersionKind: %q", keyset.Name)
	}

	return keyset, gvk.Version != keysetFormatLatest, nil
}

// loadKeyset loads a Keyset from the path.
// Returns (nil, nil) if the file is not found
// Bundles avoid the need for a list-files permission, which can be tricky on e.g. GCE
func (c *VFSCAStore) loadKeyset(p vfs.Path) (*Keyset, error) {
	bundlePath := p.Join("keyset.yaml")
	data, err := bundlePath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read bundle %q: %v", p, err)
	}

	o, legacyFormat, err := c.parseKeysetYaml(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing bundle %q: %v", p, err)
	}

	keyset, err := parseKeyset(o)
	if err != nil {
		return nil, fmt.Errorf("error mapping bundle %q: %v", p, err)
	}

	keyset.LegacyFormat = legacyFormat
	return keyset, nil
}

func (k *Keyset) ToAPIObject(name string) (*kops.Keyset, error) {
	o := &kops.Keyset{}
	o.Name = name
	o.Spec.Type = kops.SecretTypeKeypair

	keys := make([]string, 0, len(k.Items))
	for k := range k.Items {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return KeysetItemIdOlder(k.Items[keys[i]].Id, k.Items[keys[j]].Id)
	})

	for _, key := range keys {
		ki := k.Items[key]
		var distrustTimestamp *metav1.Time
		if ki.DistrustTimestamp != nil {
			distrustTimestamp = &metav1.Time{Time: *ki.DistrustTimestamp}
		}
		oki := kops.KeysetItem{
			Id:                ki.Id,
			DistrustTimestamp: distrustTimestamp,
		}

		if ki.Certificate != nil {
			var publicMaterial bytes.Buffer
			if _, err := ki.Certificate.WriteTo(&publicMaterial); err != nil {
				return nil, err
			}
			oki.PublicMaterial = publicMaterial.Bytes()
		}

		if ki.PrivateKey != nil {
			var privateMaterial bytes.Buffer
			if _, err := ki.PrivateKey.WriteTo(&privateMaterial); err != nil {
				return nil, err
			}

			oki.PrivateMaterial = privateMaterial.Bytes()
		}

		o.Spec.Keys = append(o.Spec.Keys, oki)
	}
	if k.Primary != nil {
		o.Spec.PrimaryID = k.Primary.Id
	}
	return o, nil
}

// writeKeysetBundle writes a Keyset bundle to VFS.
func writeKeysetBundle(cluster *kops.Cluster, p vfs.Path, name string, keyset *Keyset) error {
	p = p.Join("keyset.yaml")

	o, err := keyset.ToAPIObject(name)
	if err != nil {
		return err
	}

	objectData, err := serializeKeysetBundle(o)
	if err != nil {
		return err
	}

	acl, err := acls.GetACL(p, cluster)
	if err != nil {
		return err
	}
	return p.WriteFile(bytes.NewReader(objectData), acl)
}

// serializeKeysetBundle converts a Keyset bundle to yaml, for writing to VFS.
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

func (c *VFSCAStore) FindPrimaryKeypair(name string) (*pki.Certificate, *pki.PrivateKey, error) {
	return FindPrimaryKeypair(c, name)
}

var legacyKeysetMappings = map[string]string{
	// The strange name is because kOps prior to 1.19 used the api-server TLS key for this.
	"service-account": "master",
	// Renamed in kOps 1.22
	"kubernetes-ca": "ca",
}

func (c *VFSCAStore) FindKeyset(id string) (*Keyset, error) {
	keys, err := c.findPrivateKeyset(id)
	if keys == nil || os.IsNotExist(err) {
		if legacyId := legacyKeysetMappings[id]; legacyId != "" {
			keys, err = c.findPrivateKeyset(legacyId)
			if keys != nil {
				keys.LegacyFormat = true
			}
		}
	}

	return keys, err
}

// ListKeysets implements CAStore::ListKeysets
func (c *VFSCAStore) ListKeysets() (map[string]*Keyset, error) {
	baseDir := c.basedir.Join("private")
	files, err := baseDir.ReadTree()
	if err != nil {
		return nil, fmt.Errorf("error reading directory %q: %v", baseDir, err)
	}

	keysets := map[string]*Keyset{}

	for _, f := range files {
		relativePath, err := vfs.RelativePath(baseDir, f)
		if err != nil {
			return nil, err
		}

		tokens := strings.Split(relativePath, "/")
		if len(tokens) != 2 || tokens[1] != "keyset.yaml" {
			klog.V(2).Infof("ignoring unexpected file in keystore: %q", f)
			continue
		}

		name := tokens[0]
		loadedKeyset, err := c.loadKeyset(baseDir.Join(name))
		if err != nil {
			klog.Warningf("ignoring keyset %q: %w", name, err)
			continue
		}

		keysets[name] = loadedKeyset
	}

	return keysets, nil
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

	for name, keyset := range keysets {
		if err := mirrorKeyset(c.cluster, basedir, name, keyset); err != nil {
			return err
		}
	}

	sshCredentials, err := c.FindSSHPublicKeys()
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

// mirrorKeyset writes Keyset bundles for the certificates & privatekeys.
func mirrorKeyset(cluster *kops.Cluster, basedir vfs.Path, name string, keyset *Keyset) error {
	if err := writeKeysetBundle(cluster, basedir.Join("private"), name, keyset); err != nil {
		return fmt.Errorf("writing private bundle: %v", err)
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

func (c *VFSCAStore) StoreKeyset(name string, keyset *Keyset) error {
	if keyset.Primary == nil || keyset.Primary.Id == "" {
		return fmt.Errorf("keyset must have a primary key")
	}
	primaryId := keyset.Primary.Id
	if keyset.Items[primaryId] == nil {
		return fmt.Errorf("keyset's primary id %q not present in items", primaryId)
	}
	if keyset.Items[primaryId].DistrustTimestamp != nil {
		return fmt.Errorf("keyset's primary id %q must not be distrusted", primaryId)
	}
	if keyset.Items[primaryId].PrivateKey == nil {
		return fmt.Errorf("keyset's primary id %q must have a private key", primaryId)
	}

	{
		p := c.buildPrivateKeyPoolPath(name)
		if err := writeKeysetBundle(c.cluster, p, name, keyset); err != nil {
			return fmt.Errorf("writing private bundle: %v", err)
		}
	}

	return nil
}

func (c *VFSCAStore) findPrivateKeyset(id string) (*Keyset, error) {
	var keys *Keyset
	var err error
	if id == CertificateIDCA {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		cached := c.cachedCA
		if cached != nil {
			return cached, nil
		}

		keys, err = c.loadKeyset(c.buildPrivateKeyPoolPath(id))
		if err != nil {
			return nil, err
		}

		if keys == nil {
			klog.Warningf("CA private key was not found")
			// We no longer generate CA certificates automatically - too race-prone
		} else {
			c.cachedCA = keys
		}
	} else {
		p := c.buildPrivateKeyPoolPath(id)
		keys, err = c.loadKeyset(p)
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}

// AddSSHPublicKey stores an SSH public key
func (c *VFSCAStore) AddSSHPublicKey(pubkey []byte) error {
	id, err := sshcredentials.Fingerprint(strings.TrimSpace(string(pubkey)))
	if err != nil {
		return fmt.Errorf("error fingerprinting SSH public key: %v", err)
	}

	p := c.buildSSHPublicKeyPath(id)

	acl, err := acls.GetACL(p, c.cluster)
	if err != nil {
		return err
	}

	return p.WriteFile(bytes.NewReader(pubkey), acl)
}

func (c *VFSCAStore) buildSSHPublicKeyPath(id string) vfs.Path {
	// id is fingerprint with colons, but we store without colons
	id = strings.Replace(id, ":", "", -1)
	return c.basedir.Join("ssh", "public", "admin", id)
}

func (c *VFSCAStore) FindSSHPublicKeys() ([]*kops.SSHCredential, error) {
	p := c.basedir.Join("ssh", "public", "admin")

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
		item.Name = "admin"
		item.Spec.PublicKey = strings.TrimSpace(string(data))
		items = append(items, item)
	}

	return items, nil
}

func (c *VFSCAStore) DeleteSSHCredential() error {
	p := c.basedir.Join("ssh", "public", "admin")

	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, f := range files {
		if err := f.Remove(); err != nil {
			return err
		}
	}
	return nil
}
