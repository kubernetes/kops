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

const (

	// Common label keys added to resources
	// Label key that indicates that a resource is of interest to
	// cert-manager controller By default this is set on
	// certificate.spec.secretName secret as well as on the temporary
	// private key Secret. If using SecretsFilteredCaching feature, you
	// might want to set this (with a value of 'true') to any other Secrets
	// that cert-manager controller needs to read, such as issuer
	// credentials Secrets.
	// fao = 'for attention of'
	// See https://github.com/cert-manager/cert-manager/blob/master/design/20221205-memory-management.md#risks-and-mitigations
	PartOfCertManagerControllerLabelKey = "controller.cert-manager.io/fao"

	// Common annotation keys added to resources

	// Annotation key for DNS subjectAltNames.
	AltNamesAnnotationKey = "cert-manager.io/alt-names"

	// Annotation key for IP subjectAltNames.
	IPSANAnnotationKey = "cert-manager.io/ip-sans"

	// Annotation key for URI subjectAltNames.
	URISANAnnotationKey = "cert-manager.io/uri-sans"

	// Annotation key for certificate common name.
	CommonNameAnnotationKey = "cert-manager.io/common-name"

	// Duration key for certificate duration.
	DurationAnnotationKey = "cert-manager.io/duration"

	// Annotation key for certificate renewBefore.
	RenewBeforeAnnotationKey = "cert-manager.io/renew-before"

	// Annotation key for certificate renewBeforePercentage.
	RenewBeforePercentageAnnotationKey = "cert-manager.io/renew-before-percentage"

	// Annotation key for emails subjectAltNames.
	EmailsAnnotationKey = "cert-manager.io/email-sans"

	// Annotation key for subject organization.
	SubjectOrganizationsAnnotationKey = "cert-manager.io/subject-organizations"

	// Annotation key for subject organizational units.
	SubjectOrganizationalUnitsAnnotationKey = "cert-manager.io/subject-organizationalunits"

	// Annotation key for subject organizational units.
	SubjectCountriesAnnotationKey = "cert-manager.io/subject-countries"

	// Annotation key for subject provinces.
	SubjectProvincesAnnotationKey = "cert-manager.io/subject-provinces"

	// Annotation key for subject localities.
	SubjectLocalitiesAnnotationKey = "cert-manager.io/subject-localities"

	// Annotation key for subject provinces.
	SubjectStreetAddressesAnnotationKey = "cert-manager.io/subject-streetaddresses"

	// Annotation key for subject postal codes.
	SubjectPostalCodesAnnotationKey = "cert-manager.io/subject-postalcodes"

	// Annotation key for subject serial number.
	SubjectSerialNumberAnnotationKey = "cert-manager.io/subject-serialnumber"

	// Annotation key for certificate key usages.
	UsagesAnnotationKey = "cert-manager.io/usages"

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

	// Annotation key used to limit the number of CertificateRequests to be kept for a Certificate.
	// Minimum value is 1.
	// If unset all CertificateRequests will be kept.
	RevisionHistoryLimitAnnotationKey = "cert-manager.io/revision-history-limit"

	// Annotation key used to set the PrivateKeyAlgorithm for a Certificate.
	// If PrivateKeyAlgorithm is specified and `size` is not provided,
	// key size of 256 will be used for `ECDSA` key algorithm and
	// key size of 2048 will be used for `RSA` key algorithm.
	// key size is ignored when using the `Ed25519` key algorithm.
	// If unset an algorithm `RSA` will be used.
	PrivateKeyAlgorithmAnnotationKey = "cert-manager.io/private-key-algorithm"

	// Annotation key used to set the PrivateKeyEncoding for a Certificate.
	// If provided, allowed values are `PKCS1` and `PKCS8` standing for PKCS#1
	// and PKCS#8, respectively.
	// If unset an encoding `PKCS1` will be used.
	PrivateKeyEncodingAnnotationKey = "cert-manager.io/private-key-encoding"

	// Annotation key used to set the size of the private key for a Certificate.
	// If PrivateKeyAlgorithm is set to `RSA`, valid values are `2048`, `4096` or `8192`,
	// and will default to `2048` if not specified.
	// If PrivateKeyAlgorithm is set to `ECDSA`, valid values are `256`, `384` or `521`,
	// and will default to `256` if not specified.
	// If PrivateKeyAlgorithm is set to `Ed25519`, Size is ignored.
	// No other values are allowed.
	PrivateKeySizeAnnotationKey = "cert-manager.io/private-key-size"

	// Annotation key used to set the PrivateKeyRotationPolicy for a Certificate.
	// If unset a policy `Never` will be used.
	PrivateKeyRotationPolicyAnnotationKey = "cert-manager.io/private-key-rotation-policy"
)

