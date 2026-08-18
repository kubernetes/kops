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

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type == "Ready")].status`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.conditions[?(@.type == "Ready")].message`,priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:resource:scope=Cluster,shortName=ciss,categories=cert-manager
// +kubebuilder:subresource:status

// A ClusterIssuer represents a certificate issuing authority which can be
// referenced as part of `issuerRef` fields.
// It is similar to an Issuer, however it is cluster-scoped and therefore can
// be referenced by resources that exist in *any* namespace, not just the same
// namespace as the referent.
type ClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Desired state of the ClusterIssuer resource.
	Spec IssuerSpec `json:"spec"`

	// Status of the ClusterIssuer. This is set and managed automatically.
	// +optional
	Status IssuerStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterIssuerList is a list of Issuers
type ClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterIssuer `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type == "Ready")].status`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.conditions[?(@.type == "Ready")].message`,priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:resource:scope=Namespaced,shortName=iss,categories=cert-manager
// +kubebuilder:subresource:status

// An Issuer represents a certificate issuing authority which can be
// referenced as part of `issuerRef` fields.
// It is scoped to a single namespace and can therefore only be referenced by
// resources within the same namespace.
type Issuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Desired state of the Issuer resource.
	Spec IssuerSpec `json:"spec"`

	// Status of the Issuer. This is set and managed automatically.
	// +optional
	Status IssuerStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IssuerList is a list of Issuers
type IssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Issuer `json:"items"`
}

// IssuerSpec is the specification of an Issuer. This includes any
// configuration required for the issuer.
type IssuerSpec struct {
	IssuerConfig `json:",inline"`
}

// The configuration for the issuer.
// Only one of these can be set.
type IssuerConfig struct {
	// ACME configures this issuer to communicate with a RFC8555 (ACME) server
	// to obtain signed x509 certificates.
	// +optional
	ACME *cmacme.ACMEIssuer `json:"acme,omitempty"`

	// CA configures this issuer to sign certificates using a signing CA keypair
	// stored in a Secret resource.
	// This is used to build internal PKIs that are managed by cert-manager.
	// +optional
	CA *CAIssuer `json:"ca,omitempty"`

	// Vault configures this issuer to sign certificates using a HashiCorp Vault
	// PKI backend.
	// +optional
	Vault *VaultIssuer `json:"vault,omitempty"`

	// SelfSigned configures this issuer to 'self sign' certificates using the
	// private key used to create the CertificateRequest object.
	// +optional
	SelfSigned *SelfSignedIssuer `json:"selfSigned,omitempty"`

	// Venafi configures this issuer to sign certificates using a CyberArk Certificate Manager Self-Hosted
	// or SaaS policy zone.
	// +optional
	Venafi *VenafiIssuer `json:"venafi,omitempty"`
}

// Configures an issuer to sign certificates using a CyberArk Certificate Manager Self-Hosted
// or SaaS policy zone.
// +kubebuilder:validation:XValidation:rule="(has(self.tpp) ? 1 : 0) + (has(self.cloud) ? 1 : 0) + (has(self.ngts) ? 1 : 0) == 1",message="exactly one of tpp, cloud, or ngts must be configured"
type VenafiIssuer struct {
	// Zone is the Certificate Manager Policy Zone to use for this issuer.
	// All requests made to the Certificate Manager platform will be restricted by the named
	// zone policy.
	// This field is required.
	Zone string `json:"zone"`

	// TPP specifies CyberArk Certificate Manager Self-Hosted configuration settings.
	// Only one of CyberArk Certificate Manager may be specified.
	// +optional
	TPP *VenafiTPP `json:"tpp,omitempty"`

	// Cloud specifies the CyberArk Certificate Manager SaaS configuration settings.
	// Only one of CyberArk Certificate Manager may be specified.
	// +optional
	Cloud *VenafiCloud `json:"cloud,omitempty"`

	// NGTS specifies Palo Alto Networks Next Generation Trust Services (NGTS) configuration
	// using OAuth 2.0 Client Credentials. Only one of tpp, cloud, or ngts may be specified.
	// +optional
	NGTS *VenafiNGTS `json:"ngts,omitempty"`
}

// VenafiNGTS defines connection configuration for the Palo Alto Networks
// Next Generation Trust Services (NGTS) platform using OAuth 2.0 Client Credentials.
type VenafiNGTS struct {
	// URL is the base URL for the NGTS API endpoint.
	// Defaults to "https://api.strata.paloaltonetworks.com/ngts" if not set.
	// +optional
	URL string `json:"url,omitempty"`

	// TokenEndpoint is the OAuth 2.0 token endpoint URL used to obtain access tokens,
	// for example "https://auth.apps.paloaltonetworks.com/oauth2/access_token".
	// Defaults to "https://auth.apps.paloaltonetworks.com/oauth2/access_token" if not set.
	// +optional
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`

	// TSGID is the Tenant Service Group ID used to scope the OAuth 2.0 access token,
	// for example "1234567890". The tsg_id: prefix is added automatically.
	// This field is required.
	TSGID string `json:"tsgID"`

	// CredentialsRef is a reference to a Kubernetes Secret containing the OAuth 2.0
	// Client ID and Client Secret. The secret must contain the keys 'client-id' and
	// 'client-secret'.
	CredentialsRef cmmeta.LocalObjectReference `json:"credentialsRef"`
}

