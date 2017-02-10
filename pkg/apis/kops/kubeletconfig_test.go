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

package kops

import (
	"testing"
)

func TestTaintsAppliedAfter160(t *testing.T) {
	exp := map[string]bool{
		"1.4.9":         false,
		"1.5.2":         false,
		"1.6.0-alpha.1": true,
		"1.6.0":         true,
		"1.6.5":         true,
		"1.7.0":         true,
	}

	for ver, e := range exp {
		helpTestTaintsForV(t, ver, e)
	}

}

func helpTestTaintsForV(t *testing.T, version string, shouldApply bool) {
	cluster := &Cluster{Spec: ClusterSpec{KubernetesVersion: version}}
	ig := &InstanceGroup{Spec: InstanceGroupSpec{Taints: []string{"foo", "bar", "baz"}}}
	c, _ := BuildKubeletConfigSpec(cluster, ig)

	var expTaints []string

	if shouldApply {
		expTaints = []string{"foo", "bar", "baz"}

		if c.RegisterSchedulable == nil || !*c.RegisterSchedulable {
			t.Fatalf("Expected RegistSchedulable == &true, got %v", c.RegisterSchedulable)
		}
	} else {
		if c.RegisterSchedulable != nil {
			t.Fatalf("Expected RegisterSchedulable == nil, got %v", *c.RegisterSchedulable)
		}
	}

	if !aEqual(expTaints, c.Taints) {
		t.Fatalf("Expected taints %v, got %v", expTaints, c.Taints)
	}
}

func aEqual(exp, other []string) bool {
	if exp == nil && other != nil {
		return false
	}

	if exp != nil && other == nil {
		return false
	}

	if len(exp) != len(other) {
		return false
	}

	for i, e := range exp {
		if other[i] != e {
			return false
		}
	}

	return true
}
