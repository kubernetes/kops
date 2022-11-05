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

package gcetpmsigner

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/google/go-tpm-tools/client"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	gcetpm "k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm"
)

type tpmAuthenticator struct {
	projectID string
	zone      string
	instance  string
}

var _ bootstrap.Authenticator = &tpmAuthenticator{}

func NewTPMAuthenticator() (bootstrap.Authenticator, error) {
	projectID, err := metadata.ProjectID()
	if err != nil {
		return nil, fmt.Errorf("error getting projectID from metadata: %w", err)
	}
	zone, err := metadata.Zone()
	if err != nil {
		return nil, fmt.Errorf("error getting zone from metadata: %w", err)
	}
	instance, err := metadata.InstanceName()
	if err != nil {
		return nil, fmt.Errorf("error getting instance from metadata: %w", err)
	}

	return &tpmAuthenticator{
		projectID: projectID,
		zone:      zone,
		instance:  instance,
	}, nil
}

func (a *tpmAuthenticator) CreateToken(body []byte) (string, error) {
	requestHash := sha256.Sum256(body)

	tpmStart := time.Now()

	tpmDevice, err := openTPM()
	if err != nil {
		return "", fmt.Errorf("failed to open TPM: %w", err)
	}
	defer tpmDevice.Close()

	key, err := client.GceAttestationKeyRSA(tpmDevice)
	if err != nil {
		return "", fmt.Errorf("failed to get GCP RSA attestation key from TPM: %w", err)
	}
	defer key.Close()

	klog.V(2).Infof("attestation key is %v", debugToPEM(key.PublicKey()))

	klog.Infof("TPM initialization took %v", time.Since(tpmStart))

	data := gcetpm.AuthTokenData{
		GCPProjectID: a.projectID,
		Zone:         a.zone,
		Instance:     a.instance,
		Timestamp:    time.Now().Unix(),
		Audience:     gcetpm.AudienceNodeAuthentication,
		RequestHash:  requestHash[:],
	}

	payload, err := json.Marshal(&data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	signature, err := tpmSign(key, payload)
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
	return gcetpm.GCETPMAuthenticationTokenPrefix + base64.StdEncoding.EncodeToString(b), nil
}

// tpmSign performs a TPM signature with the tpmKey, and sanity checks the result.
func tpmSign(tpmKey *client.Key, payload []byte) ([]byte, error) {
	beforeSign := time.Now()
	signature, err := tpmKey.SignData(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data with TPM: %w", err)
	}

	klog.Infof("TPM signing took %v", time.Since(beforeSign))

	return signature, nil
}

func debugToPEM(key crypto.PublicKey) string {
	var b bytes.Buffer
	pkData, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Sprintf("{MarshalPKIXPublicKey failed: %v}", err)
	}
	if err := pem.Encode(&b, &pem.Block{Type: "PUBLIC KEY", Bytes: pkData}); err != nil {
		return fmt.Sprintf("{pem.Encode failed: %v}", err)
	}
	return b.String()
}
