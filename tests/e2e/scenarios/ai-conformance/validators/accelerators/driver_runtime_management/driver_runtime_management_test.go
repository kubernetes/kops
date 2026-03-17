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

package driver_runtime_management

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators/kubeobjects"
)

// TestAcceleratorsDriverRuntimeManagement corresponds to the accelerators/driver_runtime_management scenario.
// This test verifies that compatible accelerator drivers and corresponding container runtime configurations
// are correctly installed and maintained on nodes with accelerators.
func TestAcceleratorsDriverRuntimeManagement(t *testing.T) {

	h := validators.NewValidatorHarness(t)

	h.Logf("# Driver and Runtime Management Verification")

	// Step 1: Identify GPU nodes
	h.Logf("## Identifying GPU Nodes")
	result := h.ShellExec("kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{\"\\t\"}{.status.capacity.nvidia\\.com/gpu}{\"\\n\"}{end}'")
	gpuNodesOutput := result.Stdout()

	var gpuNodes []string
	for _, line := range strings.Split(strings.TrimSpace(gpuNodesOutput), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] != "" && fields[1] != "<none>" {
			nodeName := fields[0]
			gpuCount := fields[1]
			h.Logf("* Node %s has %s NVIDIA GPUs", nodeName, gpuCount)
			gpuNodes = append(gpuNodes, nodeName)
		}
	}

	if len(gpuNodes) == 0 {
		h.Skip("No GPU nodes found in cluster; skipping driver runtime management test")
		return
	}

	h.Success("Found %d GPU node(s): %s", len(gpuNodes), strings.Join(gpuNodes, ", "))

	// Step 2: Verify GPU Operator DaemonSet is deployed and healthy
	h.Logf("## Verifying GPU Operator DaemonSet")
	daemonSets := h.ListDaemonSets("gpu-operator")
	if len(daemonSets) == 0 {
		h.Fatalf("No DaemonSets found in gpu-operator namespace. GPU operator may not be installed.")
	}

	// Look for the nvidia-driver-daemonset (or similar) that manages driver installation
	var driverDaemonSet *kubeobjects.DaemonSet
	for _, ds := range daemonSets {
		h.Logf("* Found DaemonSet: %s (Ready: %d/%d)", ds.Name(), ds.NumberReady(), ds.DesiredNumberScheduled())
		// The driver daemonset typically has "driver" in its name
		if strings.Contains(ds.Name(), "driver") {
			driverDaemonSet = ds
		}
	}

	if driverDaemonSet != nil {
		if driverDaemonSet.NumberReady() != driverDaemonSet.DesiredNumberScheduled() {
			h.Fatalf("Driver DaemonSet %s is not fully ready: %d/%d pods ready",
				driverDaemonSet.Name(), driverDaemonSet.NumberReady(), driverDaemonSet.DesiredNumberScheduled())
		}
		h.Success("Driver DaemonSet %s is healthy: %d/%d pods ready",
			driverDaemonSet.Name(), driverDaemonSet.NumberReady(), driverDaemonSet.DesiredNumberScheduled())
	} else {
		h.Logf("Warning: No driver-specific DaemonSet found, but GPU operator is deployed")
	}

	// Step 3: Verify driver installation on GPU nodes using a diagnostic job
	h.Logf("## Verifying Driver Installation with Diagnostic Job")
	ns := h.TestNamespace()
	objects := h.ApplyManifest(ns, "testdata/driver-check.yaml")

	// Wait for the job to complete
	for _, obj := range objects {
		if obj.Kind() == "Job" {
			obj.KubectlWait(validators.WithTimeout("5m"))
		}
	}

	// Get the job logs to verify driver check succeeded
	logsResult := h.ShellExec(fmt.Sprintf("kubectl logs --namespace %s job/driver-check", ns))
	logs := logsResult.Stdout()

	// Verify key indicators in the logs
	if !strings.Contains(logs, "NVIDIA-SMI") {
		h.Fatalf("NVIDIA-SMI not found in driver check output")
	}

	if !strings.Contains(logs, "driver_version") {
		h.Fatalf("driver_version not found in nvidia-smi output")
	}

	if !strings.Contains(logs, "SUCCESS") {
		h.Fatalf("Driver check job did not report success")
	}

	h.Success("NVIDIA driver and runtime successfully verified on GPU nodes")

	// Step 4: Check for DRA integration (future-proofing)
	h.Logf("## Checking for DRA Driver/Runtime Version Exposure")
	if h.HasDeviceClass("gpu.nvidia.com") {
		// Once DRA exposes driver/runtime versions, we should query them here
		h.Logf("* DRA DeviceClass 'gpu.nvidia.com' found")
		h.Logf("* Note: DRA-based driver version verification not yet implemented in this test")
		h.Logf("* Future enhancement: Query driver/runtime versions via DRA APIs")
	} else {
		h.Logf("* DRA DeviceClass 'gpu.nvidia.com' not found (DRA driver version exposure may not be available yet)")
	}

	// Record conformance
	if h.AllPassed() {
		h.RecordConformance("accelerators", "driver_runtime_management")
		h.Success("Driver Runtime Management conformance test PASSED")
	}
}
