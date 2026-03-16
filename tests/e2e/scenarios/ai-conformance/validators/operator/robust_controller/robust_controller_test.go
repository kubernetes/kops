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

package robust_controller

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestOperator_RobustController_ViaKuberay corresponds to the operator/robust_controller conformance requirement,
// tested using the KubeRay operator as an example of a complex AI operator with a CRD.
func TestOperator_RobustController_ViaKuberay(t *testing.T) {
	// Description:
	// The platform must prove that at least one complex AI operator with a CRD (e.g., Ray, Kubeflow) can be installed and functions reliably. This includes verifying that the operator's pods run correctly, its webhooks are operational, and its custom resources can be reconciled.

	h := validators.NewValidatorHarness(t)

	if !h.HasCRD("rayjobs.ray.io") {
		h.Skip("Ray CRDs not found, skipping test")
	}

	h.Logf("# Robust Controller (with KubeRay)")

	h.Logf("## Verify KubeRay with a sample RayJob")
	{
		// This is based on https://docs.ray.io/en/latest/cluster/kubernetes/getting-started/rayjob-quick-start.html#kuberay-rayjob-quickstart

		ns := h.TestNamespace()

		h.ApplyManifest(ns, "testdata/rayjob-sample.yaml")
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s --for='jsonpath={.status.jobDeploymentStatus}=Complete' rayjob/rayjob-sample --timeout=300s", ns))

		logs := h.ShellExec(fmt.Sprintf("kubectl logs -n %s -l=job-name=rayjob-sample", ns))

		succeeded := false
		for _, line := range strings.Split(logs.Stdout(), "\n") {
			if strings.Contains(line, "SUCC cli.py") && strings.Contains(line, "Job 'rayjob-sample-") && strings.Contains(line, " succeeded") {
				h.Success("Found succeeded message in logs, indicating the RayJob completed successfully.")
				succeeded = true
				break
			}
		}
		if !succeeded {
			h.Fatalf("Did not find succeeded message in logs: %s", logs.Stdout())
		}
	}

	if h.AllPassed() {
		h.RecordConformance("operator", "robust_controller")
	}
}
