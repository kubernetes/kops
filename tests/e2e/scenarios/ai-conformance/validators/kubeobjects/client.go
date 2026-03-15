/*
Copyright The Kubernetes Authors.

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

package kubeobjects

import (
	"context"

	"k8s.io/client-go/dynamic"
)

// Client is a wrapper around the Kubernetes API client, for use in tests.
type Client struct {
	t             Testing
	ctx           context.Context
	dynamicClient dynamic.Interface
}

// Testing is a minimal interface for test contexts, allowing the client to report failures and log messages.
type Testing interface {
	// Fatalf reports a test failure with formatted output and stops execution.
	Fatalf(format string, args ...interface{})

	// Context returns the context for the test, which can be used for API calls.
	Context() context.Context
}

// NewClient creates a new Client.
func NewClient(t Testing, dynamicClient dynamic.Interface) *Client {
	ctx := t.Context()
	return &Client{
		dynamicClient: dynamicClient,
		ctx:           ctx,
		t:             t,
	}
}
