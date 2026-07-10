/*
Copyright 2026 The Kubernetes Authors.

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

package azuremetadata

import (
	"crypto/sha256"
	"encoding/hex"
)

// NonceLength is the length of the hex-encoded SHA256 prefix used as the Azure
// IMDS attested document nonce. IMDS enforces a 32-character maximum for the nonce parameter; 32
// hex chars is 128 bits of entropy, well above the cryptographic nonce floor.
const NonceLength = 32

// NonceForBody derives the IMDS attestation nonce from the request body. It is used by both the
// authenticator (nodeup) and verifier (kops-controller) sides.
func NonceForBody(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])[:NonceLength]
}
