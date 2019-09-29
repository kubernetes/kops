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

// alwaysAllowAuth is the implementation for a node authozier
type alwaysAllowAuth struct{}

// NewAuthorizer creates and returns a alwaysAllow node authorizer
func NewAuthorizer() (server.Authorizer, error) {
	utils.Logger.Warn("note the alwaysallow authorizer performs no authoritative checks and should only be used in test environments")

	return &alwaysAllowAuth{}, nil
}

// Authorize is responsible for accepting the request
func (a *alwaysAllowAuth) Authorize(_ context.Context, req *server.NodeRegistration) error {
	req.Status.Allowed = true

	return nil
}

// Name returns the name of the authorizer
func (a *alwaysAllowAuth) Name() string {
	return "alwaysAllow"
}

// Close is called when the authorizer is shutting down
func (a *alwaysAllowAuth) Close() error {
	return nil
}
