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

package ai_inference

import (
	"testing"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestNetworking_AIInference corresponds to the networking/ai_inference conformance requirement,
// tested using the KubeRay operator as an example of a complex AI operator with a CRD.
func TestNetworking_AIInference(t *testing.T) {
	// Description:
	// Support the Kubernetes Gateway API with an implementation for advanced traffic management for inference services,
	// which enables capabilities like weighted traffic splitting,
	// header-based routing (for OpenAI protocol headers),
	// and optional integration with service meshes."

	h := validators.NewValidatorHarness(t)

	h.Logf("# Gateway API support for AI inference")

	h.Run("weighted-traffic-splitting", func(h *validators.ValidatorHarness) {
		h.Logf("## Verify Weighted Traffic Splitting")
		ns := h.TestNamespace()

		objects := h.ApplyManifest(ns, "testdata/weighted-traffic-splitting.yaml")

		for _, obj := range objects {
			if obj.GVK().Kind == "HTTPRoute" {
				h.Logf("Waiting for HTTPRoute %s to be accepted", obj.Name())
				obj.KubectlWait()
			}
		}
	})

	h.Run("header-based-routing", func(h *validators.ValidatorHarness) {
		h.Logf("## Verify Header Based Routing")
		ns := h.TestNamespace()

		objects := h.ApplyManifest(ns, "testdata/header-based-routing.yaml")

		for _, obj := range objects {
			if obj.GVK().Kind == "HTTPRoute" {
				h.Logf("Waiting for HTTPRoute %s to be accepted", obj.Name())
				obj.KubectlWait()
			}
		}
	})

	if h.AllPassed() {
		h.RecordConformance("networking/ai_inference")
	}
}
