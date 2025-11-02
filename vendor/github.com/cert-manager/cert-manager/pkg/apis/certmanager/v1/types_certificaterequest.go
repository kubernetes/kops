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

const (
	// Pending indicates that a CertificateRequest is still in progress.
	CertificateRequestReasonPending = "Pending"

	// Failed indicates that a CertificateRequest has failed permanently,
	// either due to timing out or some other critical failure.
	// The `status.failureTime` field should be set in this case.
	CertificateRequestReasonFailed = "Failed"

	// Issued indicates that a CertificateRequest has been completed, and that
	// the `status.certificate` field is set.
	CertificateRequestReasonIssued = "Issued"

	// Denied is a Ready condition reason that indicates that a
	// CertificateRequest has been denied, and the CertificateRequest will never
	// be issued.
	// The `status.failureTime` field should be set in this case.
	CertificateRequestReasonDenied = "Denied"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Approved",type="string",JSONPath=`.status.conditions[?(@.type == "Approved")].status`
// +kubebuilder:printcolumn:name="Denied",type="string",JSONPath=`.status.conditions[?(@.type == "Denied")].status`
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type == "Ready")].status`
// +kubebuilder:printcolumn:name="Issuer",type="string",JSONPath=`.spec.issuerRef.name`
// +kubebuilder:printcolumn:name="Requester",type="string",JSONPath=`.spec.username`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.conditions[?(@.type == "Ready")].message`,priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:resource:scope=Namespaced,shortName={cr,crs},categories=cert-manager
// +kubebuilder:subresource:status

// A CertificateRequest is used to request a signed certificate from one of the
// configured issuers.
//
// All fields within the CertificateRequest's `spec` are immutable after creation.
// A CertificateRequest will either succeed or fail, as denoted by its `Ready` status
// condition and its `status.failureTime` field.
//
// A CertificateRequest is a one-shot resource, meaning it represents a single
// point in time request for a certificate and cannot be re-used.
type CertificateRequest struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired state of the CertificateRequest resource.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec CertificateRequestSpec `json:"spec"`

	// Status of the CertificateRequest.
	// This is set and managed automatically.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status CertificateRequestStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateRequestList is a list of CertificateRequests.
type CertificateRequestList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of CertificateRequests
	Items []CertificateRequest `json:"items"`
}

// CertificateRequestSpec defines the desired state of CertificateRequest
//
// NOTE: It is important to note that the issuer can choose to ignore or change
// any of the requested attributes. How the issuer maps a certificate request
// to a signed certificate is the full responsibility of the issuer itself.
// For example, as an edge case, an issuer that inverts the isCA value is
// free to do so.
type CertificateRequestSpec struct {
	// Requested 'duration' (i.e. lifetime) of the Certificate. Note that the
	// issuer may choose to ignore the requested duration, just like any other
	// requested attribute.
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// Reference to the issuer responsible for issuing the certificate.
	// If the issuer is namespace-scoped, it must be in the same namespace
	// as the Certificate. If the issuer is cluster-scoped, it can be used
	// from any namespace.
	//
	// The `name` field of the reference must always be specified.
	IssuerRef cmmeta.IssuerReference `json:"issuerRef"`

	// The PEM-encoded X.509 certificate signing request to be submitted to the
	// issuer for signing.
	//
	// If the CSR has a BasicConstraints extension, its isCA attribute must
	// match the `isCA` value of this CertificateRequest.
	// If the CSR has a KeyUsage extension, its key usages must match the
	// key usages in the `usages` field of this CertificateRequest.
	// If the CSR has a ExtKeyUsage extension, its extended key usages
	// must match the extended key usages in the `usages` field of this
	// CertificateRequest.
	Request []byte `json:"request"`

	// Requested basic constraints isCA value. Note that the issuer may choose
	// to ignore the requested isCA value, just like any other requested attribute.
	//
	// NOTE: If the CSR in the `Request` field has a BasicConstraints extension,
	// it must have the same isCA value as specified here.
	//
	// If true, this will automatically add the `cert sign` usage to the list
	// of requested `usages`.
	// +optional
	IsCA bool `json:"isCA,omitempty"`

	// Requested key usages and extended key usages.
	//
	// NOTE: If the CSR in the `Request` field has uses the KeyUsage or
	// ExtKeyUsage extension, these extensions must have the same values
	// as specified here without any additional values.
	//
	// If unset, defaults to `digital signature` and `key encipherment`.
	// +optional
	// +listType=atomic
	Usages []KeyUsage `json:"usages,omitempty"`

	// Username contains the name of the user that created the CertificateRequest.
	// Populated by the cert-manager webhook on creation and immutable.
	// +optional
	Username string `json:"username,omitempty"`
	// UID contains the uid of the user that created the CertificateRequest.
	// Populated by the cert-manager webhook on creation and immutable.
	// +optional
	UID string `json:"uid,omitempty"`
	// Groups contains group membership of the user that created the CertificateRequest.
	// Populated by the cert-manager webhook on creation and immutable.
	// +optional
	// +listType=atomic
	Groups []string `json:"groups,omitempty"`
	// Extra contains extra attributes of the user that created the CertificateRequest.
	// Populated by the cert-manager webhook on creation and immutable.
	// +optional
	Extra map[string][]string `json:"extra,omitempty"`
}

