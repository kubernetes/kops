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

package hetznertasks

import (
	"context"
	"net/mail"
	"strconv"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type SSHKey struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID        *int
	PublicKey string

	Labels map[string]string
}

var _ fi.CompareWithID = &SSHKey{}

func (v *SSHKey) CompareWithID() *string {
	return fi.String(strconv.Itoa(fi.IntValue(v.ID)))
}

func (v *SSHKey) Find(c *fi.Context) (*SSHKey, error) {
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.SSHKeyClient()

	sshkeys, err := client.All(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, sshkey := range sshkeys {
		fingerprint, err := pki.ComputeOpenSSHKeyFingerprint(v.PublicKey)
		if err != nil {
			return nil, err
		}
		if sshkey.Fingerprint == fingerprint {
			matches := &SSHKey{
				Name:      v.Name,
				Lifecycle: v.Lifecycle,
				ID:        fi.Int(sshkey.ID),
				PublicKey: sshkey.PublicKey,
				Labels:    v.Labels,
			}
			v.ID = matches.ID
			return matches, nil
		}
	}

	return nil, nil
}

func (v *SSHKey) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *SSHKey) CheckChanges(a, e, changes *SSHKey) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.PublicKey != "" {
			return fi.CannotChangeField("PublicKey")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.PublicKey == "" {
			return fi.RequiredField("PublicKey")
		}
	}
	return nil
}

func (_ *SSHKey) RenderHetzner(t *hetzner.HetznerAPITarget, a, e, changes *SSHKey) error {
	client := t.Cloud.SSHKeyClient()
	if a == nil {
		name := fi.StringValue(e.Name)
		tokens := strings.Fields(e.PublicKey)
		if len(tokens) == 3 {
			sshkeyComment := tokens[2]
			_, err := mail.ParseAddress(sshkeyComment)
			if err == nil {
				name = sshkeyComment
			}
		}
		opts := hcloud.SSHKeyCreateOpts{
			Name:      name,
			PublicKey: e.PublicKey,
			Labels:    e.Labels,
		}
		sshkey, _, err := client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}
		e.ID = fi.Int(sshkey.ID)
	}

	return nil
}

type terraformSSHKey struct {
	Name      *string           `cty:"name"`
	PublicKey *string           `cty:"public_key"`
	Labels    map[string]string `cty:"labels"`
}

func (_ *SSHKey) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SSHKey) error {
	tf := &terraformSSHKey{
		Name:      e.Name,
		PublicKey: fi.String(e.PublicKey),
		Labels:    e.Labels,
	}

	return t.RenderResource("hcloud_ssh_key", *e.Name, tf)
}

func (e *SSHKey) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("hcloud_ssh_key", *e.Name, "id")
}