const (
	// IngressIssuerNameAnnotationKey holds the issuerNameAnnotation value which can be
	// used to override the issuer specified on the created Certificate resource.
	IngressIssuerNameAnnotationKey = "cert-manager.io/issuer"
	// IngressClusterIssuerNameAnnotationKey holds the clusterIssuerNameAnnotation value which
	// can be used to override the issuer specified on the created Certificate resource. The Certificate
	// will reference the specified *ClusterIssuer* instead of normal issuer.
	IngressClusterIssuerNameAnnotationKey = "cert-manager.io/cluster-issuer"
	// IngressACMEIssuerHTTP01IngressClassAnnotationKey holds the acmeIssuerHTTP01IngressClassAnnotation value
	// which can be used to override the http01 ingressClass if the challenge type is set to http01
	IngressACMEIssuerHTTP01IngressClassAnnotationKey = "acme.cert-manager.io/http01-ingress-class"

	// IngressClassAnnotationKey picks a specific "class" for the Ingress. The
	// controller only processes Ingresses with this annotation either unset, or
	// set to either the configured value or the empty string.
	IngressClassAnnotationKey = "kubernetes.io/ingress.class"

	// IngressSecretTemplate can be used to set the secretTemplate field in the generated Certificate.
	// The value is a JSON representation of secretTemplate and must not have any unknown fields.
	IngressSecretTemplate = "cert-manager.io/secret-template"
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

	// WantInjectAPIServerCAAnnotation will - if set to "true" - make the cainjector
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

	// VenafiPickupIDAnnotationKey is the annotation key used to record the
	// Venafi Pickup ID of a certificate signing request that has been submitted
	// to the Venafi API for collection later.
	VenafiPickupIDAnnotationKey = "venafi.cert-manager.io/pickup-id"
)

// KeyUsage specifies valid usage contexts for keys.
// See:
// https://tools.ietf.org/html/rfc5280#section-4.2.1.3
// https://tools.ietf.org/html/rfc5280#section-4.2.1.12
//
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

// Keystore specific secret keys
const (
	// PKCS12SecretKey is the name of the data entry in the Secret resource
	// used to store the p12 file.
	PKCS12SecretKey = "keystore.p12"
	// Data Entry Name in the Secret resource for PKCS12 containing Certificate Authority
	PKCS12TruststoreKey = "truststore.p12"

	// JKSSecretKey is the name of the data entry in the Secret resource
	// used to store the jks file.
	JKSSecretKey = "keystore.jks"
	// Data Entry Name in the Secret resource for JKS containing Certificate Authority
	JKSTruststoreKey = "truststore.jks"

	// The password used to encrypt the keystore and truststore
	KeystorePassword = "keystorePassword"
)

// DefaultKeyUsages contains the default list of key usages
func DefaultKeyUsages() []KeyUsage {
	// The serverAuth EKU is required as of Mac OS Catalina: https://support.apple.com/en-us/HT210176
	// Without this usage, certificates will _always_ flag a warning in newer Mac OS browsers.
	// We don't explicitly add it here as it leads to strange behaviour when a user sets isCA: true
	// (in which case, 'serverAuth' on the CA can break a lot of clients).
	// CAs can (and often do) opt to automatically add usages.
	return []KeyUsage{UsageDigitalSignature, UsageKeyEncipherment}
}
