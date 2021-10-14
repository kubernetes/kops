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

package v1alpha3

// Common annotation keys added to resources.
const (
	// Annotation key for DNS subjectAltNames.
	AltNamesAnnotationKey = "cert-manager.io/alt-names"

	// Annotation key for IP subjectAltNames.
	IPSANAnnotationKey = "cert-manager.io/ip-sans"

	// Annotation key for URI subjectAltNames.
	URISANAnnotationKey = "cert-manager.io/uri-sans"

	// Annotation key for certificate common name.
	CommonNameAnnotationKey = "cert-manager.io/common-name"

	// Annotation key the 'name' of the Issuer resource.
	IssuerNameAnnotationKey = "cert-manager.io/issuer-name"

	// Annotation key for the 'kind' of the Issuer resource.
	IssuerKindAnnotationKey = "cert-manager.io/issuer-kind"

	// Annotation key for the 'group' of the Issuer resource.
	IssuerGroupAnnotationKey = "cert-manager.io/issuer-group"

	// Annotation key for the name of the certificate that a resource is related to.
	CertificateNameKey = "cert-manager.io/certificate-name"

	// Annotation key used to denote whether a Secret is named on a Certificate
	// as a 'next private key' Secret resource.
	IsNextPrivateKeySecretLabelKey = "cert-manager.io/next-private-key"
)

// Deprecated annotation names for Secrets
// These will be removed in a future release.
const (
	DeprecatedIssuerNameAnnotationKey = "certmanager.k8s.io/issuer-name"
	DeprecatedIssuerKindAnnotationKey = "certmanager.k8s.io/issuer-kind"
)

const (
	// issuerNameAnnotation can be used to override the issuer specified on the
	// created Certificate resource.
	IngressIssuerNameAnnotationKey = "cert-manager.io/issuer"
	// clusterIssuerNameAnnotation can be used to override the issuer specified on the
	// created Certificate resource. The Certificate will reference the
	// specified *ClusterIssuer* instead of normal issuer.
	IngressClusterIssuerNameAnnotationKey = "cert-manager.io/cluster-issuer"
	// acmeIssuerHTTP01IngressClassAnnotation can be used to override the http01 ingressClass
	// if the challenge type is set to http01
	IngressACMEIssuerHTTP01IngressClassAnnotationKey = "acme.cert-manager.io/http01-ingress-class"

	// IngressClassAnnotationKey picks a specific "class" for the Ingress. The
	// controller only processes Ingresses with this annotation either unset, or
	// set to either the configured value or the empty string.
	IngressClassAnnotationKey = "kubernetes.io/ingress.class"
)

// Annotation names for CertificateRequests
const (
	// Annotation added to CertificateRequest resources to denote the name of
	// a Secret resource containing the private key used to sign the CSR stored
	// on the resource.
	// This annotation *may* not be present, and is used by the 'self signing'
	// issuer type to self-sign certificates.
	CertificateRequestPrivateKeyAnnotationKey = "cert-manager.io/private-key-secret-name"

	// Annotation to declare the CertificateRequest "revision", belonging to a Certificate Resource
	CertificateRequestRevisionAnnotationKey = "cert-manager.io/certificate-revision"
)

const (
	// IssueTemporaryCertificateAnnotation is an annotation that can be added to
	// Certificate resources.
	// If it is present, a temporary internally signed certificate will be
	// stored in the target Secret resource whilst the real Issuer is processing
	// the certificate request.
	IssueTemporaryCertificateAnnotation = "cert-manager.io/issue-temporary-certificate"
)

// Common/known resource kinds.
const (
	ClusterIssuerKind      = "ClusterIssuer"
	IssuerKind             = "Issuer"
	CertificateKind        = "Certificate"
	CertificateRequestKind = "CertificateRequest"
)

