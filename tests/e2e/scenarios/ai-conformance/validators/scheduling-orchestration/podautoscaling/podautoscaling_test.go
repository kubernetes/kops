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

package podautoscaling

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// Test_SchedulingOrchestration_PodAutoscaling verifies that the HPA can scale
// pods utilizing accelerators based on custom metrics relevant to AI/ML workloads.
// It deploys vLLM with a GPU, sends load to build up a request queue,
// and verifies that HPA scales the deployment based on the vllm_num_requests_waiting metric
// (renamed from vllm:num_requests_waiting by prometheus-adapter, since colons are not
// compatible with the Kubernetes custom metrics API URL paths).
func Test_SchedulingOrchestration_PodAutoscaling(t *testing.T) {
	// Description:
	// If the platform supports the HorizontalPodAutoscaler,
	//   it must function correctly for pods utilizing accelerators.
	// This includes the ability to scale these Pods based on custom metrics relevant to AI/ML workloads.

	h := validators.NewValidatorHarness(t)

	h.Logf("# Pod Autoscaling with Custom Metrics")

	h.Run("pod-autoscaling-custom-metric", func(h *validators.ValidatorHarness) {
		ns := h.TestNamespace()

		h.Logf("## Deploy vLLM with GPU and HPA")

		h.Logf("Applying vLLM workload with Service, PodMonitor, and HPA")
		objects := h.ApplyManifest(ns, "testdata/pod-autoscaling-workload.yaml")

		// Wait for the vLLM deployment to be ready.
		for _, obj := range objects {
			if obj.Kind() == "Deployment" {
				h.Logf("Waiting for Deployment %s to be ready", obj.Name())
				obj.KubectlWait(validators.WithTimeout("600s")) // Can take a long time to pull the image
			}
		}

		// Verify vLLM is actually serving by hitting its health endpoint.
		h.Logf("### Verify vLLM is serving")
		h.ShellExec(fmt.Sprintf(
			"kubectl run vllm-health-check -n %s --image=registry.k8s.io/e2e-test-images/agnhost:2.39 --restart=Never --command -- curl -sS http://vllm-qwen25-500m.%s.svc.cluster.local/health",
			ns, ns,
		))
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s pod/vllm-health-check --for=jsonpath='{.status.phase}'=Succeeded --timeout=120s", ns))
		healthResult := h.ShellExec(fmt.Sprintf("kubectl logs -n %s vllm-health-check", ns))
		h.Logf("vLLM health check response: %s", healthResult.Stdout())

		// Verify vLLM exposes metrics.
		h.Logf("### Verify vLLM exposes Prometheus metrics")
		h.ShellExec(fmt.Sprintf(
			"kubectl run vllm-metrics-check -n %s --image=registry.k8s.io/e2e-test-images/agnhost:2.39 --restart=Never --command -- curl -sS http://vllm-qwen25-500m.%s.svc.cluster.local/metrics",
			ns, ns,
		))
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s pod/vllm-metrics-check --for=jsonpath='{.status.phase}'=Succeeded --timeout=120s", ns))
		metricsResult := h.ShellExec(fmt.Sprintf("kubectl logs -n %s vllm-metrics-check", ns))
		if !strings.Contains(metricsResult.Stdout(), "vllm:num_requests_waiting") {
			h.Fatalf("vLLM is not exposing the expected metric vllm:num_requests_waiting")
		}
		h.Logf("Confirmed vLLM exposes vllm:num_requests_waiting metric")

		// Check HPA status before load.
		h.Logf("### Check initial HPA status")
		h.ShellExec(fmt.Sprintf("kubectl get hpa -n %s vllm-qwen25-500m -o wide", ns))

		// Deploy load generator.
		h.Logf("## Deploy load generator")
		h.ApplyManifest(ns, "testdata/pod-autoscaling-load-generator.yaml")

		// Poll HPA until it scales up or we time out.
		h.Logf("## Wait for HPA to scale up")
		var scaled bool
		const maxAttempts = 30
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			result := h.ShellExec(fmt.Sprintf(
				"kubectl get hpa -n %s vllm-qwen25-500m -o jsonpath='{.status.desiredReplicas}'",
				ns,
			))
			desiredStr := strings.Trim(result.Stdout(), "'")
			desired, err := strconv.Atoi(desiredStr)
			if err == nil && desired > 1 {
				h.Logf("HPA scaled to %d desired replicas on attempt %d", desired, attempt)
				scaled = true
				break
			}

			// Also log current metrics for debugging.
			hpaStatus := h.ShellExec(fmt.Sprintf("kubectl get hpa -n %s vllm-qwen25-500m -o wide", ns))
			h.Logf("Attempt %d: HPA status: %s", attempt, hpaStatus.Stdout())

			if attempt < maxAttempts {
				h.Logf("Attempt %d: HPA has not scaled yet, waiting 30s...", attempt)
				time.Sleep(30 * time.Second)
			}
		}

		if !scaled {
			// Log HPA events and conditions for debugging.
			h.ShellExec(fmt.Sprintf("kubectl describe hpa -n %s vllm-qwen25-500m", ns))
			h.Errorf("HPA did not scale the deployment above 1 replica within the expected time")
		} else {
			h.Success("HPA successfully scaled vLLM deployment based on custom metric vllm_num_requests_waiting")
		}
	})

	if h.AllPassed() {
		h.RecordConformance("schedulingOrchestration", "pod_autoscaling")
	}
}
