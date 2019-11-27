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

package openstacktasks

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"k8s.io/klog"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=SSHKey
type SSHKey struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	PublicKey *fi.ResourceHolder

	KeyFingerprint *string
}

var _ fi.CompareWithID = &SSHKey{}

func (e *SSHKey) CompareWithID() *string {
	return e.Name
}

func (e *SSHKey) Find(c *fi.Context) (*SSHKey, error) {
	cloud := c.Cloud.(openstack.OpenstackCloud)
	rs, err := cloud.GetKeypair(openstackKeyPairName(fi.StringValue(e.Name)))
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, nil
	}
	actual := &SSHKey{
		Name:           e.Name,
		KeyFingerprint: fi.String(rs.Fingerprint),
	}

	// Avoid spurious changes
	if fi.StringValue(actual.KeyFingerprint) == fi.StringValue(e.KeyFingerprint) {
		klog.V(2).Infof("SSH key fingerprints match; assuming public keys match")
		actual.PublicKey = e.PublicKey
	} else {
		klog.V(2).Infof("Computed SSH key fingerprint mismatch: %q %q", fi.StringValue(e.KeyFingerprint), fi.StringValue(actual.KeyFingerprint))
	}
	actual.Lifecycle = e.Lifecycle
	return actual, nil
}

func (e *SSHKey) Run(c *fi.Context) error {
	if e.KeyFingerprint == nil && e.PublicKey != nil {
		publicKey, err := e.PublicKey.AsString()
		if err != nil {
			return fmt.Errorf("error reading SSH public key: %v", err)
		}

		keyFingerprint, err := pki.ComputeOpenSSHKeyFingerprint(publicKey)
		if err != nil {
			return fmt.Errorf("error computing key fingerprint for SSH key: %v", err)
		}
		klog.V(2).Infof("Computed SSH key fingerprint as %q", keyFingerprint)
		e.KeyFingerprint = &keyFingerprint
	}
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *SSHKey) CheckChanges(a, e, changes *SSHKey) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.KeyFingerprint != nil {
			return fi.CannotChangeField("KeyFingerprint")
		}
	}
	return nil
}

func openstackKeyPairName(org string) string {
	name := strings.Replace(org, ".", "-", -1)
	name = strings.Replace(name, ":", "_", -1)
	return name
}

func (_ *SSHKey) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *SSHKey) error {
	if a == nil {
		klog.V(2).Infof("Creating Keypair with name:%q", fi.StringValue(e.Name))

		opt := keypairs.CreateOpts{
			Name: openstackKeyPairName(fi.StringValue(e.Name)),
		}

		if e.PublicKey != nil {
			d, err := e.PublicKey.AsString()
			if err != nil {
				return fmt.Errorf("error rendering SSHKey PublicKey: %v", err)
			}
			opt.PublicKey = d
		}

		v, err := t.Cloud.CreateKeypair(opt)
		if err != nil {
			return fmt.Errorf("Error creating keypair: %v", err)
		}

		e.KeyFingerprint = fi.String(v.Fingerprint)
		klog.V(2).Infof("Creating a new Openstack keypair, id=%s", v.Fingerprint)
		return nil
	}
	e.KeyFingerprint = a.KeyFingerprint
	klog.V(2).Infof("Using an existing Openstack keypair, id=%s", fi.StringValue(e.KeyFingerprint))
	return nil
}
