/*
Copyright 2023 The Kubernetes Authors.

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

// AuthToken describes the authentication header data when using GCE TPM authentication.
type AuthToken struct {
	// Signature is the TPM or PKI signature for data
	Signature []byte `json:"signature,omitempty"`

	// Data is the data we are signing.
	// It is a JSON encoded form of AuthTokenData.
	Data []byte `json:"data,omitempty"`
}

// AudienceNodeAuthentication is used in case we have multiple audiences using the TPM in future
const AudienceNodeAuthentication = "kops.k8s.io/node-bootstrap"
