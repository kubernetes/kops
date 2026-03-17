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

		// Verify the custom metrics pipeline is working before generating load.
		h.Logf("### Verify custom metrics pipeline")

		// Check that the custom metrics API is registered.
		h.Logf("Checking custom metrics API registration...")
		h.ShellExec("kubectl get apiservice v1beta1.custom.metrics.k8s.io -o wide || true")

		// Check prometheus-adapter is running.
		h.Logf("Checking prometheus-adapter deployment status...")
		h.ShellExec("kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus-adapter -o wide || true")

		// Query Prometheus directly to see if vllm:num_requests_waiting is being scraped.
		h.Logf("Querying Prometheus for vllm:num_requests_waiting metric...")
		h.ShellExec(
			"kubectl exec -n monitoring statefulset/prometheus-kube-prometheus-stack-prometheus -c prometheus -- wget -qO- 'http://localhost:9090/api/v1/query?query=vllm%3Anum_requests_waiting' 2>/dev/null || echo 'Failed to query Prometheus'",
		)

		// Query Prometheus for scrape targets to see if vLLM pod is being scraped.
		h.Logf("Checking Prometheus scrape targets for vLLM...")
		h.ShellExec(
			"kubectl exec -n monitoring statefulset/prometheus-kube-prometheus-stack-prometheus -c prometheus -- wget -qO- 'http://localhost:9090/api/v1/targets?state=active' 2>/dev/null | grep -o '\"scrapePool\":\"[^\"]*vllm[^\"]*\"' || echo 'No vLLM scrape targets found'",
		)

		// Query the custom metrics API directly to see if prometheus-adapter is serving the metric.
		h.Logf("Querying custom metrics API for vllm_num_requests_waiting...")
		h.ShellExec(fmt.Sprintf(
			"kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/%s/pods/*/vllm_num_requests_waiting 2>&1 || echo 'Custom metric not available via API'",
			ns,
		))

		// Also list all available custom metrics to see what the adapter is exposing.
		h.Logf("Listing all available custom metrics...")
		h.ShellExec("kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 2>&1 || echo 'Custom metrics API not available'")

		// Check HPA status before load.
		h.Logf("### Check initial HPA status")
		h.ShellExec(fmt.Sprintf("kubectl get hpa -n %s vllm-qwen25-500m -o wide", ns))
		h.ShellExec(fmt.Sprintf("kubectl describe hpa -n %s vllm-qwen25-500m", ns))

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
			h.Logf("Checking hpa status...")
			h.ShellExec(fmt.Sprintf("kubectl get hpa -n %s vllm-qwen25-500m -o wide", ns))

			// Every 5 attempts, do deeper diagnostics on the metrics pipeline.
			if attempt%5 == 1 {
				h.Logf("### Diagnostics at attempt %d", attempt)

				// Check custom metrics API for the metric value.
				h.ShellExec(fmt.Sprintf(
					"kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/%s/pods/*/vllm_num_requests_waiting 2>&1 || echo 'Custom metric not available'",
					ns,
				))

				// Query Prometheus for the raw metric.
				h.ShellExec(
					"kubectl exec -n monitoring statefulset/prometheus-kube-prometheus-stack-prometheus -c prometheus -- wget -qO- 'http://localhost:9090/api/v1/query?query=vllm%3Anum_requests_waiting' 2>/dev/null || echo 'Failed to query Prometheus'",
				)

				// Check HPA conditions for error messages.
				h.ShellExec(fmt.Sprintf(
					"kubectl get hpa -n %s vllm-qwen25-500m -o jsonpath='{.status.conditions}' || true",
					ns,
				))

				// Check prometheus-adapter logs for errors.
				h.ShellExec(
					"kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus-adapter --tail=20 || true",
				)

				// Check if load generator is actually running.
				h.ShellExec(fmt.Sprintf(
					"kubectl get pods -n %s -l app=vllm-load-generator -o wide || true",
					ns,
				))
			}

			if attempt < maxAttempts {
				h.Logf("Attempt %d: HPA has not scaled yet, waiting 30s...", attempt)
				time.Sleep(30 * time.Second)
			}
		}

		if !scaled {
			// Comprehensive debugging on failure.
			h.Logf("### Failure diagnostics")

			h.ShellExec(fmt.Sprintf("kubectl describe hpa -n %s vllm-qwen25-500m", ns))

			// Final check of custom metrics API.
			h.ShellExec(fmt.Sprintf(
				"kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/%s/pods/*/vllm_num_requests_waiting 2>&1 || echo 'Custom metric not available'",
				ns,
			))

			// List all custom metrics the adapter knows about.
			h.ShellExec("kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 2>&1 || echo 'Custom metrics API not available'")

			// Prometheus query for the raw metric.
			h.ShellExec(
				"kubectl exec -n monitoring statefulset/prometheus-kube-prometheus-stack-prometheus -c prometheus -- wget -qO- 'http://localhost:9090/api/v1/query?query=vllm%3Anum_requests_waiting' 2>/dev/null || echo 'Failed to query Prometheus'",
			)

			// prometheus-adapter logs.
			h.ShellExec("kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus-adapter --tail=50 || true")

			// prometheus-adapter config (to verify the rule).
			h.ShellExec("kubectl get configmap -n monitoring prometheus-adapter -o yaml 2>&1 || echo 'ConfigMap not found'")

			// Check PodMonitor was created and Prometheus discovered it.
			h.ShellExec(fmt.Sprintf("kubectl get podmonitor -n %s -o yaml || true", ns))

			h.Errorf("HPA did not scale the deployment above 1 replica within the expected time")
		} else {
			h.Success("HPA successfully scaled vLLM deployment based on custom metric vllm_num_requests_waiting")
		}
	})

	if h.AllPassed() {
		h.RecordConformance("schedulingOrchestration", "pod_autoscaling")
	}
}
