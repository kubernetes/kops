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

// NOTE: Be mindful of adding OpenAPI validation- see https://github.com/cert-manager/cert-manager/issues/3644

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// A Certificate resource should be created to ensure an up to date and signed
// X.509 certificate is stored in the Kubernetes Secret resource named in `spec.secretName`.
//
// The stored certificate will be renewed before it expires (as configured by `spec.renewBefore`).
type Certificate struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired state of the Certificate resource.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec CertificateSpec `json:"spec"`

	// Status of the Certificate.
	// This is set and managed automatically.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status CertificateStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateList is a list of Certificates.
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of Certificates
	Items []Certificate `json:"items"`
}

// +kubebuilder:validation:Enum=RSA;ECDSA;Ed25519
type PrivateKeyAlgorithm string

const (
	// RSA private key algorithm.
	RSAKeyAlgorithm PrivateKeyAlgorithm = "RSA"

	// ECDSA private key algorithm.
	ECDSAKeyAlgorithm PrivateKeyAlgorithm = "ECDSA"

	// Ed25519 private key algorithm.
	Ed25519KeyAlgorithm PrivateKeyAlgorithm = "Ed25519"
)

// +kubebuilder:validation:Enum=PKCS1;PKCS8
type PrivateKeyEncoding string

const (
	// PKCS1 private key encoding.
	// PKCS1 produces a PEM block that contains the private key algorithm
	// in the header and the private key in the body. A key that uses this
	// can be recognised by its `BEGIN RSA PRIVATE KEY` or `BEGIN EC PRIVATE KEY` header.
	// NOTE: This encoding is not supported for Ed25519 keys. Attempting to use
	// this encoding with an Ed25519 key will be ignored and default to PKCS8.
	PKCS1 PrivateKeyEncoding = "PKCS1"

	// PKCS8 private key encoding.
	// PKCS8 produces a PEM block with a static header and both the private
	// key algorithm and the private key in the body. A key that uses this
	// encoding can be recognised by its `BEGIN PRIVATE KEY` header.
	PKCS8 PrivateKeyEncoding = "PKCS8"
)

