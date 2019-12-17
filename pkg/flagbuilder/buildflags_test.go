/*
Copyright 2019 The Kubernetes Authors.

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

package flagbuilder

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func resourceValue(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}

func TestBuildKCMFlags(t *testing.T) {
	grid := []struct {
		Config   interface{}
		Expected string
	}{
		{
			Config: &kops.KubeControllerManagerConfig{
				AttachDetachReconcileSyncPeriod: &metav1.Duration{Duration: time.Minute},
			},
			Expected: "--attach-detach-reconcile-sync-period=1m0s",
		},
		{
			Config: &kops.KubeControllerManagerConfig{
				TerminatedPodGCThreshold: fi.Int32(1500),
			},
			Expected: "--terminated-pod-gc-threshold=1500",
		},
		{
			Config: &kops.KubeControllerManagerConfig{
				KubeAPIQPS: resourceValue("42"),
			},
			Expected: "--kube-api-qps=42",
		},
		{
			Config: &kops.KubeControllerManagerConfig{
				KubeAPIBurst: fi.Int32(80),
			},
			Expected: "--kube-api-burst=80",
		},
		{
			Config:   &kops.KubeControllerManagerConfig{},
			Expected: "",
		},
	}

	for _, test := range grid {
		actual, err := BuildFlags(test.Config)
		if err != nil {
			t.Errorf("error from BuildFlags: %v", err)
			continue
		}

		if actual != test.Expected {
			t.Errorf("unexpected flags.  actual=%q expected=%q", actual, test.Expected)
			continue
		}
	}
}

func TestKubeletConfigSpec(t *testing.T) {
	grid := []struct {
		Config   interface{}
		Expected string
	}{
		{
			Config: &kops.KubeletConfigSpec{
				APIServers: "https://example.com",
			},
			Expected: "--api-servers=https://example.com",
		},
		{
			Config: &kops.KubeletConfigSpec{
				EvictionPressureTransitionPeriod: &metav1.Duration{Duration: 5 * time.Second},
			},
			Expected: "--eviction-pressure-transition-period=5s",
		},
		{
			Config: &kops.KubeletConfigSpec{
				LogLevel: fi.Int32(0),
			},
			Expected: "",
		},
		{
			Config: &kops.KubeletConfigSpec{
				LogLevel: fi.Int32(2),
			},
			Expected: "--v=2",
		},

		// Test string pointers without the "flag-include-empty" tag
		{
			Config: &kops.KubeletConfigSpec{
				EvictionHard: fi.String("memory.available<100Mi"),
			},
			Expected: "--eviction-hard=memory.available<100Mi",
		},
		{
			Config: &kops.KubeletConfigSpec{
				EvictionHard: fi.String(""),
			},
			Expected: "",
		},

		// Test string pointers with the "flag-include-empty" tag
		{
			Config:   &kops.KubeletConfigSpec{},
			Expected: "",
		},
		{
			Config: &kops.KubeletConfigSpec{
				ResolverConfig: fi.String("test"),
			},
			Expected: "--resolv-conf=test",
		},
		{
			Config: &kops.KubeletConfigSpec{
				ResolverConfig: fi.String(""),
			},
			Expected: "--resolv-conf=",
		},
		{
			Config: &kops.KubeletConfigSpec{
				ResolverConfig: nil,
			},
			Expected: "",
		},
	}

	for _, test := range grid {
		actual, err := BuildFlags(test.Config)
		if err != nil {
			t.Errorf("error from BuildFlags: %v", err)
			continue
		}

		if actual != test.Expected {
			t.Errorf("unexpected flags.  actual=%q expected=%q", actual, test.Expected)
			continue
		}
	}
}

func TestBuildAPIServerFlags(t *testing.T) {
	grid := []struct {
		Config   interface{}
		Expected string
	}{
		{
			Config:   &kops.KubeAPIServerConfig{},
			Expected: "--insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuditWebhookBatchThrottleQps: resourceValue("3.14"),
			},
			Expected: "--audit-webhook-batch-throttle-qps=3.14 --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuditWebhookBatchThrottleEnable: fi.Bool(true),
			},
			Expected: "--audit-webhook-batch-throttle-enable=true --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuditWebhookBatchThrottleEnable: fi.Bool(false),
			},
			Expected: "--audit-webhook-batch-throttle-enable=false --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuditWebhookInitialBackoff: &metav1.Duration{Duration: 120 * time.Second},
			},
			Expected: "--audit-webhook-initial-backoff=2m0s --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuditWebhookBatchMaxSize: fi.Int32(1000),
			},
			Expected: "--audit-webhook-batch-max-size=1000 --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuthorizationWebhookConfigFile: fi.String("/authorization.yaml"),
			},
			Expected: "--authorization-webhook-config-file=/authorization.yaml --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuthorizationWebhookCacheAuthorizedTTL: &metav1.Duration{Duration: 100 * time.Second},
			},
			Expected: "--authorization-webhook-cache-authorized-ttl=1m40s --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				AuthorizationWebhookCacheUnauthorizedTTL: &metav1.Duration{Duration: 10 * time.Second},
			},
			Expected: "--authorization-webhook-cache-unauthorized-ttl=10s --insecure-port=0 --secure-port=0",
		},
		{
			Config: &kops.KubeAPIServerConfig{
				EventTTL: &metav1.Duration{Duration: 3 * time.Hour},
			},
			Expected: "--event-ttl=3h0m0s --insecure-port=0 --secure-port=0",
		},
	}

	for _, test := range grid {
		actual, err := BuildFlags(test.Config)
		if err != nil {
			t.Errorf("error from BuildFlags: %v", err)
			continue
		}

		if actual != test.Expected {
			t.Errorf("unexpected flags.  actual=%q expected=%q", actual, test.Expected)
			continue
		}
	}
}
