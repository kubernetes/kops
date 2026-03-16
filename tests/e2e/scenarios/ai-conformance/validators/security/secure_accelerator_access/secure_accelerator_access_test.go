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

package secure_accelerator_access

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestSecurity_SecureAcceleratorAccess corresponds to the security/secure_accelerator_access conformance requirement
func TestSecurity_SecureAcceleratorAccess(t *testing.T) {
	// Description:
	// Ensure that access to accelerators from within containers is properly isolated and mediated by the Kubernetes resource management framework (device plugin or DRA) and container runtime, preventing unauthorized access or interference between workloads.
	h := validators.NewValidatorHarness(t)

	h.Logf("# Secure Accelerator Access")

	h.Logf("## Checking that GPUs are available if requested")

	h.Run("accelerator-requested", func(h *validators.ValidatorHarness) {
		ns := h.TestNamespace()

		h.ApplyManifest(ns, "testdata/accelerator-requested.yaml")
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for=condition=complete job/accelerator-requested --timeout=60s", ns))

		logs := h.ShellExec(fmt.Sprintf("kubectl logs -n %s job/accelerator-requested", ns))
		if !strings.Contains(logs.Stdout(), "<product_brand>NVIDIA</product_brand>") {
			h.Errorf("Expected to find nvidia GPUs available when requested, but did not find them in the logs: %s", logs.Stdout())
		} else {
			h.Success("GPUs were requested, and nvidia-smi reported available GPUs.")
		}
	})

	h.Logf("## Checking that GPUs are not available if not requested")
	h.Run("accelerator-not-requested", func(h *validators.ValidatorHarness) {
		ns := h.TestNamespace()

		h.ApplyManifest(ns, "testdata/accelerator-not-requested.yaml")
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for=condition=complete job/accelerator-not-requested --timeout=60s", ns))

		logs := h.ShellExec(fmt.Sprintf("kubectl logs -n %s job/accelerator-not-requested", ns))
		if !strings.Contains(logs.Stdout(), "nvidia-smi failed (as expected)") {
			h.Errorf("Expected nvidia-smi to fail when GPUs are not requested, but found them in the logs: %s", logs.Stdout())
		} else {
			h.Success("No GPUs were requested, and nvidia-smi did not report any GPUs.")
		}
	})

	h.Logf("## Pods with GPU requests should be isolated from each other")
	h.Run("accelerator-isolation", func(h *validators.ValidatorHarness) {
		ns := h.TestNamespace()

		h.ApplyManifest(ns, "testdata/accelerator-isolation.yaml")
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for=condition=available deployment/accelerator-isolation-1 --timeout=60s", ns))
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for=condition=available deployment/accelerator-isolation-2 --timeout=60s", ns))

		logs1 := h.ShellExec(fmt.Sprintf("kubectl logs -n %s deployment/accelerator-isolation-1", ns))
		logs2 := h.ShellExec(fmt.Sprintf("kubectl logs -n %s deployment/accelerator-isolation-2", ns))

		uuid1 := extractGPUUUID(logs1.Stdout())
		uuid2 := extractGPUUUID(logs2.Stdout())

		if uuid1 == "" {
			h.Errorf("Failed to extract GPU UUID from logs of accelerator-isolation-1:\n%s", logs1.Stdout())
		} else if uuid2 == "" {
			h.Errorf("Failed to extract GPU UUID from logs of accelerator-isolation-2:\n%s", logs2.Stdout())
		} else if uuid1 == uuid2 {
			h.Errorf("Expected that pods with GPU requests would be isolated from each other, but both pods saw the same GPU UUID: %s", uuid1)
		} else {
			h.Success("Pods with GPU requests were isolated from each other as expected.")
		}
	})

	if h.AllPassed() {
		h.RecordConformance("security", "secure_accelerator_access")
	}
}

// extractGPUUUID is a helper function to extract the GPU UUID from nvidia-smi XML output in the logs.
func extractGPUUUID(logs string) string {
	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<uuid>") && strings.HasSuffix(line, "</uuid>") {
			value := strings.TrimPrefix(line, "<uuid>")
			value = strings.TrimSuffix(value, "</uuid>")
			return value
		}
	}
	return ""
}
