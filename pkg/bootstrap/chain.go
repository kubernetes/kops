/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

// NewChainVerifier creates a new Verifier that will return the first positive verification from the provided Verifiers.
func NewChainVerifier(chain ...Verifier) Verifier {
	return &ChainVerifier{chain: chain}
}

// ChainVerifier wraps multiple Verifiers; the first positive verification from any Verifier will be returned.
type ChainVerifier struct {
	chain []Verifier
}

// VerifyToken will return the first positive verification from any Verifier in the chain.
func (v *ChainVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*VerifyResult, error) {
	for _, verifier := range v.chain {
		result, err := verifier.VerifyToken(ctx, rawRequest, token, body)
		if err == nil {
			return result, nil
		}
		if err == ErrNotThisVerifier {
			continue
		}
		if err == ErrAlreadyExists {
			return nil, ErrAlreadyExists
		}
		klog.Infof("failed to verify token: %v", err)
	}
	return nil, fmt.Errorf("unable to verify token")
}
