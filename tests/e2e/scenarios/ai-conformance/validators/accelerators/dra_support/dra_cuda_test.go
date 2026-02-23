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

package dra_support

import (
	"fmt"
	"testing"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestDRAWorks is an additional test to verify that DRA is not only available, but also functional.
// This is not currently required by AI conformance.
func TestDRAWorks(t *testing.T) {
	h := validators.NewValidatorHarness(t)

	h.Logf("# DRA functional tests")

	h.Logf("## Listing device classes")
	deviceClasses := h.ListDeviceClasses()
	for _, deviceClass := range deviceClasses {
		h.Logf("* %s", deviceClass.Name())
	}

	h.Logf("## Listing resource slices")
	resourceSlices := h.ListResourceSlices()
	for _, resourceSlice := range resourceSlices {
		h.Logf("* %s", resourceSlice.Name())
	}

	if !h.HasDeviceClass("gpu.nvidia.com") {
		t.Skipf("gpu.nvidia.com device class not found; skipping")
	}

	h.Logf("## Run cuda-smoketest")
	ns := "default"
	h.ShellExec(fmt.Sprintf("kubectl apply --namespace %s -f testdata/cuda-smoketest.yaml", ns))
	h.ShellExec(fmt.Sprintf("kubectl wait --for=condition=complete --namespace %s job/cuda-smoketest --timeout=5m", ns))
	h.ShellExec(fmt.Sprintf("kubectl logs --namespace %s job/cuda-smoketest", ns))
}
