package awstasks

import (
	"fmt"

	"bytes"
	"crypto"
	"crypto/dsa"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"reflect"
	"strings"
)

//go:generate fitask -type=SSHKey
type SSHKey struct {
	Name *string

	PublicKey *fi.ResourceHolder

	KeyFingerprint *string
}

var _ fi.CompareWithID = &SSHKey{}

func (e *SSHKey) CompareWithID() *string {
	return e.Name
}

func (e *SSHKey) Find(c *fi.Context) (*SSHKey, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{e.Name},
	}

	response, err := cloud.EC2.DescribeKeyPairs(request)
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

	return actual, nil
}

func computeAwsKeyFingerprint(publicKey *fi.ResourceHolder) (string, error) {
	publicKeyString, err := publicKey.AsString()
	if err != nil {
		return "", fmt.Errorf("error reading SSH public key: %v", err)
	}

	tokens := strings.Split(publicKeyString, " ")
	if len(tokens) < 2 {
		return "", fmt.Errorf("error parsing SSH public key: %s", publicKeyString)
	}

	sshPublicKeyBytes, err := base64.StdEncoding.DecodeString(tokens[1])
	if len(tokens) < 2 {
		return "", fmt.Errorf("error decoding SSH public key: %s", publicKeyString)
	}

	sshPublicKey, err := ssh.ParsePublicKey(sshPublicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("error parsing SSH public key: %v", err)
	}

	der, err := toDER(sshPublicKey)
	if err != nil {
		return "", fmt.Errorf("error computing fingerprint for SSH public key: %v", err)
	}
	h := md5.Sum(der)
	sshKeyFingerprint := fmt.Sprintf("%x", h)

	var colonSeparated bytes.Buffer
	for i := 0; i < len(sshKeyFingerprint); i++ {
		if (i%2) == 0 && i != 0 {
			colonSeparated.WriteByte(':')
		}
		colonSeparated.WriteByte(sshKeyFingerprint[i])
	}

	return colonSeparated.String(), nil
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

	case "*dsaPublicKey":
		var dsaPublicKey *dsa.PublicKey
		targetType := reflect.ValueOf(dsaPublicKey).Type()
		dsaPublicKey = pubkeyValue.Convert(targetType).Interface().(*dsa.PublicKey)
		cryptoKey = dsaPublicKey

	default:
		return nil, fmt.Errorf("Unknown type for SSH PublicKey; cannot compute fingerprint: %q", typeName)
	}

	der, err := x509.MarshalPKIXPublicKey(cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("error marshalling SSH public key: %v", err)
	}
	return der, nil
}

func (e *SSHKey) Run(c *fi.Context) error {
	if e.KeyFingerprint == nil && e.PublicKey != nil {
		keyFingerprint, err := computeAwsKeyFingerprint(e.PublicKey)
		if err != nil {
			return fmt.Errorf("error computing key fingerpring for SSH key: %v", err)
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

func (_ *SSHKey) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SSHKey) error {
	if a == nil {
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

		response, err := t.Cloud.EC2.ImportKeyPair(request)
		if err != nil {
			return fmt.Errorf("error creating SSHKey: %v", err)
		}

		e.KeyFingerprint = response.KeyFingerprint
	}

	return nil //return output.AddAWSTags(cloud.Tags(), v, "vpc")
}

type terraformSSHKey struct {
	Name      *string            `json:"key_name"`
	PublicKey *terraform.Literal `json:"public_key"`
}

func (_ *SSHKey) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SSHKey) error {
	publicKey, err := t.AddFile("aws_key_pair", *e.Name, "public_key", e.PublicKey)
	if err != nil {
		return fmt.Errorf("error rendering PublicKey: %v", err)
	}

	tf := &terraformSSHKey{
		Name:      e.Name,
		PublicKey: publicKey,
	}

	return t.RenderResource("aws_key_pair", *e.Name, tf)
}

func (e *SSHKey) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_key_pair", *e.Name, "id")
}
