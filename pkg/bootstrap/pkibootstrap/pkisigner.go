/*
Copyright 2021 The Kubernetes Authors.

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

package pkibootstrap

import (
	"bytes"
	"crypto"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/pki"
	gcetpm "k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm"
)

type pkiAuthenticator struct {
	signer   crypto.Signer
	keyID    string
	hostname string
}

var _ bootstrap.Authenticator = &pkiAuthenticator{}

func NewAuthenticator(hostname string, signer crypto.Signer) (bootstrap.Authenticator, error) {
	keyID, err := computeKeyID(signer)
	if err != nil {
		return nil, err
	}

	return &pkiAuthenticator{hostname: hostname, signer: signer, keyID: keyID}, nil
}

func computeKeyID(signer crypto.Signer) (string, error) {
	publicKey := signer.Public()
	pkData, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("error converting public key to x509: %w", err)
	}

	var b bytes.Buffer
	if err := pem.Encode(&b, &pem.Block{Type: "PUBLIC KEY", Bytes: pkData}); err != nil {
		return "", fmt.Errorf("error encoding public key: %w", err)
	}
	return b.String(), nil
}

func NewAuthenticatorFromFile(p string) (bootstrap.Authenticator, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("couldn't determine hostname: %w", err)
	}

	keyBytes, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", p, err)
	}
	key, err := pki.ParsePEMPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing key from %q: %w", p, err)
	}

	return NewAuthenticator(hostname, key.Key)
}

func (a *pkiAuthenticator) CreateToken(body []byte) (string, error) {
	requestHash := sha256.Sum256(body)

	data := gcetpm.AuthTokenData{
		Timestamp:   time.Now().Unix(),
		Audience:    gcetpm.AudienceNodeAuthentication,
		RequestHash: requestHash[:],

		KeyID:    a.keyID,
		Instance: a.hostname,
	}

	payload, err := json.Marshal(&data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	signature, err := a.sign(payload)
	if err != nil {
		return "", fmt.Errorf("failed to sign token data: %w", err)
	}
	token := &gcetpm.AuthToken{
		Data:      payload,
		Signature: signature,
	}

	b, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	return AuthenticationTokenPrefix + base64.StdEncoding.EncodeToString(b), nil
}

// sign performs a TPM signature with the tpmKey, and sanity checks the result.
func (a *pkiAuthenticator) sign(payload []byte) ([]byte, error) {
	beforeSign := time.Now()

	digest := sha256.Sum256(payload)

	signature, err := a.signer.Sign(cryptorand.Reader, digest[:], crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	klog.Infof("signing took %v", time.Since(beforeSign))

	return signature, nil
}
