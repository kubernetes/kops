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

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators/kubeobjects"
)

// TestObservability_AIServiceMetrics_ViaPrometheus corresponds to the observability/ai_service_metrics conformance requirement.
// It verifies that the cluster has a monitoring system capable of discovering and collecting
// metrics from workloads that expose them in Prometheus exposition format.
func TestObservability_AIServiceMetrics_ViaPrometheus(t *testing.T) {
	// Description:
	// Provide a monitoring system capable of discovering and collecting metrics
	//  from workloads that expose them in a standard format (e.g. Prometheus exposition format).
	// This ensures easy integration for collecting key metrics from common AI frameworks and servers.

	h := validators.NewValidatorHarness(t)

	h.Logf("# Observability: AI Service Metrics")

	// Find Prometheus in the cluster.
	prometheusEndpoint := findPrometheusEndpoint(h)
	if prometheusEndpoint == "" {
		h.Skip("no Prometheus service found in the cluster; skipping ai_service_metrics test")
		return
	}

	h.Logf("Found Prometheus endpoint: %s", prometheusEndpoint)

	h.Run("ai-service-metrics", func(h *validators.ValidatorHarness) {
		h.Logf("## Deploy a workload exposing Prometheus metrics and verify collection")

		ns := h.TestNamespace()

		// Deploy a workload that exposes a known metric in Prometheus exposition format.
		metricName := "aiconformancetest_requests_total"

		objects := h.ApplyManifest(ns, "testdata/ai-service-metrics.yaml")

		// Wait for the deployment to be ready.
		for _, obj := range objects {
			if obj.Kind() == "Deployment" {
				h.Logf("Waiting for Deployment %s to be ready", obj.Name())
				obj.KubectlWait()
			}
		}

		// Verify the workload actually serves metrics by scraping it directly.
		h.Logf("### Verify workload exposes metrics")
		h.ShellExec(fmt.Sprintf(
			"kubectl run direct-scrape -n %s --image=registry.k8s.io/e2e-test-images/agnhost:2.39 --restart=Never --command -- curl -sS http://metrics-fake-workload.%s.svc.cluster.local:8080/metrics",
			ns, ns,
		))
		h.ShellExec(fmt.Sprintf("kubectl wait -n %s pod/direct-scrape --for=jsonpath='{.status.phase}'=Succeeded --timeout=120s", ns))
		directOutput := h.ShellExec(fmt.Sprintf("kubectl logs -n %s direct-scrape", ns))

		if !strings.Contains(directOutput.Stdout(), metricName) {
			h.Fatalf("Workload is not serving the expected metric %q; cannot proceed with collection test", metricName)
		}
		h.Logf("Confirmed workload exposes metric %s", metricName)

		// Now verify that Prometheus has discovered and collected the metric.
		// We query the Prometheus API, retrying to allow time for scrape discovery.
		h.Logf("### Verify Prometheus collects the metric")

		var found bool
		const maxAttempts = 12
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			queryPodName := fmt.Sprintf("prom-query-%d", attempt)

			// Query Prometheus for our test metric, filtering by namespace.
			queryURL := fmt.Sprintf("%s/api/v1/query", prometheusEndpoint)
			query := fmt.Sprintf("query=%s{namespace=\"%s\"}", metricName, ns)
			h.ShellExec(fmt.Sprintf(
				"kubectl run %s -n %s --image=registry.k8s.io/e2e-test-images/agnhost:2.39 --restart=Never --command -- curl -sS -X POST --data-urlencode '%s' '%s'",
				queryPodName, ns,
				query, queryURL,
			))
			h.ShellExec(fmt.Sprintf("kubectl wait -n %s pod/%s --for=jsonpath='{.status.phase}'=Succeeded --timeout=120s", ns, queryPodName))

			result := h.ShellExec(fmt.Sprintf("kubectl logs -n %s %s", ns, queryPodName))
			output := result.Stdout()
			h.Logf("Prometheus query attempt %d response: %s", attempt, output)

			// A successful Prometheus response looks like: {"status":"success","data":{"resultType":"vector","result":[...]}}
			// If result array is non-empty, the metric was collected.
			if strings.Contains(output, `"success"`) && strings.Contains(output, metricName) && !strings.Contains(output, `"result":[]`) {
				found = true
				h.Logf("Prometheus collected metric %s on attempt %d", metricName, attempt)
				break
			}

			if attempt < maxAttempts {
				h.Logf("Attempt %d: metric not yet collected by Prometheus, retrying in 15s...", attempt)
				time.Sleep(15 * time.Second)
			}
		}

		if !found {
			h.Errorf("Prometheus did not collect metric %s from the test workload within the expected time", metricName)
		} else {
			h.Logf("Successfully verified: monitoring system discovered and collected metrics from a workload exposing Prometheus exposition format")
		}
	})

	if h.AllPassed() {
		h.RecordConformance("observability/ai_service_metrics")
	}
}

// findPrometheusEndpoint searches for a Prometheus service in the cluster and returns its in-cluster URL.
func findPrometheusEndpoint(h *validators.ValidatorHarness) string {
	// Search for services with "prometheus" in the name across all namespaces,
	// excluding operator and alertmanager services.
	services := h.ListServices("")

	var matches []*kubeobjects.Service
	for _, service := range services {
		name := service.Name()
		ns := service.Namespace()

		// I wonder if there's a standard label we could look for?

		if name == "kube-prometheus-stack-prometheus" && ns == "monitoring" {
			matches = append(matches, service)
		}
	}

	if len(matches) == 0 {
		h.Logf("No Prometheus service found in the cluster")
		return ""
	}

	if len(matches) > 1 {
		h.Logf("Multiple Prometheus services found in the cluster; using the first one: %s/%s", matches[0].Namespace(), matches[0].Name())
	}

	match := matches[0]
	h.Logf("Found Prometheus service: %s/%s", match.Namespace(), match.Name())

	return fmt.Sprintf("http://%s.%s.svc.cluster.local:9090", match.Name(), match.Namespace())
}
