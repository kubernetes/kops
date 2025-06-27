/*
Copyright 2020 The cert-manager Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// Order is a type to represent an Order with an ACME server
// +k8s:openapi-gen=true
type Order struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec OrderSpec `json:"spec"`
	// +optional
	Status OrderStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OrderList is a list of Orders
type OrderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Order `json:"items"`
}

type OrderSpec struct {
	// Certificate signing request bytes in DER encoding.
	// This will be used when finalizing the order.
	// This field must be set on the order.
	Request []byte `json:"request"`

	// IssuerRef references a properly configured ACME-type Issuer which should
	// be used to create this Order.
	// If the Issuer does not exist, processing will be retried.
	// If the Issuer is not an 'ACME' Issuer, an error will be returned and the
	// Order will be marked as failed.
	IssuerRef cmmeta.ObjectReference `json:"issuerRef"`

	// CommonName is the common name as specified on the DER encoded CSR.
	// If specified, this value must also be present in `dnsNames` or `ipAddresses`.
	// This field must match the corresponding field on the DER encoded CSR.
	// +optional
	CommonName string `json:"commonName,omitempty"`

	// DNSNames is a list of DNS names that should be included as part of the Order
	// validation process.
	// This field must match the corresponding field on the DER encoded CSR.
	//+optional
	DNSNames []string `json:"dnsNames,omitempty"`

	// IPAddresses is a list of IP addresses that should be included as part of the Order
	// validation process.
	// This field must match the corresponding field on the DER encoded CSR.
	// +optional
	IPAddresses []string `json:"ipAddresses,omitempty"`

	// Duration is the duration for the not after date for the requested certificate.
	// this is set on order creation as pe the ACME spec.
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// Profile allows requesting a certificate profile from the ACME server.
	// Supported profiles are listed by the server's ACME directory URL.
	// +optional
	Profile string `json:"profile,omitempty"`
}

type OrderStatus struct {
	// URL of the Order.
	// This will initially be empty when the resource is first created.
	// The Order controller will populate this field when the Order is first processed.
	// This field will be immutable after it is initially set.
	// +optional
	URL string `json:"url,omitempty"`

	// FinalizeURL of the Order.
	// This is used to obtain certificates for this order once it has been completed.
	// +optional
	FinalizeURL string `json:"finalizeURL,omitempty"`

	// Authorizations contains data returned from the ACME server on what
	// authorizations must be completed in order to validate the DNS names
	// specified on the Order.
	// +optional
	Authorizations []ACMEAuthorization `json:"authorizations,omitempty"`

	// Certificate is a copy of the PEM encoded certificate for this Order.
	// This field will be populated after the order has been successfully
	// finalized with the ACME server, and the order has transitioned to the
	// 'valid' state.
	// +optional
	Certificate []byte `json:"certificate,omitempty"`

	// State contains the current state of this Order resource.
	// States 'success' and 'expired' are 'final'
	// +optional
	State State `json:"state,omitempty"`

	// Reason optionally provides more information about a why the order is in
	// the current state.
	// +optional
	Reason string `json:"reason,omitempty"`

	// FailureTime stores the time that this order failed.
	// This is used to influence garbage collection and back-off.
	// +optional
	FailureTime *metav1.Time `json:"failureTime,omitempty"`
}

// ACMEAuthorization contains data returned from the ACME server on an
// authorization that must be completed in order validate a DNS name on an ACME
// Order resource.
type ACMEAuthorization struct {
	// URL is the URL of the Authorization that must be completed
	URL string `json:"url"`

	// Identifier is the DNS name to be validated as part of this authorization
	// +optional
	Identifier string `json:"identifier,omitempty"`

	// Wildcard will be true if this authorization is for a wildcard DNS name.
	// If this is true, the identifier will be the *non-wildcard* version of
	// the DNS name.
	// For example, if '*.example.com' is the DNS name being validated, this
	// field will be 'true' and the 'identifier' field will be 'example.com'.
	// +optional
	Wildcard *bool `json:"wildcard,omitempty"`

	// InitialState is the initial state of the ACME authorization when first
	// fetched from the ACME server.
	// If an Authorization is already 'valid', the Order controller will not
	// create a Challenge resource for the authorization. This will occur when
	// working with an ACME server that enables 'authz reuse' (such as Let's
	// Encrypt's production endpoint).
	// If not set and 'identifier' is set, the state is assumed to be pending
	// and a Challenge will be created.
	// +optional
	InitialState State `json:"initialState,omitempty"`

	// Challenges specifies the challenge types offered by the ACME server.
	// One of these challenge types will be selected when validating the DNS
	// name and an appropriate Challenge resource will be created to perform
	// the ACME challenge process.
	// +optional
	Challenges []ACMEChallenge `json:"challenges,omitempty"`
}

// Challenge specifies a challenge offered by the ACME server for an Order.
// An appropriate Challenge resource can be created to perform the ACME
// challenge process.
type ACMEChallenge struct {
	// URL is the URL of this challenge. It can be used to retrieve additional
	// metadata about the Challenge from the ACME server.
	URL string `json:"url"`

	// Token is the token that must be presented for this challenge.
	// This is used to compute the 'key' that must also be presented.
	Token string `json:"token"`

	// Type is the type of challenge being offered, e.g., 'http-01', 'dns-01',
	// 'tls-sni-01', etc.
	// This is the raw value retrieved from the ACME server.
	// Only 'http-01' and 'dns-01' are supported by cert-manager, other values
	// will be ignored.
	Type string `json:"type"`
}

// State represents the state of an ACME resource, such as an Order.
// The possible options here map to the corresponding values in the
// ACME specification.
// Full details of these values can be found here: https://tools.ietf.org/html/draft-ietf-acme-acme-15#section-7.1.6
// Clients utilising this type must also gracefully handle unknown
// values, as the contents of this enumeration may be added to over time.
// +kubebuilder:validation:Enum=valid;ready;pending;processing;invalid;expired;errored
type State string

const (
	// Unknown is not a real state as part of the ACME spec.
	// It is used to represent an unrecognised value.
	Unknown State = ""

	// Valid signifies that an ACME resource is in a valid state.
	// If an order is 'valid', it has been finalized with the ACME server and
	// the certificate can be retrieved from the ACME server using the
	// certificate URL stored in the Order's status subresource.
	// This is a final state.
	Valid State = "valid"

	// Ready signifies that an ACME resource is in a ready state.
	// If an order is 'ready', all of its challenges have been completed
	// successfully and the order is ready to be finalized.
	// Once finalized, it will transition to the Valid state.
	// This is a transient state.
	Ready State = "ready"

	// Pending signifies that an ACME resource is still pending and is not yet ready.
	// If an Order is marked 'Pending', the validations for that Order are still in progress.
	// This is a transient state.
	Pending State = "pending"

	// Processing signifies that an ACME resource is being processed by the server.
	// If an Order is marked 'Processing', the validations for that Order are currently being processed.
	// This is a transient state.
	Processing State = "processing"

	// Invalid signifies that an ACME resource is invalid for some reason.
	// If an Order is marked 'invalid', one of its validations must be invalid for some reason.
	// This is a final state.
	Invalid State = "invalid"

	// Expired signifies that an ACME resource has expired.
	// If an Order is marked 'Expired', one of its validations may have expired or the Order itself.
	// This is a final state.
	Expired State = "expired"

	// Errored signifies that the ACME resource has errored for some reason.
	// This is a catch-all state, and is used for marking internal cert-manager
	// errors such as validation failures.
	// This is a final state.
	Errored State = "errored"
)
