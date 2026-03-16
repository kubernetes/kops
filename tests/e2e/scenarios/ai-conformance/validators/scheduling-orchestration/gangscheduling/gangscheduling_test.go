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

package gangscheduling

import (
	"fmt"
	"testing"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestSchedulingOrchestration_GangScheduling_ViaKueue corresponds to the schedulingOrchestration/gang_scheduling scenario,
// for the case that the vendor chooses to demonstrate gang scheduling support via Kueue.
func TestSchedulingOrchestration_GangScheduling_ViaKueue(t *testing.T) {
	// Description:
	//   The platform must allow for the installation and successful operation of at least one gang scheduling solution that ensures all-or-nothing scheduling for distributed AI workloads (e.g. Kueue, Volcano, etc.) To be conformant, the vendor must demonstrate that their platform can successfully run at least one such solution.

	h := validators.NewValidatorHarness(t)

	if !h.HasCRD("localqueues.kueue.x-k8s.io") {
		h.Skip("Kueue CRDs not found, skipping gang scheduling test via Kueue")
	}

	h.Logf("# Gang Scheduling (via Kueue)")

	h.Run("gangscheduling-kueue", func(h *validators.ValidatorHarness) {
		jobName := "gangscheduling-kueue"

		h.Logf("## Simple gang scheduling test using Kueue")

		h.Logf("Creating a Kueue Job that requires gang scheduling")
		ns := h.TestNamespace()
		h.ApplyManifest(ns, "testdata/gangscheduling-kueue.yaml")

		h.Logf("Waiting for Job to complete")
		h.ShellExec(fmt.Sprintf("kubectl wait --namespace %s --for=condition=complete job/%s --timeout=300s", ns, jobName))

		h.Success("Gang scheduling via Kueue test completed successfully.")
	})

	if h.AllPassed() {
		h.RecordConformance("schedulingOrchestration", "gang_scheduling")
	}
}
