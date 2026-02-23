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

package validators

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ValidatorHarness provides a common context and utilities for AI conformance validation.
type ValidatorHarness struct {
	t *testing.T

	output OutputSink

	dynamicClient dynamic.Interface
	restConfig    *rest.Config
}

// NewValidatorHarness creates a new ValidatorHarness.
func NewValidatorHarness(t *testing.T) *ValidatorHarness {
	h := &ValidatorHarness{t: t}

	h.output = createMarkdownOutput(t)
	h.t.Cleanup(func() {
		if err := h.output.Close(); err != nil {
			h.t.Errorf("failed to close output: %v", err)
		}
	})

	// use the current context in kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			h.Fatalf("failed to get user home directory: %v", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		h.Fatalf("failed to build get kubeconfig: %v", err)
	}
	h.restConfig = restConfig

	return h
}

// Context returns the context associated with the test, which can be used for command execution and API calls.
func (h *ValidatorHarness) Context() context.Context {
	return h.t.Context()
}