// CertificateSpec defines the desired state of Certificate.
//
// NOTE: The specification contains a lot of "requested" certificate attributes, it is
// important to note that the issuer can choose to ignore or change any of
// these requested attributes. How the issuer maps a certificate request to a
// signed certificate is the full responsibility of the issuer itself. For example,
// as an edge case, an issuer that inverts the isCA value is free to do so.
//
// A valid Certificate requires at least one of a CommonName, LiteralSubject, DNSName, or
// URI to be valid.
type CertificateSpec struct {
	// Requested set of X509 certificate subject attributes.
	// More info: https://datatracker.ietf.org/doc/html/rfc5280#section-4.1.2.6
	//
	// The common name attribute is specified separately in the `commonName` field.
	// Cannot be set if the `literalSubject` field is set.
	// +optional
	Subject *X509Subject `json:"subject,omitempty"`

	// Requested X.509 certificate subject, represented using the LDAP "String
	// Representation of a Distinguished Name" [1].
	// Important: the LDAP string format also specifies the order of the attributes
	// in the subject, this is important when issuing certs for LDAP authentication.
	// Example: `CN=foo,DC=corp,DC=example,DC=com`
	// More info [1]: https://datatracker.ietf.org/doc/html/rfc4514
	// More info: https://github.com/cert-manager/cert-manager/issues/3203
	// More info: https://github.com/cert-manager/cert-manager/issues/4424
	//
	// Cannot be set if the `subject` or `commonName` field is set.
	// +optional
	LiteralSubject string `json:"literalSubject,omitempty"`

	// Requested common name X509 certificate subject attribute.
	// More info: https://datatracker.ietf.org/doc/html/rfc5280#section-4.1.2.6
	// NOTE: TLS clients will ignore this value when any subject alternative name is
	// set (see https://tools.ietf.org/html/rfc6125#section-6.4.4).
	//
	// Should have a length of 64 characters or fewer to avoid generating invalid CSRs.
	// Cannot be set if the `literalSubject` field is set.
	// +optional
	CommonName string `json:"commonName,omitempty"`

	// Requested 'duration' (i.e. lifetime) of the Certificate. Note that the
	// issuer may choose to ignore the requested duration, just like any other
	// requested attribute.
	//
	// If unset, this defaults to 90 days.
	// Minimum accepted duration is 1 hour.
	// Value must be in units accepted by Go time.ParseDuration https://golang.org/pkg/time/#ParseDuration.
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// How long before the currently issued certificate's expiry cert-manager should
	// renew the certificate. For example, if a certificate is valid for 60 minutes,
	// and `renewBefore=10m`, cert-manager will begin to attempt to renew the certificate
	// 50 minutes after it was issued (i.e. when there are 10 minutes remaining until
	// the certificate is no longer valid).
	//
	// NOTE: The actual lifetime of the issued certificate is used to determine the
	// renewal time. If an issuer returns a certificate with a different lifetime than
	// the one requested, cert-manager will use the lifetime of the issued certificate.
	//
	// If unset, this defaults to 1/3 of the issued certificate's lifetime.
	// Minimum accepted value is 5 minutes.
	// Value must be in units accepted by Go time.ParseDuration https://golang.org/pkg/time/#ParseDuration.
	// +optional
	RenewBefore *metav1.Duration `json:"renewBefore,omitempty"`

	// Requested DNS subject alternative names.
	// +optional
	DNSNames []string `json:"dnsNames,omitempty"`

	// Requested IP address subject alternative names.
	// +optional
	IPAddresses []string `json:"ipAddresses,omitempty"`

	// Requested URI subject alternative names.
	// +optional
	URIs []string `json:"uris,omitempty"`

	// `otherNames` is an escape hatch for SAN that allows any type. We currently restrict the support to string like otherNames, cf RFC 5280 p 37
	// Any UTF8 String valued otherName can be passed with by setting the keys oid: x.x.x.x and UTF8Value: somevalue for `otherName`.
	// Most commonly this would be UPN set with oid: 1.3.6.1.4.1.311.20.2.3
	// You should ensure that any OID passed is valid for the UTF8String type as we do not explicitly validate this.
	// +optional
	OtherNames []OtherName `json:"otherNames,omitempty"`

	// Requested email subject alternative names.
	// +optional
	EmailAddresses []string `json:"emailAddresses,omitempty"`

	// Name of the Secret resource that will be automatically created and
	// managed by this Certificate resource. It will be populated with a
	// private key and certificate, signed by the denoted issuer. The Secret
	// resource lives in the same namespace as the Certificate resource.
	SecretName string `json:"secretName"`

	// Defines annotations and labels to be copied to the Certificate's Secret.
	// Labels and annotations on the Secret will be changed as they appear on the
	// SecretTemplate when added or removed. SecretTemplate annotations are added
	// in conjunction with, and cannot overwrite, the base set of annotations
	// cert-manager sets on the Certificate's Secret.
	// +optional
	SecretTemplate *CertificateSecretTemplate `json:"secretTemplate,omitempty"`

	// Additional keystore output formats to be stored in the Certificate's Secret.
	// +optional
	Keystores *CertificateKeystores `json:"keystores,omitempty"`

	// Reference to the issuer responsible for issuing the certificate.
	// If the issuer is namespace-scoped, it must be in the same namespace
	// as the Certificate. If the issuer is cluster-scoped, it can be used
	// from any namespace.
	//
	// The `name` field of the reference must always be specified.
	IssuerRef cmmeta.ObjectReference `json:"issuerRef"`

	// Requested basic constraints isCA value.
	// The isCA value is used to set the `isCA` field on the created CertificateRequest
	// resources. Note that the issuer may choose to ignore the requested isCA value, just
	// like any other requested attribute.
	//
	// If true, this will automatically add the `cert sign` usage to the list
	// of requested `usages`.
	// +optional
	IsCA bool `json:"isCA,omitempty"`

	// Requested key usages and extended key usages.
	// These usages are used to set the `usages` field on the created CertificateRequest
	// resources. If `encodeUsagesInRequest` is unset or set to `true`, the usages
	// will additionally be encoded in the `request` field which contains the CSR blob.
	//
	// If unset, defaults to `digital signature` and `key encipherment`.
	// +optional
	Usages []KeyUsage `json:"usages,omitempty"`

	// Private key options. These include the key algorithm and size, the used
	// encoding and the rotation policy.
	// +optional
	PrivateKey *CertificatePrivateKey `json:"privateKey,omitempty"`

	// Whether the KeyUsage and ExtKeyUsage extensions should be set in the encoded CSR.
	//
	// This option defaults to true, and should only be disabled if the target
	// issuer does not support CSRs with these X509 KeyUsage/ ExtKeyUsage extensions.
	// +optional
	EncodeUsagesInRequest *bool `json:"encodeUsagesInRequest,omitempty"`

	// The maximum number of CertificateRequest revisions that are maintained in
	// the Certificate's history. Each revision represents a single `CertificateRequest`
	// created by this Certificate, either when it was created, renewed, or Spec
	// was changed. Revisions will be removed by oldest first if the number of
	// revisions exceeds this number.
	//
	// If set, revisionHistoryLimit must be a value of `1` or greater.
	// If unset (`nil`), revisions will not be garbage collected.
	// Default value is `nil`.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`

	// Defines extra output formats of the private key and signed certificate chain
	// to be written to this Certificate's target Secret.
	//
	// This is a Beta Feature enabled by default. It can be disabled with the
	// `--feature-gates=AdditionalCertificateOutputFormats=false` option set on both
	// the controller and webhook components.
	// +optional
	AdditionalOutputFormats []CertificateAdditionalOutputFormat `json:"additionalOutputFormats,omitempty"`

	// x.509 certificate NameConstraint extension which MUST NOT be used in a non-CA certificate.
	// More Info: https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.10
	//
	// This is an Alpha Feature and is only enabled with the
	// `--feature-gates=NameConstraints=true` option set on both
	// the controller and webhook components.
	// +optional
	NameConstraints *NameConstraints `json:"nameConstraints,omitempty"`
}

