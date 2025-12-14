/*
Copyright 2025 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var DiscoveryEndpointGVR = schema.GroupVersionResource{
	Group:    "discovery.kops.k8s.io",
	Version:  "v1alpha1",
	Resource: "discoveryendpoints",
}

var DiscoveryEndpointGVK = schema.GroupVersionKind{
	Group:   "discovery.kops.k8s.io",
	Version: "v1alpha1",
	Kind:    "DiscoveryEndpoint",
}

// DiscoveryEndpoint represents a registered client in the discovery service.
type DiscoveryEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DiscoveryEndpointSpec `json:"spec,omitempty"`
}

// DiscoveryEndpointSpec corresponds to our internal Node data.
type DiscoveryEndpointSpec struct {
	Addresses []string  `json:"addresses,omitempty"`
	LastSeen  string    `json:"lastSeen,omitempty"`
	OIDC      *OIDCSpec `json:"oidc,omitempty"`
}

type OIDCSpec struct {
	// IssuerURL string       `json:"issuerURL,omitempty"`
	Keys []JSONWebKey `json:"keys,omitempty"`
}

type JSONWebKey struct {
	Use       string `json:"use,omitempty"`
	KeyType   string `json:"kty,omitempty"`
	KeyID     string `json:"kid,omitempty"`
	Algorithm string `json:"alg,omitempty"`
	N         string `json:"n,omitempty"`
	E         string `json:"e,omitempty"`
	// Crv       string `json:"crv,omitempty"`
	// X         string `json:"x,omitempty"`
	// Y         string `json:"y,omitempty"`
}

// DiscoveryEndpointList is a list of DiscoveryEndpoint objects.
type DiscoveryEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// We implement a minimal subset.
	Metadata metav1.ListMeta     `json:"metadata,omitempty"`
	Items    []DiscoveryEndpoint `json:"items"`
}
