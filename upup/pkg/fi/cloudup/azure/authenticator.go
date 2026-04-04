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

package azure

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
)

// AzureAuthenticationTokenPrefix prefixes bootstrap tokens created from Azure
// IMDS instance identity data.
const AzureAuthenticationTokenPrefix = "x-azure-id "

type azureAuthenticator struct{}

var _ bootstrap.Authenticator = (*azureAuthenticator)(nil)

// NewAzureAuthenticator returns an authenticator that mints Azure bootstrap
// tokens backed by IMDS metadata and an attested document signature.
func NewAzureAuthenticator() (bootstrap.Authenticator, error) {
	return &azureAuthenticator{}, nil
}

// CreateToken fetches the local VM identity from IMDS and returns a bootstrap
// token containing the resource ID and signed attested document.
func (h *azureAuthenticator) CreateToken(body []byte) (string, error) {
	klog.V(4).Infof("Azure authenticator creating bootstrap token")

	// Query IMDS for the VM's resource ID.
	metadata, err := QueryComputeInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("querying instance metadata: %w", err)
	}
	if metadata == nil || metadata.ResourceID == "" {
		return "", fmt.Errorf("missing resource ID")
	}
	klog.V(4).Infof("Azure authenticator obtained resource ID %q", metadata.ResourceID)

	// Query IMDS for a PKCS7-signed attested document containing the nonce.
	nonce := nonceForBody(body)
	doc, err := queryIMDSAttestedDocument(nonce)
	if err != nil {
		return "", fmt.Errorf("querying attested document: %w", err)
	}
	if doc.Signature == "" {
		return "", fmt.Errorf("empty attested document signature")
	}
	klog.V(2).Infof("Azure authenticator obtained attested document for %q", metadata.ResourceID)

	// Token format: "x-azure-id <resourceID> <base64-pkcs7-signature>"
	return AzureAuthenticationTokenPrefix + metadata.ResourceID + " " + doc.Signature, nil
}
