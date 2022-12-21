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
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/sshcredentials"
	"k8s.io/kops/util/pkg/vfs"
)

type VFSCAStore struct {
	VFSKeystoreReader
	cluster *kops.Cluster
}

var (
	_ CAStore            = &VFSCAStore{}
	_ SSHCredentialStore = &VFSCAStore{}
)

func NewVFSCAStore(cluster *kops.Cluster, basedir vfs.Path) *VFSCAStore {
	c := &VFSCAStore{
		VFSKeystoreReader: VFSKeystoreReader{
			basedir: basedir,
		},
		cluster: cluster,
	}

	return c
}

// NewVFSSSHCredentialStore creates a SSHCredentialStore backed by VFS
func NewVFSSSHCredentialStore(cluster *kops.Cluster, basedir vfs.Path) SSHCredentialStore {
	// Note currently identical to NewVFSCAStore
	c := &VFSCAStore{
		VFSKeystoreReader: VFSKeystoreReader{
			basedir: basedir,
		},
		cluster: cluster,
	}

	return c
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
func writeKeysetBundle(ctx context.Context, cluster *kops.Cluster, p vfs.Path, name string, keyset *Keyset) error {
	p = p.Join("keyset.yaml")

	o, err := keyset.ToAPIObject(name)
	if err != nil {
		return err
	}

	objectData, err := serializeKeysetBundle(o)
	if err != nil {
		return err
	}

	acl, err := acls.GetACL(ctx, p, cluster)
	if err != nil {
		return err
	}
	return p.WriteFile(ctx, bytes.NewReader(objectData), acl)
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
func (c *VFSCAStore) MirrorTo(ctx context.Context, basedir vfs.Path) error {
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

// mirrorKeyset writes Keyset bundles for the certificates & privatekeys.
func mirrorKeyset(ctx context.Context, cluster *kops.Cluster, basedir vfs.Path, name string, keyset *Keyset) error {
	if err := writeKeysetBundle(ctx, cluster, basedir.Join("private"), name, keyset); err != nil {
		return fmt.Errorf("writing private bundle: %v", err)
	}

	return nil
}

// mirrorSSHCredential writes the SSH credential file to the mirror location
func mirrorSSHCredential(ctx context.Context, cluster *kops.Cluster, basedir vfs.Path, sshCredential *kops.SSHCredential) error {
	id, err := sshcredentials.Fingerprint(sshCredential.Spec.PublicKey)
	if err != nil {
		return fmt.Errorf("error fingerprinting SSH public key %q: %v", sshCredential.Name, err)
	}

	p := basedir.Join("ssh", "public", sshCredential.Name, id)
	acl, err := acls.GetACL(ctx, p, cluster)
	if err != nil {
		return err
	}

	err = p.WriteFile(ctx, bytes.NewReader([]byte(sshCredential.Spec.PublicKey)), acl)
	if err != nil {
		return fmt.Errorf("error writing %q: %v", p, err)
	}

	return nil
}

func (c *VFSCAStore) StoreKeyset(ctx context.Context, name string, keyset *Keyset) error {
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
	if keyset.Items[primaryId].Certificate == nil {
		return fmt.Errorf("keyset's primary id %q must have a certificate", primaryId)
	}

	{
		p := c.buildPrivateKeyPoolPath(name)
		if err := writeKeysetBundle(ctx, c.cluster, p, name, keyset); err != nil {
			return fmt.Errorf("writing private bundle: %v", err)
		}
	}

	return nil
}

// AddSSHPublicKey stores an SSH public key
func (c *VFSCAStore) AddSSHPublicKey(ctx context.Context, pubkey []byte) error {
	id, err := sshcredentials.Fingerprint(strings.TrimSpace(string(pubkey)))
	if err != nil {
		return fmt.Errorf("error fingerprinting SSH public key: %v", err)
	}

	p := c.buildSSHPublicKeyPath(id)

	acl, err := acls.GetACL(ctx, p, c.cluster)
	if err != nil {
		return err
	}

	return p.WriteFile(ctx, bytes.NewReader(pubkey), acl)
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
