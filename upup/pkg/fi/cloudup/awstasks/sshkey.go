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

package awstasks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"

	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
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
	cloud := c.Cloud.(awsup.AWSCloud)

	return e.find(cloud)
}

func (e *SSHKey) find(cloud awsup.AWSCloud) (*SSHKey, error) {
	request := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{e.Name},
	}

	response, err := cloud.EC2().DescribeKeyPairs(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "InvalidKeyPair.NotFound" {
			return nil, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error listing SSHKeys: %v", err)
	}

	if response == nil || len(response.KeyPairs) == 0 {
		return nil, nil
	}

	if len(response.KeyPairs) != 1 {
		return nil, fmt.Errorf("Found multiple SSHKeys with Name %q", *e.Name)
	}

	k := response.KeyPairs[0]

	actual := &SSHKey{
		Name:           k.KeyName,
		KeyFingerprint: k.KeyFingerprint,
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

		keyFingerprint, err := pki.ComputeAWSKeyFingerprint(publicKey)
		if err != nil {
			return fmt.Errorf("error computing key fingerprint for SSH key: %v", err)
		}
		klog.V(2).Infof("Computed SSH key fingerprint as %q", keyFingerprint)
		e.KeyFingerprint = &keyFingerprint
	} else if e.IsExistingKey() && *e.Name != "" {
		a, err := e.Find(c)
		if err != nil {
			return err
		}
		if a == nil {
			return fmt.Errorf("unable to find specified SSH key %q", *e.Name)
		}

		e.KeyFingerprint = a.KeyFingerprint
	}
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *SSHKey) CheckChanges(a, e, changes *SSHKey) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (e *SSHKey) createKeypair(cloud awsup.AWSCloud) error {
	klog.V(2).Infof("Creating SSHKey with Name:%q", *e.Name)

	request := &ec2.ImportKeyPairInput{
		KeyName: e.Name,
	}

	if e.PublicKey != nil {
		d, err := e.PublicKey.AsBytes()
		if err != nil {
			return fmt.Errorf("error rendering SSHKey PublicKey: %v", err)
		}
		request.PublicKeyMaterial = d
	}

	response, err := cloud.EC2().ImportKeyPair(request)
	if err != nil {
		return fmt.Errorf("error creating SSHKey: %v", err)
	}

	e.KeyFingerprint = response.KeyFingerprint

	return nil
}

func (_ *SSHKey) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SSHKey) error {
	if a == nil {
		return e.createKeypair(t.Cloud)
	}

	// No tags on SSH public key
	return nil //return output.AddAWSTags(cloud.Tags(), v, "vpc")
}

type terraformSSHKey struct {
	Name      *string            `json:"key_name"`
	PublicKey *terraform.Literal `json:"public_key"`
}

func (_ *SSHKey) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SSHKey) error {
	// We don't want to render a key definition when we're using one that already exists
	if e.IsExistingKey() {
		return nil
	}
	tfName := strings.Replace(*e.Name, ":", "", -1)
	publicKey, err := t.AddFile("aws_key_pair", tfName, "public_key", e.PublicKey)
	if err != nil {
		return fmt.Errorf("error rendering PublicKey: %v", err)
	}

	tf := &terraformSSHKey{
		Name:      e.Name,
		PublicKey: publicKey,
	}

	return t.RenderResource("aws_key_pair", tfName, tf)
}

// IsExistingKey will be true if the task has been initialized without using a public key
// this is when we want to use a key that is already present in AWS.
func (e *SSHKey) IsExistingKey() bool {
	return e.PublicKey == nil
}

func (e *SSHKey) TerraformLink() *terraform.Literal {
	if e.NoSSHKey() {
		return nil
	}
	if e.IsExistingKey() {
		return terraform.LiteralFromStringValue(*e.Name)
	}
	tfName := strings.Replace(*e.Name, ":", "", -1)
	return terraform.LiteralProperty("aws_key_pair", tfName, "id")
}

func (_ *SSHKey) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *SSHKey) error {
	if e.NoSSHKey() {
		return nil
	}

	cloud := t.Cloud.(awsup.AWSCloud)

	klog.Warningf("Cloudformation does not manage SSH keys; pre-creating SSH key")

	keypair, err := e.find(cloud)
	if err != nil {
		return err
	}

	if keypair == nil {
		err := e.createKeypair(cloud)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *SSHKey) NoSSHKey() bool {
	return *e == SSHKey{}
}
