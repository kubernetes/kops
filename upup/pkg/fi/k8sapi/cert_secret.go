package k8sapi

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/api/v1"
)

// KeypairSecret is a wrapper around a k8s Secret object that holds a TLS keypair
type KeypairSecret struct {
	Namespace string
	Name      string

	Certificate *fi.Certificate
	PrivateKey  *fi.PrivateKey
}

// ParseKeypairSecret parses the secret object, decoding the certificate & private-key, if present
func ParseKeypairSecret(secret *v1.Secret) (*KeypairSecret, error) {
	k := &KeypairSecret{}
	k.Namespace = secret.Namespace
	k.Name = secret.Name

	certData := secret.Data[v1.TLSCertKey]
	if certData != nil {
		cert, err := fi.LoadPEMCertificate(certData)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate in %s/%s: %q", k.Namespace, k.Name, err)
		}
		k.Certificate = cert
	}
	keyData := secret.Data[v1.TLSPrivateKeyKey]
	if keyData != nil {
		key, err := fi.ParsePEMPrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("error parsing key in %s/%s: %q", k.Namespace, k.Name, err)
		}
		k.PrivateKey = key
	}

	return k, nil
}

// Encode maps a KeypairSecret into a k8s Secret
func (k *KeypairSecret) Encode() (*v1.Secret, error) {
	secret := &v1.Secret{}
	secret.Namespace = k.Namespace
	secret.Name = k.Name
	secret.Type = v1.SecretTypeTLS

	secret.Data = make(map[string][]byte)

	if k.Certificate != nil {
		data, err := k.Certificate.AsBytes()
		if err != nil {
			return nil, err
		}
		secret.Data[v1.TLSCertKey] = data
	}

	if k.PrivateKey != nil {
		data, err := k.PrivateKey.AsBytes()
		if err != nil {
			return nil, err
		}
		secret.Data[v1.TLSPrivateKeyKey] = data
	}

	return secret, nil
}
