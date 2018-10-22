/*
Copyright 2018 The Kubernetes Authors.

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
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"k8s.io/kops/node-authorizer/pkg/authorizers"
	"k8s.io/kops/node-authorizer/pkg/server"
)

// hc is the http client
var hc = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

type gceNodeVerifier struct{}

// NewVerifier creates and returns a verifier
func NewVerifier() (server.Verifier, error) {
	return &gceNodeVerifier{}, nil
}

// Verify is responsible for build a identification document
func (a *gceNodeVerifier) VerifyIdentity(ctx context.Context) ([]byte, error) {
	claim, err := getLocalInstanceIdentityClaim(ctx, AudienceNodeBootstrap)
	if err != nil {
		return nil, err
	}

	// @step: construct request for the request
	request := &authorizers.Request{
		Document: []byte(claim),
	}

	j, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error serializing request: %v", err)
	}

	return j, nil
}