type OtherName struct {
	// OID is the object identifier for the otherName SAN.
	// The object identifier must be expressed as a dotted string, for
	// example, "1.2.840.113556.1.4.221".
	OID string `json:"oid,omitempty"`

	// utf8Value is the string value of the otherName SAN.
	// The utf8Value accepts any valid UTF8 string to set as value for the otherName SAN.
	UTF8Value string `json:"utf8Value,omitempty"`
}

// CertificatePrivateKey contains configuration options for private keys
// used by the Certificate controller.
// These include the key algorithm and size, the used encoding and the
// rotation policy.
type CertificatePrivateKey struct {
	// RotationPolicy controls how private keys should be regenerated when a
	// re-issuance is being processed.
	//
	// If set to `Never`, a private key will only be generated if one does not
	// already exist in the target `spec.secretName`. If one does exists but it
	// does not have the correct algorithm or size, a warning will be raised
	// to await user intervention.
	// If set to `Always`, a private key matching the specified requirements
	// will be generated whenever a re-issuance occurs.
	// Default is `Never` for backward compatibility.
	// +optional
	RotationPolicy PrivateKeyRotationPolicy `json:"rotationPolicy,omitempty"`

	// The private key cryptography standards (PKCS) encoding for this
	// certificate's private key to be encoded in.
	//
	// If provided, allowed values are `PKCS1` and `PKCS8` standing for PKCS#1
	// and PKCS#8, respectively.
	// Defaults to `PKCS1` if not specified.
	// +optional
	Encoding PrivateKeyEncoding `json:"encoding,omitempty"`

	// Algorithm is the private key algorithm of the corresponding private key
	// for this certificate.
	//
	// If provided, allowed values are either `RSA`, `ECDSA` or `Ed25519`.
	// If `algorithm` is specified and `size` is not provided,
	// key size of 2048 will be used for `RSA` key algorithm and
	// key size of 256 will be used for `ECDSA` key algorithm.
	// key size is ignored when using the `Ed25519` key algorithm.
	// +optional
	Algorithm PrivateKeyAlgorithm `json:"algorithm,omitempty"`

	// Size is the key bit size of the corresponding private key for this certificate.
	//
	// If `algorithm` is set to `RSA`, valid values are `2048`, `4096` or `8192`,
	// and will default to `2048` if not specified.
	// If `algorithm` is set to `ECDSA`, valid values are `256`, `384` or `521`,
	// and will default to `256` if not specified.
	// If `algorithm` is set to `Ed25519`, Size is ignored.
	// No other values are allowed.
	// +optional
	Size int `json:"size,omitempty"`
}

// Denotes how private keys should be generated or sourced when a Certificate
// is being issued.
// +kubebuilder:validation:Enum=Never;Always
type PrivateKeyRotationPolicy string

