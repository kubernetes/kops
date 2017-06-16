/*
Copyright 2016 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"testing"
	"time"
)

func stringPointer(s string) *string {
	return &s
}

func int32Pointer(i32 int32) *int32 {
	return &i32
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
				LogLevel: new(int32),
			},
			Expected: "",
		},
		{
			Config: &kops.KubeletConfigSpec{
				LogLevel: int32Pointer(2),
			},
			Expected: "--v=2",
		},

		// Test string pointers without the "flag-include-empty" tag
		{
			Config: &kops.KubeletConfigSpec{
				EvictionHard: stringPointer("memory.available<100Mi"),
			},
			Expected: "--eviction-hard=memory.available<100Mi",
		},
		{
			Config: &kops.KubeletConfigSpec{
				EvictionHard: stringPointer(""),
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
				ResolverConfig: stringPointer("test"),
			},
			Expected: "--resolv-conf=test",
		},
		{
			Config: &kops.KubeletConfigSpec{
				ResolverConfig: stringPointer(""),
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