// VenafiTPP defines connection configuration details for a CyberArk Certificate Manager Self-Hosted instance
type VenafiTPP struct {
	// URL is the base URL for the vedsdk endpoint of the CyberArk Certificate Manager Self-Hosted instance,
	// for example: "https://tpp.example.com/vedsdk".
	URL string `json:"url"`

	// CredentialsRef is a reference to a Secret containing the CyberArk Certificate Manager Self-Hosted API credentials.
	// The secret must contain the key 'access-token' for the Access Token Authentication,
	// or two keys, 'username' and 'password' for the API Keys Authentication.
	CredentialsRef cmmeta.LocalObjectReference `json:"credentialsRef"`

	// Base64-encoded bundle of PEM CAs which will be used to validate the certificate
	// chain presented by the CyberArk Certificate Manager Self-Hosted server. Only used if using HTTPS; ignored for HTTP.
	// If undefined, the certificate bundle in the cert-manager controller container
	// is used to validate the chain.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// Reference to a Secret containing a base64-encoded bundle of PEM CAs
	// which will be used to validate the certificate chain presented by the CyberArk Certificate Manager Self-Hosted server.
	// Only used if using HTTPS; ignored for HTTP. Mutually exclusive with CABundle.
	// If neither CABundle nor CABundleSecretRef is defined, the certificate bundle in
	// the cert-manager controller container is used to validate the TLS connection.
	// +optional
	CABundleSecretRef *cmmeta.SecretKeySelector `json:"caBundleSecretRef,omitempty"`
}

// VenafiCloud defines connection configuration details for CyberArk Certificate Manager SaaS
type VenafiCloud struct {
	// URL is the base URL for CyberArk Certificate Manager SaaS.
	// Defaults to "https://api.venafi.cloud/".
	// +optional
	URL string `json:"url,omitempty"`

	// APITokenSecretRef is a secret key selector for the CyberArk Certificate Manager SaaS API token.
	APITokenSecretRef cmmeta.SecretKeySelector `json:"apiTokenSecretRef"`
}

// Configures an issuer to 'self sign' certificates using the
// private key used to create the CertificateRequest object.
type SelfSignedIssuer struct {
	// The CRL distribution points is an X.509 v3 certificate extension which identifies
	// the location of the CRL from which the revocation of this certificate can be checked.
	// If not set certificate will be issued without CDP. Values are strings.
	// +optional
	// +listType=atomic
	CRLDistributionPoints []string `json:"crlDistributionPoints,omitempty"`
}

// Configures an issuer to sign certificates using a HashiCorp Vault
// PKI backend.
type VaultIssuer struct {
	// Auth configures how cert-manager authenticates with the Vault server.
	Auth VaultAuth `json:"auth"`

	// Server is the connection address for the Vault server, e.g: "https://vault.example.com:8200".
	Server string `json:"server"`

	// ServerName is used to verify the hostname on the returned certificates
	// by the Vault server.
	// +optional
	ServerName string `json:"serverName,omitempty"`

	// Path is the mount path of the Vault PKI backend's `sign` endpoint, e.g:
	// "my_pki_mount/sign/my-role-name".
	Path string `json:"path"`

	// Name of the vault namespace. Namespaces is a set of features within Vault Enterprise that allows Vault environments to support Secure Multi-tenancy. e.g: "ns1"
	// More about namespaces can be found here https://www.vaultproject.io/docs/enterprise/namespaces
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Base64-encoded bundle of PEM CAs which will be used to validate the certificate
	// chain presented by Vault. Only used if using HTTPS to connect to Vault and
	// ignored for HTTP connections.
	// Mutually exclusive with CABundleSecretRef.
	// If neither CABundle nor CABundleSecretRef are defined, the certificate bundle in
	// the cert-manager controller container is used to validate the TLS connection.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// Reference to a Secret containing a bundle of PEM-encoded CAs to use when
	// verifying the certificate chain presented by Vault when using HTTPS.
	// Mutually exclusive with CABundle.
	// If neither CABundle nor CABundleSecretRef are defined, the certificate bundle in
	// the cert-manager controller container is used to validate the TLS connection.
	// If no key for the Secret is specified, cert-manager will default to 'ca.crt'.
	// +optional
	CABundleSecretRef *cmmeta.SecretKeySelector `json:"caBundleSecretRef,omitempty"`

	// Reference to a Secret containing a PEM-encoded Client Certificate to use when the
	// Vault server requires mTLS.
	// +optional
	ClientCertSecretRef *cmmeta.SecretKeySelector `json:"clientCertSecretRef,omitempty"`

	// Reference to a Secret containing a PEM-encoded Client Private Key to use when the
	// Vault server requires mTLS.
	// +optional
	ClientKeySecretRef *cmmeta.SecretKeySelector `json:"clientKeySecretRef,omitempty"`
}

