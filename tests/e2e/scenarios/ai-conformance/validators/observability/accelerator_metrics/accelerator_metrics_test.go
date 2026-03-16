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
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestObservability_AcceleratorMetrics_ViaDCGM corresponds to the observability/accelerator_metrics conformance requirement.
// It verifies the metrics in the case that the cluster is using DCGM as the accelerator metrics solution.
func TestObservability_AcceleratorMetrics_ViaDCGM(t *testing.T) {
	// Description:
	// For supported accelerator types, the platform must allow for the installation and successful operation of at least one accelerator metrics solution
	// that exposes fine-grained performance metrics via a standardized, machine-readable metrics endpoint.
	// This must include a core set of metrics for per-accelerator utilization and memory usage.
	// Additionally, other relevant metrics such as temperature, power draw, and interconnect bandwidth should be exposed
	// if the underlying hardware or virtualization layer makes them available.
	// The list of metrics should align with emerging standards, such as OpenTelemetry metrics, to ensure interoperability.
	// The platform may provide a managed solution, but this is not required for conformance."

	h := validators.NewValidatorHarness(t)

	if !h.HasService(types.NamespacedName{Namespace: "gpu-operator", Name: "nvidia-dcgm-exporter"}) {
		h.Skip("nvidia-dcgm-exporter service not found in gpu-operator namespace; skipping accelerator metrics test")
		return
	}

	h.Logf("# Observability: Accelerator Metrics")

	h.Run("nvidia-metrics", func(h *validators.ValidatorHarness) {
		h.Logf("## Verify NVIDIA Metrics")

		h.ShellExec("kubectl get service -n gpu-operator")

		ns := h.TestNamespace()

		requiredMetrics := []string{
			"DCGM_FI_DEV_GPU_TEMP",
			"DCGM_FI_DEV_POWER_USAGE",
			"DCGM_FI_DEV_GPU_UTIL",
			"DCGM_FI_DEV_FB_USED",
		}

		var metricClasses map[string]bool
		var metricsOutput string

		// Retry scraping metrics, as the DCGM exporter may not have completed its first collection cycle yet.
		const maxAttempts = 5
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			podName := fmt.Sprintf("scrape-accelerator-metrics-%d", attempt)

			h.ShellExec(fmt.Sprintf(
				"kubectl run %s -n %s --image=registry.k8s.io/e2e-test-images/agnhost:2.39 --restart=Never --command -- curl -sS http://nvidia-dcgm-exporter.gpu-operator.svc.cluster.local:9400/metrics",
				podName, ns,
			))
			h.ShellExec(fmt.Sprintf("kubectl wait -n %s pod/%s --for=jsonpath='{.status.phase}'=Succeeded --timeout=120s", ns, podName))

			logs := h.ShellExec(fmt.Sprintf("kubectl logs -n %s %s", ns, podName))
			metricsOutput = logs.Stdout()

			metricClasses = make(map[string]bool)
			for _, line := range strings.Split(metricsOutput, "\n") {
				line = strings.TrimSpace(line)
				// Ignore comment lines
				if strings.HasPrefix(line, "#") {
					continue
				}
				fields := strings.Fields(line)
				// Ignore lines that don't have at least a metric name and value
				if len(fields) < 2 {
					continue
				}
				metric := fields[0]

				// Extract out the metric class, ignoring any labels. For example, from "DCGM_FI_DEV_GPU_TEMP{gpu=\"0\"}" we want "DCGM_FI_DEV_GPU_TEMP".
				metricClass := metric
				if prefix, _, ok := strings.Cut(metric, "{"); ok {
					metricClass = prefix
				}

				// Record the metric class as found
				metricClasses[metricClass] = true
			}

			allFound := true
			for _, m := range requiredMetrics {
				if !metricClasses[m] {
					allFound = false
					break
				}
			}
			if allFound {
				h.Logf("All required metrics found on attempt %d", attempt)
				break
			}

			if attempt < maxAttempts {
				h.Logf("Attempt %d: not all required metrics found, retrying in 10s...", attempt)
				time.Sleep(10 * time.Second)
			}
		}

		h.Logf("Received metrics:\n%s", metricsOutput)

		for _, m := range requiredMetrics {
			if !metricClasses[m] {
				h.Errorf("Did not find expected metric: %s", m)
			} else {
				h.Logf("Found expected metric: %s", m)
			}
		}
	})

	if h.AllPassed() {
		h.RecordConformance("observability", "accelerator_metrics")
	}
}
