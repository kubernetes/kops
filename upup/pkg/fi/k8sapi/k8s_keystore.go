package k8sapi

import (
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"math/big"
	"time"
)

type KubernetesKeystore struct {
	client    release_1_3.Interface
	namespace string

	//mutex     sync.Mutex
	//cacheCaCertificates *certificates
	//cacheCaPrivateKeys  *privateKeys
}

var _ fi.Keystore = &KubernetesKeystore{}

func NewKubernetesKeystore(client release_1_3.Interface, namespace string) fi.Keystore {
	c := &KubernetesKeystore{
		client:    client,
		namespace: namespace,
	}

	return c
}

func (c *KubernetesKeystore) issueCert(id string, serial *big.Int, privateKey *fi.PrivateKey, template *x509.Certificate) (*fi.Certificate, error) {
	glog.Infof("Issuing new certificate: %q", id)

	template.SerialNumber = serial

	caCert, caKey, err := c.FindKeypair(fi.CertificateId_CA)
	if err != nil {
		return nil, err
	}

	if caCert == nil || caCert.Certificate == nil || caKey == nil || caKey.Key == nil {
		return nil, fmt.Errorf("CA keypair was not found; cannot issue certificates")
	}

	cert, err := fi.SignNewCertificate(privateKey, template, caCert.Certificate, caKey)
	if err != nil {
		return nil, err
	}

	err = c.StoreKeypair(id, cert, privateKey)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *KubernetesKeystore) findSecret(id string) (*v1.Secret, error) {
	secret, err := c.client.Core().Secrets(c.namespace).Get(id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading secret %s/%s from kubernetes: %v", c.namespace, id, err)
	}
	return secret, nil
}

func (c *KubernetesKeystore) FindKeypair(id string) (*fi.Certificate, *fi.PrivateKey, error) {
	secret, err := c.findSecret(id)
	if err != nil {
		return nil, nil, err
	}

	if secret == nil {
		return nil, nil, nil
	}

	keypair, err := ParseKeypairSecret(secret)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing secret %s/%s from kubernetes: %v", c.namespace, id, err)
	}

	return keypair.Certificate, keypair.PrivateKey, nil
}

func (c *KubernetesKeystore) CreateKeypair(id string, template *x509.Certificate) (*fi.Certificate, *fi.PrivateKey, error) {
	t := time.Now().UnixNano()
	serial := fi.BuildPKISerial(t)

	rsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	privateKey := &fi.PrivateKey{Key: rsaKey}
	cert, err := c.issueCert(id, serial, privateKey, template)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func (c *KubernetesKeystore) StoreKeypair(id string, cert *fi.Certificate, privateKey *fi.PrivateKey) error {
	keypair := &KeypairSecret{
		Namespace:   c.namespace,
		Name:        id,
		Certificate: cert,
		PrivateKey:  privateKey,
	}

	secret, err := keypair.Encode()
	createdSecret, err := c.client.Core().Secrets(c.namespace).Create(secret)
	if err != nil {
		return fmt.Errorf("error creating secret %s/%s: %v", secret.Namespace, secret.Name, err)
	}

	created, err := ParseKeypairSecret(createdSecret)
	if err != nil {
		return fmt.Errorf("created secret did not round-trip (%s/%s): %v", c.namespace, id, err)
	}
	if created == nil || created.Certificate == nil || created.PrivateKey == nil {
		return fmt.Errorf("created secret did not round-trip (%s/%s): could not read back", c.namespace, id)
	}

	return err
}