var (
	// RotationPolicyNever means a private key will only be generated if one
	// does not already exist in the target `spec.secretName`.
	// If one does exists but it does not have the correct algorithm or size,
	// a warning will be raised to await user intervention.
	RotationPolicyNever PrivateKeyRotationPolicy = "Never"

	// RotationPolicyAlways means a private key matching the specified
	// requirements will be generated whenever a re-issuance occurs.
	RotationPolicyAlways PrivateKeyRotationPolicy = "Always"
)

// CertificateOutputFormatType specifies which additional output formats should
// be written to the Certificate's target Secret.
// Allowed values are `DER` or `CombinedPEM`.
// When Type is set to `DER` an additional entry `key.der` will be written to
// the Secret, containing the binary format of the private key.
// When Type is set to `CombinedPEM` an additional entry `tls-combined.pem`
// will be written to the Secret, containing the PEM formatted private key and
// signed certificate chain (tls.key + tls.crt concatenated).
// +kubebuilder:validation:Enum=DER;CombinedPEM
type CertificateOutputFormatType string

const (
	// CertificateOutputFormatDERKey is the name of the data entry in the Secret
	// resource used to store the DER formatted private key.
	CertificateOutputFormatDERKey string = "key.der"

	// CertificateOutputFormatDER  writes the Certificate's private key in DER
	// binary format to the `key.der` target Secret Data key.
	CertificateOutputFormatDER CertificateOutputFormatType = "DER"

	// CertificateOutputFormatCombinedPEMKey is the name of the data entry in the Secret
	// resource used to store the combined PEM (key + signed certificate).
	CertificateOutputFormatCombinedPEMKey string = "tls-combined.pem"

	// CertificateOutputFormatCombinedPEM  writes the Certificate's signed
	// certificate chain and private key, in PEM format, to the
	// `tls-combined.pem` target Secret Data key. The value at this key will
	// include the private key PEM document, followed by at least one new line
	// character, followed by the chain of signed certificate PEM documents
	// (`<private key> + \n + <signed certificate chain>`).
	CertificateOutputFormatCombinedPEM CertificateOutputFormatType = "CombinedPEM"
)

// CertificateAdditionalOutputFormat defines an additional output format of a
// Certificate resource. These contain supplementary data formats of the signed
// certificate chain and paired private key.
type CertificateAdditionalOutputFormat struct {
	// Type is the name of the format type that should be written to the
	// Certificate's target Secret.
	Type CertificateOutputFormatType `json:"type"`
}

// X509Subject Full X509 name specification
type X509Subject struct {
	// Organizations to be used on the Certificate.
	// +optional
	Organizations []string `json:"organizations,omitempty"`
	// Countries to be used on the Certificate.
	// +optional
	Countries []string `json:"countries,omitempty"`
	// Organizational Units to be used on the Certificate.
	// +optional
	OrganizationalUnits []string `json:"organizationalUnits,omitempty"`
	// Cities to be used on the Certificate.
	// +optional
	Localities []string `json:"localities,omitempty"`
	// State/Provinces to be used on the Certificate.
	// +optional
	Provinces []string `json:"provinces,omitempty"`
	// Street addresses to be used on the Certificate.
	// +optional
	StreetAddresses []string `json:"streetAddresses,omitempty"`
	// Postal codes to be used on the Certificate.
	// +optional
	PostalCodes []string `json:"postalCodes,omitempty"`
	// Serial number to be used on the Certificate.
	// +optional
	SerialNumber string `json:"serialNumber,omitempty"`
}

// CertificateKeystores configures additional keystore output formats to be
// created in the Certificate's output Secret.
type CertificateKeystores struct {
	// JKS configures options for storing a JKS keystore in the
	// `spec.secretName` Secret resource.
	// +optional
	JKS *JKSKeystore `json:"jks,omitempty"`

	// PKCS12 configures options for storing a PKCS12 keystore in the
	// `spec.secretName` Secret resource.
	// +optional
	PKCS12 *PKCS12Keystore `json:"pkcs12,omitempty"`
}