// VaultAuth is configuration used to authenticate with a Vault server. The
// order of precedence is [`tokenSecretRef`, `appRole`, `clientCertificate`, `kubernetes`, `aws`].
type VaultAuth struct {
	// TokenSecretRef authenticates with Vault by presenting a token.
	// +optional
	TokenSecretRef *cmmeta.SecretKeySelector `json:"tokenSecretRef,omitempty"`

	// AppRole authenticates with Vault using the App Role auth mechanism,
	// with the role and secret stored in a Kubernetes Secret resource.
	// +optional
	AppRole *VaultAppRole `json:"appRole,omitempty"`

	// ClientCertificate authenticates with Vault by presenting a client
	// certificate during the request's TLS handshake.
	// Works only when using HTTPS protocol.
	// +optional
	ClientCertificate *VaultClientCertificateAuth `json:"clientCertificate,omitempty"`

	// Kubernetes authenticates with Vault by passing the ServiceAccount
	// token stored in the named Secret resource to the Vault server.
	// +optional
	Kubernetes *VaultKubernetesAuth `json:"kubernetes,omitempty"`

	// AWS authenticates with Vault using AWS IAM authentication.
	// This allows authentication using IAM roles for service accounts (IRSA),
	// EKS Pod Identity (PIA), or ambient credentials (EC2 instance profiles, ECS task role).
	// +optional
	AWS *VaultAWSAuth `json:"aws,omitempty"`
}

// VaultAppRole authenticates with Vault using the App Role auth mechanism,
// with the role and secret stored in a Kubernetes Secret resource.
type VaultAppRole struct {
	// Path where the App Role authentication backend is mounted in Vault, e.g:
	// "approle"
	Path string `json:"path"`

	// RoleID configured in the App Role authentication backend when setting
	// up the authentication backend in Vault.
	RoleId string `json:"roleId"`

	// Reference to a key in a Secret that contains the App Role secret used
	// to authenticate with Vault.
	// The `key` field must be specified and denotes which entry within the Secret
	// resource is used as the app role secret.
	SecretRef cmmeta.SecretKeySelector `json:"secretRef"`
}

// VaultClientCertificateAuth is used to authenticate against Vault using a client
// certificate stored in a Secret.
type VaultClientCertificateAuth struct {
	// The Vault mountPath here is the mount path to use when authenticating with
	// Vault. For example, setting a value to `/v1/auth/foo`, will use the path
	// `/v1/auth/foo/login` to authenticate with Vault. If unspecified, the
	// default value "/v1/auth/cert" will be used.
	// +optional
	Path string `json:"mountPath,omitempty"`

	// Reference to Kubernetes Secret of type "kubernetes.io/tls" (hence containing
	// tls.crt and tls.key) used to authenticate to Vault using TLS client
	// authentication.
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// Name of the certificate role to authenticate against.
	// If not set, matching any certificate role, if available.
	// +optional
	Name string `json:"name,omitempty"`
}

// Authenticate against Vault using a Kubernetes ServiceAccount token stored in
// a Secret.
type VaultKubernetesAuth struct {
	// The Vault mountPath here is the mount path to use when authenticating with
	// Vault. For example, setting a value to `/v1/auth/foo`, will use the path
	// `/v1/auth/foo/login` to authenticate with Vault. If unspecified, the
	// default value "/v1/auth/kubernetes" will be used.
	// +optional
	Path string `json:"mountPath,omitempty"`

	// The required Secret field containing a Kubernetes ServiceAccount JWT used
	// for authenticating with Vault. Use of 'ambient credentials' is not
	// supported.
	// +optional
	SecretRef cmmeta.SecretKeySelector `json:"secretRef,omitempty"`
	// Note: we don't use a pointer here for backwards compatibility.

	// A reference to a service account that will be used to request a bound
	// token (also known as "projected token"). Compared to using "secretRef",
	// using this field means that you don't rely on statically bound tokens. To
	// use this field, you must configure an RBAC rule to let cert-manager
	// request a token.
	// +optional
	ServiceAccountRef *ServiceAccountRef `json:"serviceAccountRef,omitempty"`

	// A required field containing the Vault Role to assume. A Role binds a
	// Kubernetes ServiceAccount with a set of Vault policies.
	Role string `json:"role"`
}

