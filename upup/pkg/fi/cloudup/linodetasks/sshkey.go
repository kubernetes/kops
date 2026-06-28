/*
Copyright 2026 The Kubernetes Authors.

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

package linodetasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/linode/linodego"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type SSHKey struct {
	Name      *string
	ID        *int
	Lifecycle fi.Lifecycle
	PublicKey *fi.Resource
}

var _ fi.CloudupTask = &SSHKey{}
var _ fi.CompareWithID = &SSHKey{}

func (s *SSHKey) CompareWithID() *string {
	return s.Name
}

func (s *SSHKey) Find(c *fi.CloudupContext) (*SSHKey, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)
	name := fi.ValueOf(s.Name)
	if name == "" {
		return nil, fmt.Errorf("SSHKey.Name is required")
	}

	keys, err := cloud.Client().ListSSHKeys(c.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) SSH keys: %w", err)
	}

	var matched *linodego.SSHKey
	for i := range keys {
		key := &keys[i]
		if key.Label != name {
			continue
		}

		if matched != nil {
			return nil, fmt.Errorf("found multiple SSH keys named %q", name)
		}
		matched = key
	}

	if matched == nil {
		return nil, nil
	}

	actual := &SSHKey{
		ID:        fi.PtrTo(matched.ID),
		Name:      fi.PtrTo(matched.Label),
		Lifecycle: s.Lifecycle,
	}

	if s.PublicKey != nil {
		expectedPublicKey, err := fi.ResourceAsString(*s.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("error rendering SSH key data: %w", err)
		}

		if strings.TrimSpace(expectedPublicKey) != strings.TrimSpace(matched.SSHKey) {
			return nil, fmt.Errorf("found SSH key %q in Linode (Akamai), but public key data did not match", name)
		}

		// Avoid spurious changes.
		actual.PublicKey = s.PublicKey
	}

	return actual, nil
}

func (e *SSHKey) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *SSHKey) CheckChanges(actual, expected, changes *SSHKey) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.PublicKey != nil {
			return fi.CannotChangeField("PublicKey")
		}
	}

	return nil
}

func (*SSHKey) RenderLinode(t *linode.APITarget, actual, expected, changes *SSHKey) error {
	if actual != nil {
		return nil
	}

	name := fi.ValueOf(expected.Name)
	if name == "" {
		return fi.RequiredField("Name")
	}

	if expected.PublicKey == nil {
		return fi.RequiredField("PublicKey")
	}

	publicKey, err := fi.ResourceAsString(*expected.PublicKey)
	if err != nil {
		return fmt.Errorf("error rendering SSH key data: %w", err)
	}

	created, err := t.Cloud.Client().CreateSSHKey(context.Background(), linodego.SSHKeyCreateOptions{
		Label:  name,
		SSHKey: strings.TrimSpace(publicKey),
	})
	if err != nil {
		return fmt.Errorf("error creating Linode (Akamai) SSH key %q: %w", name, err)
	}

	expected.ID = fi.PtrTo(created.ID)
	klog.V(2).Infof("Created Linode (Akamai) SSH key %q (id=%d)", created.Label, created.ID)

	return nil
}
