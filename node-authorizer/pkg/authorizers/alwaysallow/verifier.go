/*
Copyright 2019 The Kubernetes Authors.

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

package alwaysallow

import (
	"context"

	"k8s.io/kops/node-authorizer/pkg/server"
	"k8s.io/kops/node-authorizer/pkg/utils"
)

// alwaysallowVerifier implements the verifier
type alwaysallowVerifier struct{}

// NewVerifier creates and returns a new Verifier
func NewVerifier() (server.Verifier, error) {
	utils.Logger.Warn("note the alwaysallow authorizer performs no authoritative checks and should only be used in test environments")

	return &alwaysallowVerifier{}, nil
}

// VerifyIdentity is responsible for providing proof of identity
func (a *alwaysallowVerifier) VerifyIdentity(context.Context) ([]byte, error) {
	return []byte{}, nil
}
