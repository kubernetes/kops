/*
Copyright 2020 The Jetstack cert-manager contributors.

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
	// If this annotation is specified on a Certificate or Order resource when
	// using the HTTP01 solver type, the ingress.name field of the HTTP01
	// solver's configuration will be set to the value given here.
	// This is especially useful for users of Ingress controllers that maintain
	// a 1:1 mapping between endpoint IP and Ingress resource.
	ACMECertificateHTTP01IngressNameOverride = "acme.cert-manager.io/http01-override-ingress-name"

	// If this annotation is specified on a Certificate or Order resource when
	// using the HTTP01 solver type, the ingress.class field of the HTTP01
	// solver's configuration will be set to the value given here.
	// This is especially useful for users deploying many different ingress
	// classes into a single cluster that want to be able to re-use a single
	// solver for each ingress class.
	ACMECertificateHTTP01IngressClassOverride = "acme.cert-manager.io/http01-override-ingress-class"

	// IngressEditInPlaceAnnotation is used to toggle the use of ingressClass instead
	// of ingress on the created Certificate resource
	IngressEditInPlaceAnnotationKey = "acme.cert-manager.io/http01-edit-in-place"

	// DomainLabelKey is added to the labels of a Pod serving an ACME challenge.
	// Its value will be the hash of the domain name that is being verified.
	DomainLabelKey = "acme.cert-manager.io/http-domain"

	// TokenLabelKey is added to the labels of a Pod serving an ACME challenge.
	// Its value will be the hash of the challenge token that is being served by the pod.
	TokenLabelKey = "acme.cert-manager.io/http-token"

	// SolverIdentificationLabelKey is added to the labels of a Pod serving an ACME challenge.
	// Its value will be the "true" if the Pod is an HTTP-01 solver.
	SolverIdentificationLabelKey = "acme.cert-manager.io/http01-solver"
)

const (
	OrderKind     = "Order"
	ChallengeKind = "Challenge"
)
