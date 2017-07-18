/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KeysetType describes the type of keys in a KeySet
type KeysetType string

const (
	// TODO: Move CA to use these values
	SecretTypeSSHPublicKey KeysetType = "SSHPublicKey"
	SecretTypeKeypair      KeysetType = "Keypair"
	SecretTypeSecret       KeysetType = "Secret"

	// Name for the primary SSH key
	SecretNameSSHPrimary = "admin"
)

// +genclient=true

type Keyset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KeysetSpec `json:"spec,omitempty"`
}

type KeysetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Keyset `json:"items"`
}

type KeyItem struct {
	// Id is the unique identifier for this key in the keyset
	Id string `json:"id,omitempty"`

	// PublicMaterial holds non-secret material (e.g. a certificate, or SSH public key)
	PublicMaterial []byte `json:"publicMaterial,omitempty"`

	// PrivateMaterial holds secret material (e.g. a private key, SSH private key, or symmetric token)
	PrivateMaterial []byte `json:"privateMaterial,omitempty"`
}

type KeysetSpec struct {
	// Type is the type of the Key
	Type KeysetType `json:"channel,omitempty"`

	// Keys is the set of keys that make up the keyset
	Keys []KeyItem `json:"keys,omitempty"`
}
