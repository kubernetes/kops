/*
Copyright 2019 The Jetstack cert-manager contributors.

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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChallengePayload describes a request/response for presenting or cleaning up
// an ACME challenge resource
type ChallengePayload struct {
	metav1.TypeMeta `json:",inline"`

	// Request describes the attributes for the ACME solver request
	// +optional
	Request *ChallengeRequest `json:"request,omitempty"`

	// Response describes the attributes for the ACME solver response
	// +optional
	Response *ChallengeResponse `json:"response,omitempty"`
}

// ChallengeRequest is a payload that can be sent to external ACME webhook
// solvers in order to 'Present' or 'CleanUp' a challenge with an ACME server.
type ChallengeRequest struct {
	// UID is an identifier for the individual request/response. It allows us to distinguish instances of requests which are
	// otherwise identical (parallel requests, requests when earlier requests did not modify etc)
	// The UID is meant to track the round trip (request/response) between the KAS and the WebHook, not the user request.
	// It is suitable for correlating log entries between the webhook and apiserver, for either auditing or debugging.
	UID types.UID `json:"uid"`

	// Action is one of 'present' or 'cleanup'.
	// If the action is 'present', the record will be presented with the
	// solving service.
	// If the action is 'cleanup', the record will be cleaned up with the
	// solving service.
	Action ChallengeAction `json:"action"`

	// Type is the type of ACME challenge.
	// Only dns-01 is currently supported.
	Type string `json:"type"`

	// DNSName is the name of the domain that is actually being validated, as
	// requested by the user on the Certificate resource.
	// This will be of the form 'example.com' from normal hostnames, and
	// '*.example.com' for wildcards.
	DNSName string `json:"dnsName"`

	// Key is the key that should be presented.
	// This key will already be signed by the account that owns the challenge.
	// For DNS01, this is the key that should be set for the TXT record for
	// ResolveFQDN.
	Key string `json:"key"`

	// ResourceNamespace is the namespace containing resources that are
	// referenced in the providers config.
	// If this request is solving for an Issuer resource, this will be the
	// namespace of the Issuer.
	// If this request is solving for a ClusterIssuer resource, this will be
	// the configured 'cluster resource namespace'
	ResourceNamespace string `json:"resourceNamespace"`

	// ResolvedFQDN is the fully-qualified domain name that should be
	// updated/presented after resolving all CNAMEs.
	// This should be honoured when using the DNS01 solver type.
	// This will be of the form '_acme-challenge.example.com.'.
	// +optional
	ResolvedFQDN string `json:"resolvedFQDN,omitempty"`

	// ResolvedZone is the zone encompassing the ResolvedFQDN.
	// This is included as part of the ChallengeRequest so that webhook
	// implementers do not need to implement their own SOA recursion logic.
	// This indicates the zone that the provided FQDN is encompassed within,
	// determined by performing SOA record queries for each part of the FQDN
	// until an authoritative zone is found.
	// This will be of the form 'example.com.'.
	ResolvedZone string `json:"resolvedZone,omitempty"`

	// AllowAmbientCredentials advises webhook implementations that they can
	// use 'ambient credentials' for authenticating with their respective
	// DNS provider services.
	// This field SHOULD be honoured by all DNS webhook implementations, but
	// in certain instances where it does not make sense to honour this option,
	// an implementation may ignore it.
	AllowAmbientCredentials bool `json:"allowAmbientCredentials"`

	// Config contains unstructured JSON configuration data that the webhook
	// implementation can unmarshal in order to fetch secrets or configure
	// connection details etc.
	// Secret values should not be passed in this field, in favour of
	// references to Kubernetes Secret resources that the webhook can fetch.
	// +optional
	Config *apiext.JSON `json:"config,omitempty"`
}

type ChallengeAction string

const (
	ChallengeActionPresent ChallengeAction = "Present"
	ChallengeActionCleanUp ChallengeAction = "CleanUp"
)

type ChallengeResponse struct {
	// UID is an identifier for the individual request/response.
	// This should be copied over from the corresponding ChallengeRequest.
	UID types.UID `json:"uid"`

	// Success will be set to true if the request action (i.e. presenting or
	// cleaning up) was successful.
	Success bool `json:"success"`

	// Result contains extra details into why a challenge request failed.
	// This field will be completely ignored if 'success' is true.
	// +optional
	Result *metav1.Status `json:"status,omitempty"`
}
