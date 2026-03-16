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
	"sync"
	"testing"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators/conformance"
	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators/kubeobjects"
)

// ValidatorHarness provides a common context and utilities for AI conformance validation.
type ValidatorHarness struct {
	*kubeobjects.Client

	*conformance.Reporter

	t *testing.T

	output OutputSink

	dynamicClient dynamic.Interface
	restConfig    *rest.Config

	// mutex guards our mutable state
	mutex sync.Mutex

	// testNamespace is a per-test namespace to use for creating resources. It is lazily initialized when TestNamespace() is called.
	testNamespace string

	// objectIDs tracks the Kubernetes objects that have been created or observed during the test. This can be used for cleanup or reporting.
	objectIDs []*KubeObjectID
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

	h.Client = kubeobjects.NewClient(h, h.DynamicClient())

	h.Reporter = conformance.NewReporter(h.t)

	return h
}

// Context returns the context associated with the test, which can be used for command execution and API calls.
func (h *ValidatorHarness) Context() context.Context {
	return h.t.Context()
}

// Skip allows the test to be skipped with a message, and ensures that the skip is recorded in the output.
func (h *ValidatorHarness) Skip(message string) {
	h.output.Skip(message)
	h.t.Skip(message)
}