// ServiceAccountRef is a service account used by cert-manager to request a
// token. By default two audiences are included: the address of the Vault server as specified
// on the issuer, and a generated audience taking the form of `vault://namespace-name/issuer-name`
// for an Issuer and `vault://issuer-name` for a ClusterIssuer. The expiration of the
// token is also set by cert-manager to 10 minutes.
type ServiceAccountRef struct {
	// Name of the ServiceAccount used to request a token.
	Name string `json:"name"`
	// TokenAudiences is an optional list of extra audiences to include in the token passed to Vault.
	// The default audiences are always included in the token.
	// +optional
	// +listType=atomic
	TokenAudiences []string `json:"audiences,omitempty"`
}

// VaultAWSAuth authenticates with Vault using AWS IAM authentication.
// See https://www.vaultproject.io/docs/auth/aws for more details.
type VaultAWSAuth struct {
	// The Vault mountPath here is the mount path to use when authenticating with
	// Vault. For example, setting a value to `/v1/auth/foo`, will use the path
	// `/v1/auth/foo/login` to authenticate with Vault. If unspecified, the
	// default value "/v1/auth/aws" will be used.
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// A required field containing the Vault Role to assume when authenticating.
	// +required
	// +kubebuilder:validation:MinLength=1
	Role string `json:"role"`

	// The AWS region to use for authentication. If not specified, the region
	// will be determined from AWS_REGION or AWS_DEFAULT_REGION environment
	// variables, falling back to "us-east-1" if not set.
	// +optional
	Region string `json:"region,omitempty"`

	// A reference to a service account that will be used to request a web identity
	// token for IRSA (IAM Roles for Service Accounts) authentication.
	// +optional
	ServiceAccountRef *ServiceAccountRef `json:"serviceAccountRef,omitempty"`

	// The ARN of the AWS IAM role to assume using the Kubernetes service account
	// token. Required when using IRSA (serviceAccountRef is set).
	// This role must have a trust policy that allows the OIDC provider to assume it.
	// +optional
	IAMRoleARN string `json:"iamRoleArn,omitempty"`

	// The Vault header value to include in the STS signing request.
	// This is used to prevent replay attacks.
	// +optional
	VaultHeaderValue string `json:"vaultHeaderValue,omitempty"`
}

type CAIssuer struct {
	// SecretName is the name of the secret used to sign Certificates issued
	// by this Issuer.
	SecretName string `json:"secretName"`

	// The CRL distribution points is an X.509 v3 certificate extension which identifies
	// the location of the CRL from which the revocation of this certificate can be checked.
	// If not set, certificates will be issued without distribution points set.
	// +optional
	// +listType=atomic
	CRLDistributionPoints []string `json:"crlDistributionPoints,omitempty"`

	// The OCSP server list is an X.509 v3 extension that defines a list of
	// URLs of OCSP responders. The OCSP responders can be queried for the
	// revocation status of an issued certificate. If not set, the
	// certificate will be issued with no OCSP servers set. For example, an
	// OCSP server URL could be "http://ocsp.int-x3.letsencrypt.org".
	// +optional
	// +listType=atomic
	OCSPServers []string `json:"ocspServers,omitempty"`

	// IssuingCertificateURLs is a list of URLs which this issuer should embed into certificates
	// it creates. See https://www.rfc-editor.org/rfc/rfc5280#section-4.2.2.1 for more details.
	// As an example, such a URL might be "http://ca.domain.com/ca.crt".
	// +optional
	// +listType=atomic
	IssuingCertificateURLs []string `json:"issuingCertificateURLs,omitempty"`
}

// IssuerStatus contains status information about an Issuer
type IssuerStatus struct {
	// List of status conditions to indicate the status of a CertificateRequest.
	// Known condition types are `Ready`.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []IssuerCondition `json:"conditions,omitempty"`

	// ACME specific status options.
	// This field should only be set if the Issuer is configured to use an ACME
	// server to issue certificates.
	// +optional
	ACME *cmacme.ACMEIssuerStatus `json:"acme,omitempty"`
}

// IssuerCondition contains condition information for an Issuer.
type IssuerCondition struct {
	// Type of the condition, known values are (`Ready`).
	Type IssuerConditionType `json:"type"`

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
	// with respect to the current state of the Issuer.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// IssuerConditionType represents an Issuer condition value.
type IssuerConditionType string

const (
	// IssuerConditionReady represents the fact that a given Issuer condition
	// is in ready state and able to issue certificates.
	// If the `status` of this condition is `False`, CertificateRequest controllers
	// should prevent attempts to sign certificates.
	IssuerConditionReady IssuerConditionType = "Ready"
)
