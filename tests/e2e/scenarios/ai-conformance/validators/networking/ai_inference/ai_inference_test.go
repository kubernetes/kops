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
	"fmt"
	"strings"
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

		h.ApplyManifest(ns, "testdata/weighted-traffic-splitting.yaml")
		// h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for='jsonpath={.status.jobDeploymentStatus}=Complete' rayjob/rayjob-sample --timeout=300s", ns))

		status := h.ShellExec(fmt.Sprintf("kubectl get httproute weighted-traffic-splitting -n %s -oyaml", ns))
		if !strings.Contains(status.Stdout(), "Accepted") {
			h.Fatalf("Did not find Accepted message in status: %s", status.Stdout())
		} else {
			h.Success("Found Accepted message in status, indicating the HTTPRoute was accepted successfully.")
		}
	})

	h.Run("header-based-routing", func(h *validators.ValidatorHarness) {
		h.Logf("## Verify Header Based Routing")
		ns := h.TestNamespace()

		h.ApplyManifest(ns, "testdata/header-based-routing.yaml")
		// h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for='jsonpath={.status.jobDeploymentStatus}=Complete' rayjob/rayjob-sample --timeout=300s", ns))

		status := h.ShellExec(fmt.Sprintf("kubectl get httproute header-based-routing -n %s -oyaml", ns))
		if !strings.Contains(status.Stdout(), "Accepted") {
			h.Fatalf("Did not find Accepted message in status: %s", status.Stdout())
		} else {
			h.Success("Found Accepted message in status, indicating the HTTPRoute was accepted successfully.")
		}
	})

	if h.AllPassed() {
		h.RecordConformance("networking/ai_inference")
	}
}
