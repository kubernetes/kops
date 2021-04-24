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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
)

const (
	// Pending indicates that a CertificateRequest is still in progress.
	CertificateRequestReasonPending = "Pending"

	// Failed indicates that a CertificateRequest has failed, either due to
	// timing out or some other critical failure.
	CertificateRequestReasonFailed = "Failed"

	// Issued indicates that a CertificateRequest has been completed, and that
	// the `status.certificate` field is set.
	CertificateRequestReasonIssued = "Issued"

	// Denied is a Ready condition reason that indicates that a
	// CertificateRequest has been denied, and the CertificateRequest will never
	// be issued.
	CertificateRequestReasonDenied = "Denied"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// A CertificateRequest is used to request a signed certificate from one of the
// configured issuers.
//
// All fields within the CertificateRequest's `spec` are immutable after creation.
// A CertificateRequest will either succeed or fail, as denoted by its `status.state`
// field.
//
// A CertificateRequest is a one-shot resource, meaning it represents a single
// point in time request for a certificate and cannot be re-used.
// +k8s:openapi-gen=true
type CertificateRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Desired state of the CertificateRequest resource.
	Spec CertificateRequestSpec `json:"spec,omitempty"`

	// Status of the CertificateRequest. This is set and managed automatically.
	Status CertificateRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateRequestList is a list of Certificates
type CertificateRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CertificateRequest `json:"items"`
}

// CertificateRequestSpec defines the desired state of CertificateRequest
type CertificateRequestSpec struct {
	// The requested 'duration' (i.e. lifetime) of the Certificate.
	// This option may be ignored/overridden by some issuer types.
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// IssuerRef is a reference to the issuer for this CertificateRequest.  If
	// the `kind` field is not set, or set to `Issuer`, an Issuer resource with
	// the given name in the same namespace as the CertificateRequest will be
	// used.  If the `kind` field is set to `ClusterIssuer`, a ClusterIssuer with
	// the provided name will be used. The `name` field in this stanza is
	// required at all times. The group field refers to the API group of the
	// issuer which defaults to `cert-manager.io` if empty.
	IssuerRef cmmeta.ObjectReference `json:"issuerRef"`

	// The PEM-encoded x509 certificate signing request to be submitted to the
	// CA for signing.
	CSRPEM []byte `json:"csr"`

	// IsCA will request to mark the certificate as valid for certificate signing
	// when submitting to the issuer.
	// This will automatically add the `cert sign` usage to the list of `usages`.
	// +optional
	IsCA bool `json:"isCA,omitempty"`

	// Usages is the set of x509 usages that are requested for the certificate.
	// Defaults to `digital signature` and `key encipherment` if not specified.
	// +optional
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
	// +listType=atomic
	// +optional
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
	// Known condition types are `Ready` and `InvalidRequest`.
	// +optional
	Conditions []CertificateRequestCondition `json:"conditions,omitempty"`

	// The PEM encoded x509 certificate resulting from the certificate
	// signing request.
	// If not set, the CertificateRequest has either not been completed or has
	// failed. More information on failure can be found by checking the
	// `conditions` field.
	// +optional
	Certificate []byte `json:"certificate,omitempty"`

	// The PEM encoded x509 certificate of the signer, also known as the CA
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
	// Type of the condition, known values are (`Ready`,
	// `InvalidRequest`, `Approved`, `Denied`).
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

// CertificateRequestConditionType represents an Certificate condition value.
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