// JKS configures options for storing a JKS keystore in the `spec.secretName`
// Secret resource.
type JKSKeystore struct {
	// Create enables JKS keystore creation for the Certificate.
	// If true, a file named `keystore.jks` will be created in the target
	// Secret resource, encrypted using the password stored in
	// `passwordSecretRef`.
	// The keystore file will be updated immediately.
	// If the issuer provided a CA certificate, a file named `truststore.jks`
	// will also be created in the target Secret resource, encrypted using the
	// password stored in `passwordSecretRef`
	// containing the issuing Certificate Authority
	Create bool `json:"create"`

	// PasswordSecretRef is a reference to a key in a Secret resource
	// containing the password used to encrypt the JKS keystore.
	PasswordSecretRef cmmeta.SecretKeySelector `json:"passwordSecretRef"`

	// Alias specifies the alias of the key in the keystore, required by the JKS format.
	// If not provided, the default alias `certificate` will be used.
	// +optional
	Alias *string `json:"alias,omitempty"`
}

// PKCS12 configures options for storing a PKCS12 keystore in the
// `spec.secretName` Secret resource.
type PKCS12Keystore struct {
	// Create enables PKCS12 keystore creation for the Certificate.
	// If true, a file named `keystore.p12` will be created in the target
	// Secret resource, encrypted using the password stored in
	// `passwordSecretRef`.
	// The keystore file will be updated immediately.
	// If the issuer provided a CA certificate, a file named `truststore.p12` will
	// also be created in the target Secret resource, encrypted using the
	// password stored in `passwordSecretRef` containing the issuing Certificate
	// Authority
	Create bool `json:"create"`

	// PasswordSecretRef is a reference to a key in a Secret resource
	// containing the password used to encrypt the PKCS12 keystore.
	PasswordSecretRef cmmeta.SecretKeySelector `json:"passwordSecretRef"`

	// Profile specifies the key and certificate encryption algorithms and the HMAC algorithm
	// used to create the PKCS12 keystore. Default value is `LegacyRC2` for backward compatibility.
	//
	// If provided, allowed values are:
	// `LegacyRC2`: Deprecated. Not supported by default in OpenSSL 3 or Java 20.
	// `LegacyDES`: Less secure algorithm. Use this option for maximal compatibility.
	// `Modern2023`: Secure algorithm. Use this option in case you have to always use secure algorithms
	// (eg. because of company policy). Please note that the security of the algorithm is not that important
	// in reality, because the unencrypted certificate and private key are also stored in the Secret.
	// +optional
	Profile PKCS12Profile `json:"profile,omitempty"`
}

// +kubebuilder:validation:Enum=LegacyRC2;LegacyDES;Modern2023
type PKCS12Profile string

const (
	// see: https://pkg.go.dev/software.sslmate.com/src/go-pkcs12#LegacyRC2
	LegacyRC2PKCS12Profile PKCS12Profile = "LegacyRC2"

	// see: https://pkg.go.dev/software.sslmate.com/src/go-pkcs12#LegacyDES
	LegacyDESPKCS12Profile PKCS12Profile = "LegacyDES"

	// see: https://pkg.go.dev/software.sslmate.com/src/go-pkcs12#Modern2023
	Modern2023PKCS12Profile PKCS12Profile = "Modern2023"
)

// CertificateStatus defines the observed state of Certificate
type CertificateStatus struct {
	// List of status conditions to indicate the status of certificates.
	// Known condition types are `Ready` and `Issuing`.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []CertificateCondition `json:"conditions,omitempty"`

	// LastFailureTime is set only if the lastest issuance for this
	// Certificate failed and contains the time of the failure. If an
	// issuance has failed, the delay till the next issuance will be
	// calculated using formula time.Hour * 2 ^ (failedIssuanceAttempts -
	// 1). If the latest issuance has succeeded this field will be unset.
	// +optional
	LastFailureTime *metav1.Time `json:"lastFailureTime,omitempty"`

	// The time after which the certificate stored in the secret named
	// by this resource in `spec.secretName` is valid.
	// +optional
	NotBefore *metav1.Time `json:"notBefore,omitempty"`

	// The expiration time of the certificate stored in the secret named
	// by this resource in `spec.secretName`.
	// +optional
	NotAfter *metav1.Time `json:"notAfter,omitempty"`

	// RenewalTime is the time at which the certificate will be next
	// renewed.
	// If not set, no upcoming renewal is scheduled.
	// +optional
	RenewalTime *metav1.Time `json:"renewalTime,omitempty"`

	// The current 'revision' of the certificate as issued.
	//
	// When a CertificateRequest resource is created, it will have the
	// `cert-manager.io/certificate-revision` set to one greater than the
	// current value of this field.
	//
	// Upon issuance, this field will be set to the value of the annotation
	// on the CertificateRequest resource used to issue the certificate.
	//
	// Persisting the value on the CertificateRequest resource allows the
	// certificates controller to know whether a request is part of an old
	// issuance or if it is part of the ongoing revision's issuance by
	// checking if the revision value in the annotation is greater than this
	// field.
	// +optional
	Revision *int `json:"revision,omitempty"`

	// The name of the Secret resource containing the private key to be used
	// for the next certificate iteration.
	// The keymanager controller will automatically set this field if the
	// `Issuing` condition is set to `True`.
	// It will automatically unset this field when the Issuing condition is
	// not set or False.
	// +optional
	NextPrivateKeySecretName *string `json:"nextPrivateKeySecretName,omitempty"`

	// The number of continuous failed issuance attempts up till now. This
	// field gets removed (if set) on a successful issuance and gets set to
	// 1 if unset and an issuance has failed. If an issuance has failed, the
	// delay till the next issuance will be calculated using formula
	// time.Hour * 2 ^ (failedIssuanceAttempts - 1).
	// +optional
	FailedIssuanceAttempts *int `json:"failedIssuanceAttempts,omitempty"`
}