const (
	// WantInjectAnnotation is the annotation that specifies that a particular
	// object wants injection of CAs.  It takes the form of a reference to a certificate
	// as namespace/name.  The certificate is expected to have the is-serving-for annotations.
	WantInjectAnnotation = "cert-manager.io/inject-ca-from"

	// WantInjectAPIServerCAAnnotation, if set to "true", will make the cainjector
	// inject the CA certificate for the Kubernetes apiserver into the resource.
	// It discovers the apiserver's CA by inspecting the service account credentials
	// mounted into the cainjector pod.
	WantInjectAPIServerCAAnnotation = "cert-manager.io/inject-apiserver-ca"

	// WantInjectFromSecretAnnotation is the annotation that specifies that a particular
	// object wants injection of CAs.  It takes the form of a reference to a Secret
	// as namespace/name.
	WantInjectFromSecretAnnotation = "cert-manager.io/inject-ca-from-secret"

	// AllowsInjectionFromSecretAnnotation is an annotation that must be added
	// to Secret resource that want to denote that they can be directly
	// injected into injectables that have a `inject-ca-from-secret` annotation.
	// If an injectable references a Secret that does NOT have this annotation,
	// the cainjector will refuse to inject the secret.
	AllowsInjectionFromSecretAnnotation = "cert-manager.io/allow-direct-injection"
)

// Issuer specific Annotations
const (
	// VenafiCustomFieldsAnnotationKey is the annotation that passes on JSON encoded custom fields to the Venafi issuer
	// This will only work with Venafi TPP v19.3 and higher
	// The value is an array with objects containing the name and value keys
	// for example: `[{"name": "custom-field", "value": "custom-value"}]`
	VenafiCustomFieldsAnnotationKey = "venafi.cert-manager.io/custom-fields"
)

// KeyUsage specifies valid usage contexts for keys.
// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
// Valid KeyUsage values are as follows:
// "signing",
// "digital signature",
// "content commitment",
// "key encipherment",
// "key agreement",
// "data encipherment",
// "cert sign",
// "crl sign",
// "encipher only",
// "decipher only",
// "any",
// "server auth",
// "client auth",
// "code signing",
// "email protection",
// "s/mime",
// "ipsec end system",
// "ipsec tunnel",
// "ipsec user",
// "timestamping",
// "ocsp signing",
// "microsoft sgc",
// "netscape sgc"
// +kubebuilder:validation:Enum="signing";"digital signature";"content commitment";"key encipherment";"key agreement";"data encipherment";"cert sign";"crl sign";"encipher only";"decipher only";"any";"server auth";"client auth";"code signing";"email protection";"s/mime";"ipsec end system";"ipsec tunnel";"ipsec user";"timestamping";"ocsp signing";"microsoft sgc";"netscape sgc"
type KeyUsage string

const (
	UsageSigning           KeyUsage = "signing"
	UsageDigitalSignature  KeyUsage = "digital signature"
	UsageContentCommitment KeyUsage = "content commitment"
	UsageKeyEncipherment   KeyUsage = "key encipherment"
	UsageKeyAgreement      KeyUsage = "key agreement"
	UsageDataEncipherment  KeyUsage = "data encipherment"
	UsageCertSign          KeyUsage = "cert sign"
	UsageCRLSign           KeyUsage = "crl sign"
	UsageEncipherOnly      KeyUsage = "encipher only"
	UsageDecipherOnly      KeyUsage = "decipher only"
	UsageAny               KeyUsage = "any"
	UsageServerAuth        KeyUsage = "server auth"
	UsageClientAuth        KeyUsage = "client auth"
	UsageCodeSigning       KeyUsage = "code signing"
	UsageEmailProtection   KeyUsage = "email protection"
	UsageSMIME             KeyUsage = "s/mime"
	UsageIPsecEndSystem    KeyUsage = "ipsec end system"
	UsageIPsecTunnel       KeyUsage = "ipsec tunnel"
	UsageIPsecUser         KeyUsage = "ipsec user"
	UsageTimestamping      KeyUsage = "timestamping"
	UsageOCSPSigning       KeyUsage = "ocsp signing"
	UsageMicrosoftSGC      KeyUsage = "microsoft sgc"
	UsageNetscapeSGC       KeyUsage = "netscape sgc"
)
