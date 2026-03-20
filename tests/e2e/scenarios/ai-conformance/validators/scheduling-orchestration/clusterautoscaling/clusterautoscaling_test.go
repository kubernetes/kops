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

package clusterautoscaling

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// Test_SchedulingOrchestration_ClusterAutoscaling verifies that the cluster autoscaler
// (or equivalent mechanism) can scale up node groups containing GPU accelerators
// based on pending pods requesting those accelerators.
//
// It counts the current number of GPU nodes, deploys N+1 replicas of a simple
// GPU workload (each requesting one GPU), and verifies that the cluster
// scales up to accommodate the additional pod.
func Test_SchedulingOrchestration_ClusterAutoscaling(t *testing.T) {
	// Description:
	// If the platform provides a cluster autoscaler or an equivalent mechanism,
	//   it must be able to scale up/down node groups containing specific accelerator types
	//   based on pending pods requesting those accelerators.

	h := validators.NewValidatorHarness(t)

	h.Logf("# Cluster Autoscaling for GPU Nodes")

	h.Run("cluster-autoscaling-gpu", func(h *validators.ValidatorHarness) {
		ns := h.TestNamespace()

		// Count the current number of nodes with GPUs by looking at resource slices
		// that advertise GPU devices.
		h.Logf("## Determine current GPU node count")

		listGPUNodes := func() []string {
			result := h.ShellExec("kubectl get nodes -l nvidia.com/gpu.present=true -o name")
			nodes := strings.Split(strings.TrimSpace(result.Stdout()), "\n")
			return nodes
		}

		initialGPUNodes := listGPUNodes()
		h.Logf("Found %d GPU nodes initially (%v)", len(initialGPUNodes), initialGPUNodes)

		if len(initialGPUNodes) == 0 {
			h.Fatalf("No GPU nodes found in the cluster; cannot test cluster autoscaling for GPUs")
		}

		// Deploy the GPU probe workload with 1 replica first.
		h.Logf("## Deploy GPU probe workload")
		h.ApplyManifest(ns, "testdata/cluster-autoscaling-workload.yaml")

		// Scale to N+1 replicas to force the autoscaler to add a GPU node.
		targetReplicas := len(initialGPUNodes) + 1
		h.Logf("## Scale deployment to %d replicas (initial GPU nodes: %d)", targetReplicas, len(initialGPUNodes))
		h.ShellExec(fmt.Sprintf("kubectl scale deployment/cluster-autoscaling-workload -n %s --replicas=%d", ns, targetReplicas))

		// Wait for at least one pod to be Pending (confirming we need a new node).
		h.Logf("### Verify at least one pod is pending")
		h.ShellExec(fmt.Sprintf("kubectl get pods -n %s -l app=cluster-autoscaling-workload -o wide", ns))

		// Poll for the GPU node count to increase.
		h.Logf("## Wait for cluster to scale up")
		var scaledUp bool
		const maxAttempts = 40 // 40 * 30s = 20 minutes
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			currentGPUNodes := listGPUNodes()

			if len(currentGPUNodes) > len(initialGPUNodes) {
				h.Logf("Cluster scaled up: GPU nodes increased from %d to %d on attempt %d", len(initialGPUNodes), len(currentGPUNodes), attempt)
				scaledUp = true
				break
			}

			// Periodic diagnostics.
			if attempt%5 == 1 {
				h.Logf("### Diagnostics at attempt %d", attempt)
				h.ShellExec(fmt.Sprintf("kubectl get pods -n %s -l app=cluster-autoscaling-workload -o wide", ns))
				h.ShellExec("kubectl get nodes -o wide")
			}

			if attempt < maxAttempts {
				h.Logf("Attempt %d: GPU node count is still %d (need > %d), waiting 30s...", attempt, len(currentGPUNodes), len(initialGPUNodes))
				time.Sleep(30 * time.Second)
			}
		}

		if !scaledUp {
			// Failure diagnostics.
			h.Logf("### Failure diagnostics")
			h.ShellExec(fmt.Sprintf("kubectl get pods -n %s -l app=cluster-autoscaling-workload -o wide", ns))
			h.ShellExec(fmt.Sprintf("kubectl describe pods -n %s -l app=cluster-autoscaling-workload", ns))
			h.ShellExec("kubectl get nodes -o wide")
			h.ShellExec("kubectl describe nodes")
			h.Errorf("Cluster did not scale up GPU nodes within the expected time (initial: %d)", len(initialGPUNodes))
		}

		// Verify all replicas eventually become ready.
		if scaledUp {
			h.Logf("## Wait for all replicas to be ready")
			result := h.ShellExec(fmt.Sprintf(
				"kubectl rollout status deployment/cluster-autoscaling-workload -n %s --timeout=600s",
				ns,
			))
			if result.Err() != nil {
				h.Errorf("Deployment did not become fully ready: %v", result.Err())
			}

			// Verify GPU pods are actually running nvidia-smi.
			h.Logf("### Verify GPU pods are running")
			h.ShellExec(fmt.Sprintf("kubectl get pods -n %s -l app=cluster-autoscaling-workload -o wide", ns))

			// Check logs from one of the pods to confirm GPU access.
			podListResult := h.ShellExec(fmt.Sprintf(
				"kubectl get pods -n %s -l app=cluster-autoscaling-workload -o name",
				ns,
			))
			for _, podName := range strings.Split(strings.TrimSpace(podListResult.Stdout()), "\n") {
				h.ShellExec(fmt.Sprintf("kubectl logs -n %s %s --tail=5", ns, podName))
			}

			h.Success("Cluster autoscaler scaled up GPU nodes from %d to accommodate %d GPU pods", len(initialGPUNodes), targetReplicas)
		}

		// Scale down and verify the cluster scales back down.
		h.Logf("## Scale down and verify cluster scale-down")
		h.ShellExec(fmt.Sprintf("kubectl scale deployment/cluster-autoscaling-workload -n %s --replicas=0", ns))

		h.Logf("Waiting for cluster to scale down (this may take several minutes)...")
		var scaledDown bool
		const scaleDownMaxAttempts = 40 // 40 * 30s = 20 minutes
		for attempt := 1; attempt <= scaleDownMaxAttempts; attempt++ {
			currentGPUNodes := listGPUNodes()

			if len(currentGPUNodes) <= len(initialGPUNodes) {
				h.Logf("Cluster scaled down: GPU nodes decreased to %d on attempt %d", len(currentGPUNodes), attempt)
				scaledDown = true
				break
			}

			if attempt%5 == 1 {
				h.Logf("### Scale-down diagnostics at attempt %d", attempt)
				h.ShellExec("kubectl get nodes -o wide")
			}

			if attempt < scaleDownMaxAttempts {
				h.Logf("Attempt %d: GPU node count is still %d (need <= %d), waiting 30s...", attempt, len(currentGPUNodes), len(initialGPUNodes))
				time.Sleep(30 * time.Second)
			}
		}

		if !scaledDown {
			h.Logf("### Scale-down failure diagnostics")
			h.ShellExec("kubectl get nodes -o wide")
			h.ShellExec("kubectl describe nodes")
			h.Errorf("Cluster did not scale down GPU nodes within the expected time")
		} else {
			h.Success("Cluster autoscaler scaled down GPU nodes back to %d", len(initialGPUNodes))
		}
	})

	if h.AllPassed() {
		h.RecordConformance("schedulingOrchestration", "cluster_autoscaling")
	}
}