// CertificateRequestStatus defines the observed state of CertificateRequest and
// resulting signed certificate.
type CertificateRequestStatus struct {
	// List of status conditions to indicate the status of a CertificateRequest.
	// Known condition types are `Ready`, `InvalidRequest`, `Approved` and `Denied`.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []CertificateRequestCondition `json:"conditions,omitempty"`

	// The PEM encoded X.509 certificate resulting from the certificate
	// signing request.
	// If not set, the CertificateRequest has either not been completed or has
	// failed. More information on failure can be found by checking the
	// `conditions` field.
	// +optional
	Certificate []byte `json:"certificate,omitempty"`

	// The PEM encoded X.509 certificate of the signer, also known as the CA
	// (Certificate Authority).
	// This is set on a best-effort basis by different issuers.
	// If not set, the CA is assumed to be unknown/not available.
	// +optional
	CA []byte `json:"ca,omitempty"`

	// FailureTime stores the time that this CertificateRequest failed. This is
	// used to influence garbage collection and back-off.
	// +optional
	FailureTime *metav1.Time `json:"failureTime,omitempty"`
}

// CertificateRequestCondition contains condition information for a CertificateRequest.
type CertificateRequestCondition struct {
	// Type of the condition, known values are (`Ready`, `InvalidRequest`,
	// `Approved`, `Denied`).
	Type CertificateRequestConditionType `json:"type"`

	// Status of the condition, one of (`True`, `False`, `Unknown`).
	Status cmmeta.ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}

// CertificateRequestConditionType represents a Certificate condition value.
type CertificateRequestConditionType string

const (
	// CertificateRequestConditionReady indicates that a certificate is ready for use.
	// This is defined as:
	// - The target certificate exists in CertificateRequest.Status
	CertificateRequestConditionReady CertificateRequestConditionType = "Ready"

	// CertificateRequestConditionInvalidRequest indicates that a certificate
	// signer has refused to sign the request due to at least one of the input
	// parameters being invalid. Additional information about why the request
	// was rejected can be found in the `reason` and `message` fields.
	CertificateRequestConditionInvalidRequest CertificateRequestConditionType = "InvalidRequest"

	// CertificateRequestConditionApproved indicates that a certificate request
	// is approved and ready for signing. Condition must never have a status of
	// `False`, and cannot be modified once set. Cannot be set alongside
	// `Denied`.
	CertificateRequestConditionApproved CertificateRequestConditionType = "Approved"

	// CertificateRequestConditionDenied indicates that a certificate request is
	// denied, and must never be signed. Condition must never have a status of
	// `False`, and cannot be modified once set. Cannot be set alongside
	// `Approved`.
	CertificateRequestConditionDenied CertificateRequestConditionType = "Denied"
)
