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

package dotasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	_ "k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type SSHKey struct {
	ID        *int
	Name      *string
	Lifecycle fi.Lifecycle

	PublicKey fi.Resource

	KeyFingerprint *string
}

var _ fi.CompareWithID = &SSHKey{}
var _ fi.CloudupTaskNormalize = &SSHKey{}

func (e *SSHKey) CompareWithID() *string {
	return e.Name
}

func (e *SSHKey) Find(c *fi.CloudupContext) (*SSHKey, error) {
	ctx := c.Context()

	cloud := c.T.Cloud.(do.DOCloud)

	return e.find(ctx, cloud)
}

func (e *SSHKey) find(ctx context.Context, cloud do.DOCloud) (*SSHKey, error) {
	// We aren't allowed to have two keys with the same fingerprint here.
	// So if we find a matching key, we use that one (with that name).

	k, response, err := cloud.KeysService().GetByFingerprint(ctx, *e.KeyFingerprint)
	if response.StatusCode == 404 {
		if e.IsExistingKey() && *e.Name != "" {
			return nil, fmt.Errorf("unable to find specified SSH key %q", *e.Name)
		}
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing SSH keys: %w", err)
	}

	actual := &SSHKey{
		ID:             &k.ID,
		Name:           &k.Name,
		KeyFingerprint: &k.Fingerprint,
	}

	if fi.ValueOf(actual.KeyFingerprint) == fi.ValueOf(e.KeyFingerprint) {
		klog.V(2).Infof("SSH key fingerprints match; assuming public keys match")
		actual.PublicKey = e.PublicKey
	} else {
		klog.V(2).Infof("Computed SSH key fingerprint mismatch: %q %q", fi.ValueOf(e.KeyFingerprint), fi.ValueOf(actual.KeyFingerprint))
	}
	actual.Lifecycle = e.Lifecycle

	e.ID = actual.ID
	if e.IsExistingKey() && *e.Name != "" {
		e.KeyFingerprint = actual.KeyFingerprint
	}

	// We aren't allowed two keys with the same fingerprint but different names...
	e.Name = actual.Name

	return actual, nil
}

func (e *SSHKey) Normalize(c *fi.CloudupContext) error {
	if e.KeyFingerprint == nil && e.PublicKey != nil {
		publicKey, err := fi.ResourceAsString(e.PublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH public key: %v", err)
		}

		keyFingerprint, err := pki.ComputeAWSKeyFingerprint(publicKey)
		if err != nil {
			return fmt.Errorf("error computing key fingerprint for SSH key: %v", err)
		}
		klog.V(2).Infof("Computed SSH key fingerprint as %q", keyFingerprint)
		e.KeyFingerprint = &keyFingerprint
	}

	return nil
}

func (e *SSHKey) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (s *SSHKey) CheckChanges(a, e, changes *SSHKey) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (e *SSHKey) createKeypair(ctx context.Context, cloud do.DOCloud) error {
	klog.V(2).Infof("Creating SSHKey with Name:%q", *e.Name)

	publicKey, err := fi.ResourceAsString(e.PublicKey)
	if err != nil {
		return err
	}

	req := &godo.KeyCreateRequest{
		Name:      *e.Name,
		PublicKey: publicKey,
	}
	klog.V(2).Infof("creating SSH public key %q", req.Name)
	response, _, err := cloud.KeysService().Create(ctx, req)
	if err != nil {
		return err
	}

	e.KeyFingerprint = &response.Fingerprint
	e.ID = &response.ID

	return nil
}

func (_ *SSHKey) RenderDO(t *do.DOAPITarget, a, e, changes *SSHKey) error {
	ctx := context.TODO()

	if a == nil {
		return e.createKeypair(ctx, t.Cloud)
	}

	return nil
}

type terraformSSHKey struct {
	Name      *string                  `cty:"key_name"`
	PublicKey *terraformWriter.Literal `cty:"public_key"`
}

func (_ *SSHKey) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SSHKey) error {
	// We don't want to render a key definition when we're using one that already exists
	if e.IsExistingKey() {
		return nil
	}
	tfName := strings.Replace(*e.Name, ":", "", -1)
	publicKey, err := t.AddFileResource("digitalocean_ssh_key", tfName, "public_key", e.PublicKey, false)
	if err != nil {
		return fmt.Errorf("error rendering PublicKey: %v", err)
	}

	tf := &terraformSSHKey{
		Name:      e.Name,
		PublicKey: publicKey,
	}

	return t.RenderResource("digitalocean_ssh_key", tfName, tf)
}

// IsExistingKey will be true if the task has been initialized without using a public key
// this is when we want to use a key that is already present in the cloud.
func (e *SSHKey) IsExistingKey() bool {
	return e.PublicKey == nil
}

func (e *SSHKey) TerraformLink() *terraformWriter.Literal {
	if e.NoSSHKey() {
		return nil
	}
	if e.IsExistingKey() {
		return terraformWriter.LiteralFromStringValue(*e.Name)
	}
	tfName := strings.Replace(*e.Name, ":", "", -1)
	return terraformWriter.LiteralProperty("digitalocean_ssh_key", tfName, "id")
}

func (e *SSHKey) NoSSHKey() bool {
	return e.ID == nil && e.Name == nil && e.PublicKey == nil && e.KeyFingerprint == nil
}
