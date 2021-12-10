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

package gcetpm

// AuthToken describes the authentication header data when using GCE TPM authentication.
type AuthToken struct {
	// Signature is the TPM signature for data
	Signature []byte `json:"signature,omitempty"`

	// Data is the data we are signing.
	// It is a JSON encoded form of AuthTokenData.
	Data []byte `json:"data,omitempty"`
}

// AuthTokenData is the code data that is signed as part of the header.
type AuthTokenData struct {
	// GCPProjectID is the GCP project we claim to be part of
	GCPProjectID string `json:"gcpProjectID,omitempty"`

	// Zone is the availability zone we claim to be part of
	Zone string `json:"zone,omitempty"`

	// Instance is the name/id of the instance we are claiming
	Instance string `json:"instance,omitempty"`

	// KeyID is the identifier of the public key we are signing with, if we're using a fixed key.
	KeyID string `json:"keyID,omitempty"`

	// RequestHash is the hash of the request
	RequestHash []byte `json:"requestHash,omitempty"`

	// Timestamp is the time of this request (to help prevent replay attacks)
	Timestamp int64 `json:"timestamp,omitempty"`

	// Audience is the audience for this request (to help prevent replay attacks)
	Audience string `json:"audience,omitempty"`
}

// GCETPMAuthenticationTokenPrefix is the prefix used for authentication using the GCE TPM
const GCETPMAuthenticationTokenPrefix = "x-gce-tpm "

// AudienceNodeAuthentication is used in case we have multiple audiences using the TPM in future
const AudienceNodeAuthentication = "kops.k8s.io/node-bootstrap"
