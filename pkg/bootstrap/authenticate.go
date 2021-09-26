/*
Copyright 2020 The Kubernetes Authors.

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

package bootstrap

import (
	"context"

	"k8s.io/kops/pkg/apis/nodeup"
)

// Authenticator generates authentication credentials for requests.
type Authenticator interface {
	SignBootstrapRequest(req *nodeup.BootstrapRequest) (*SignedRequestData, error)
}

// SignedRequestData holds info for the bootstrap request.
type SignedRequestData struct {
	// Body is the main content to send
	Body []byte

	// Authorization is the authorization header to use
	Authorization string
}

// VerifyResult is the result of a successfully verified request.
type VerifyResult struct {
	// Request holds the deserialized (and verified) request
	Request *nodeup.BootstrapRequest

	// Nodename is the name that this node is authorized to use.
	NodeName string

	// InstanceGroupName is the name of the kops InstanceGroup this node is a member of.
	InstanceGroupName string

	// CertificateNames is the names the node is authorized to use for certificates.
	CertificateNames []string
}

// Verifier verifies authentication credentials for requests.
type Verifier interface {
	Verify(ctx context.Context, token string, body []byte) (*VerifyResult, error)
}
