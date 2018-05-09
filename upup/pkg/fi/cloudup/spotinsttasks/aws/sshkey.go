/*
Copyright 2016 The Kubernetes Authors.

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

package aws

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
	"k8s.io/kops/upup/pkg/fi/utils"
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
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)
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
		glog.V(2).Infof("SSH key fingerprints match; assuming public keys match")
		actual.PublicKey = e.PublicKey
	} else {
		glog.V(2).Infof("Computed SSH key fingerprint mismatch: %q %q", fi.StringValue(e.KeyFingerprint), fi.StringValue(actual.KeyFingerprint))
	}
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

// parseSSHPublicKey parses the SSH public key string
func parseSSHPublicKey(publicKey string) (ssh.PublicKey, error) {
	tokens := strings.Fields(publicKey)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("error parsing SSH public key: %q", publicKey)
	}

	sshPublicKeyBytes, err := base64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding SSH public key: %q err: %s", publicKey, err)
	}
	if len(tokens) < 2 {
		return nil, fmt.Errorf("error decoding SSH public key: %q", publicKey)
	}

	sshPublicKey, err := ssh.ParsePublicKey(sshPublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing SSH public key: %v", err)
	}
	return sshPublicKey, nil
}

// colonSeparatedHex formats the byte slice SSH-fingerprint style: hex bytes separated by colons
func colonSeparatedHex(data []byte) string {
	sshKeyFingerprint := fmt.Sprintf("%x", data)
	var colonSeparated bytes.Buffer
	for i := 0; i < len(sshKeyFingerprint); i++ {
		if (i%2) == 0 && i != 0 {
			colonSeparated.WriteByte(':')
		}
		colonSeparated.WriteByte(sshKeyFingerprint[i])
	}

	return colonSeparated.String()
}

// computeAWSKeyFingerprint computes the AWS-specific fingerprint of the SSH public key
func computeAWSKeyFingerprint(publicKey string) (string, error) {
	sshPublicKey, err := parseSSHPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	der, err := toDER(sshPublicKey)
	if err != nil {
		return "", fmt.Errorf("error computing fingerprint for SSH public key: %v", err)
	}
	h := md5.Sum(der)

	return colonSeparatedHex(h[:]), nil
}

// ComputeOpenSSHKeyFingerprint computes the OpenSSH fingerprint of the SSH public key
func ComputeOpenSSHKeyFingerprint(publicKey string) (string, error) {
	sshPublicKey, err := parseSSHPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	h := md5.Sum(sshPublicKey.Marshal())
	return colonSeparatedHex(h[:]), nil
}

// toDER gets the DER encoding of the SSH public key
// Annoyingly, the ssh code wraps the actual crypto keys, so we have to use reflection tricks
func toDER(pubkey ssh.PublicKey) ([]byte, error) {
	pubkeyValue := reflect.ValueOf(pubkey)
	typeName := utils.BuildTypeName(pubkeyValue.Type())

	var cryptoKey crypto.PublicKey
	switch typeName {
	case "*rsaPublicKey":
		var rsaPublicKey *rsa.PublicKey
		targetType := reflect.ValueOf(rsaPublicKey).Type()
		rsaPublicKey = pubkeyValue.Convert(targetType).Interface().(*rsa.PublicKey)
		cryptoKey = rsaPublicKey

	//case "*dsaPublicKey":
	//	var dsaPublicKey *dsa.PublicKey
	//	targetType := reflect.ValueOf(dsaPublicKey).Type()
	//	dsaPublicKey = pubkeyValue.Convert(targetType).Interface().(*dsa.PublicKey)
	//	cryptoKey = dsaPublicKey

	default:
		return nil, fmt.Errorf("Unexpected type of SSH key (%q); AWS can only import RSA keys", typeName)
	}

	der, err := x509.MarshalPKIXPublicKey(cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("error marshalling SSH public key: %v", err)
	}
	return der, nil
}

func (e *SSHKey) Run(c *fi.Context) error {
	if e.KeyFingerprint == nil && e.PublicKey != nil {
		publicKey, err := e.PublicKey.AsString()
		if err != nil {
			return fmt.Errorf("error reading SSH public key: %v", err)
		}

		keyFingerprint, err := computeAWSKeyFingerprint(publicKey)
		if err != nil {
			return fmt.Errorf("error computing key fingerprint for SSH key: %v", err)
		}
		glog.V(2).Infof("Computed SSH key fingerprint as %q", keyFingerprint)
		e.KeyFingerprint = &keyFingerprint
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
	glog.V(2).Infof("Creating SSHKey with Name:%q", *e.Name)

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

func (_ *SSHKey) Render(t *spotinst.Target, a, e, changes *SSHKey) error {
	if a == nil {
		return e.createKeypair(t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud))
	}

	// No tags on SSH public key
	return nil //return output.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(cloud.Tags(), v, "vpc")
}
