/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaytasks

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"

	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// +kops:fitask
type SSHKey struct {
	ID                 *string
	Name               *string
	Lifecycle          fi.Lifecycle
	PublicKey          *fi.Resource
	KeyPairFingerPrint *string
}

var _ fi.CompareWithID = &SSHKey{}

func (s *SSHKey) CompareWithID() *string {
	return s.Name
}

func (s *SSHKey) Find(c *fi.Context) (*SSHKey, error) {
	cloud := c.Cloud.(scaleway.ScwCloud)

	keysResp, err := cloud.AccountService().ListSSHKeys(&account.ListSSHKeysRequest{
		Name: s.Name,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("error listing SSH keys: %v", err)
	}
	if keysResp.TotalCount == 0 {
		return nil, nil
	}
	if keysResp.TotalCount != 1 {
		return nil, fmt.Errorf("found multiple SSH keys named %q", *s.Name)
	}

	klog.V(2).Infof("found matching SSH key named %q", *s.Name)
	k := keysResp.SSHKeys[0]
	sshKey := &SSHKey{
		ID:                 fi.PtrTo(k.ID),
		Name:               fi.PtrTo(k.Name),
		KeyPairFingerPrint: fi.PtrTo(k.Fingerprint),
		Lifecycle:          s.Lifecycle,
	}

	// Avoid spurious changes
	if !strings.Contains(fi.ValueOf(sshKey.KeyPairFingerPrint), fi.ValueOf(s.KeyPairFingerPrint)) {
		return nil, fmt.Errorf("computed SSH key fingerprint mismatch: %q %q", fi.ValueOf(s.KeyPairFingerPrint), fi.ValueOf(sshKey.KeyPairFingerPrint))
	}

	klog.V(2).Infof("SSH key fingerprints match; assuming public keys match")
	sshKey.PublicKey = s.PublicKey
	sshKey.KeyPairFingerPrint = s.KeyPairFingerPrint

	return sshKey, nil
}

func (s *SSHKey) Run(c *fi.Context) error {
	if s.KeyPairFingerPrint == nil && s.PublicKey != nil {
		publicKey, err := fi.ResourceAsString(*s.PublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH public key: %w", err)
		}

		keyPairFingerPrint, err := pki.ComputeOpenSSHKeyFingerprint(publicKey)
		if err != nil {
			return fmt.Errorf("error computing key fingerprint for SSH key: %v", err)
		}
		klog.V(2).Infof("Computed SSH key fingerprint as %q", keyPairFingerPrint)
		s.KeyPairFingerPrint = &keyPairFingerPrint
	}
	return fi.DefaultDeltaRunMethod(s, c)
}

func (s *SSHKey) CheckChanges(actual, expected, changes *SSHKey) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (*SSHKey) RenderScw(c *fi.Context, actual, expected, changes *SSHKey) error {
	if actual != nil {
		klog.Infof("Scaleway does not support changes to ssh keys for the moment")
		return nil
	}

	cloud := c.Cloud.(scaleway.ScwCloud)

	name := fi.ValueOf(expected.Name)
	if name == "" {
		return fi.RequiredField("Name")
	}
	klog.V(2).Infof("Creating keypair with name: %q", name)

	keyArgs := &account.CreateSSHKeyRequest{
		Name: name}
	if expected.PublicKey != nil {
		d, err := fi.ResourceAsString(*expected.PublicKey)
		if err != nil {
			return fmt.Errorf("error rendering SSH public key: %w", err)
		}
		keyArgs.PublicKey = d
	}

	key, err := cloud.AccountService().CreateSSHKey(keyArgs)
	if err != nil {
		return fmt.Errorf("error creating SSH keypair: %w", err)
	}
	expected.KeyPairFingerPrint = fi.PtrTo(key.Fingerprint)
	klog.V(2).Infof("Created a new SSH keypair, id=%q fingerprint=%q", key.ID, key.Fingerprint)

	return nil
}
