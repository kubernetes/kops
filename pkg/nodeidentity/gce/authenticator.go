/*
Copyright 2021 The Kubernetes Authors.

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

package gce

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"

	"cloud.google.com/go/compute/metadata"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// gceAuthenticator performs authentication using GCE signed tokens.
type gceAuthenticator struct {
	audience string
}

var _ fi.Authenticator = &gceAuthenticator{}

// AudidenceFor returns the expected audience used in the token for the cluster.
func AudienceFor(cluster *kops.Cluster) string {
	// The audience shouldn't collide with other services (impersonation/reuse attack),
	// but doesn't have to match the DNS name
	return "kops-controller." + cluster.Name
}

// NewAuthenticator constructs and returns an Authenticator using GCE signed tokens.
func NewAuthenticator(audience string) (fi.Authenticator, error) {
	return &gceAuthenticator{
		audience: audience,
	}, nil
}

// CreateToken creates a signed token bound to the specified body.
func (a *gceAuthenticator) CreateToken(body []byte) (string, error) {
	sha := sha256.Sum256(body)

	// Ensure the signature is only valid for this particular body content.
	audience := a.audience + "//" + base64.URLEncoding.EncodeToString(sha[:])

	suffix := "instance/service-accounts/default/identity?format=full&audience=" + url.QueryEscape(audience)

	token, err := metadata.Get(suffix)
	if err != nil {
		return "", fmt.Errorf("unable to get token from GCE metadata service: %w", err)
	}

	return token, nil
}