// CertificateCondition contains condition information for an Certificate.
type CertificateCondition struct {
	// Type of the condition, known values are (`Ready`, `Issuing`).
	Type CertificateConditionType `json:"type"`

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

	// If set, this represents the .metadata.generation that the condition was
	// set based upon.
	// For instance, if .metadata.generation is currently 12, but the
	// .status.condition[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the Certificate.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// CertificateConditionType represents an Certificate condition value.
type CertificateConditionType string

const (
	// CertificateConditionReady indicates that a certificate is ready for use.
	// This is defined as:
	// - The target secret exists
	// - The target secret contains a certificate that has not expired
	// - The target secret contains a private key valid for the certificate
	// - The commonName and dnsNames attributes match those specified on the Certificate
	CertificateConditionReady CertificateConditionType = "Ready"

	// A condition added to Certificate resources when an issuance is required.
	// This condition will be automatically added and set to true if:
	//   * No keypair data exists in the target Secret
	//   * The data stored in the Secret cannot be decoded
	//   * The private key and certificate do not have matching public keys
	//   * If a CertificateRequest for the current revision exists and the
	//     certificate data stored in the Secret does not match the
	//    `status.certificate` on the CertificateRequest.
	//   * If no CertificateRequest resource exists for the current revision,
	//     the options on the Certificate resource are compared against the
	//     X.509 data in the Secret, similar to what's done in earlier versions.
	//     If there is a mismatch, an issuance is triggered.
	// This condition may also be added by external API consumers to trigger
	// a re-issuance manually for any other reason.
	//
	// It will be removed by the 'issuing' controller upon completing issuance.
	CertificateConditionIssuing CertificateConditionType = "Issuing"
)

// CertificateSecretTemplate defines the default labels and annotations
// to be copied to the Kubernetes Secret resource named in `CertificateSpec.secretName`.
type CertificateSecretTemplate struct {
	// Annotations is a key value map to be copied to the target Kubernetes Secret.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels is a key value map to be copied to the target Kubernetes Secret.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// NameConstraints is a type to represent x509 NameConstraints
type NameConstraints struct {
	// if true then the name constraints are marked critical.
	//
	// +optional
	Critical bool `json:"critical,omitempty"`
	// Permitted contains the constraints in which the names must be located.
	//
	// +optional
	Permitted *NameConstraintItem `json:"permitted,omitempty"`
	// Excluded contains the constraints which must be disallowed. Any name matching a
	// restriction in the excluded field is invalid regardless
	// of information appearing in the permitted
	//
	// +optional
	Excluded *NameConstraintItem `json:"excluded,omitempty"`
}

type NameConstraintItem struct {
	// DNSDomains is a list of DNS domains that are permitted or excluded.
	//
	// +optional
	DNSDomains []string `json:"dnsDomains,omitempty"`
	// IPRanges is a list of IP Ranges that are permitted or excluded.
	// This should be a valid CIDR notation.
	//
	// +optional
	IPRanges []string `json:"ipRanges,omitempty"`
	// EmailAddresses is a list of Email Addresses that are permitted or excluded.
	//
	// +optional
	EmailAddresses []string `json:"emailAddresses,omitempty"`
	// URIDomains is a list of URI domains that are permitted or excluded.
	//
	// +optional
	URIDomains []string `json:"uriDomains,omitempty"`
}
