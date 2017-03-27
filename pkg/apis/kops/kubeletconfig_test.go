/*
Copyright 2017 The Kubernetes Authors.

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

var taintValidationError = "User-specified taints are not supported before kubernetes version 1.6.0"

func Test_InstanceGroupKubeletMerge(t *testing.T) {
	var cluster = &Cluster{}
	cluster.Spec.Kubelet = &KubeletConfigSpec{}
	cluster.Spec.Kubelet.NvidiaGPUs = 0
	cluster.Spec.KubernetesVersion = "1.6.0"

	var instanceGroup = &InstanceGroup{}
	instanceGroup.Spec.Kubelet = &KubeletConfigSpec{}
	instanceGroup.Spec.Kubelet.NvidiaGPUs = 1
	instanceGroup.Spec.Role = InstanceGroupRoleNode

	var mergedKubeletSpec, err = BuildKubeletConfigSpec(cluster, instanceGroup)
	if err != nil {
		t.Error(err)
	}
	if mergedKubeletSpec == nil {
		t.Error("Returned nil kubelet spec")
	}

	if mergedKubeletSpec.NvidiaGPUs != instanceGroup.Spec.Kubelet.NvidiaGPUs {
		t.Errorf("InstanceGroup kubelet value (%d) should be reflected in merged output", instanceGroup.Spec.Kubelet.NvidiaGPUs)
	}
}

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

func TestDefaultTaintsEnforcedBefore160(t *testing.T) {
	type param struct {
		ver       string
		role      InstanceGroupRole
		taints    []string
		shouldErr bool
	}

	params := []param{
		{"1.5.0", InstanceGroupRoleNode, []string{TaintNoScheduleMaster}, true},
		{"1.5.1", InstanceGroupRoleNode, nil, false},
		{"1.5.2", InstanceGroupRoleNode, []string{}, false},
		{"1.6.0", InstanceGroupRoleNode, []string{TaintNoScheduleMaster}, false},
		{"1.6.1", InstanceGroupRoleNode, []string{"Foo"}, false},
	}

	for _, p := range params {
		cluster := &Cluster{Spec: ClusterSpec{KubernetesVersion: p.ver}}
		ig := &InstanceGroup{Spec: InstanceGroupSpec{
			Taints: p.taints,
			Role:   p.role,
		}}
		_, err := BuildKubeletConfigSpec(cluster, ig)
		if p.shouldErr {
			if err == nil {
				t.Fatal("Expected error building kubelet config, received nil.")
			} else if err.Error() != taintValidationError {
				t.Fatalf("Received an unexpected error validating taints: '%s'", err.Error())
			}
		} else {
			if err != nil {
				t.Fatalf("Received an unexpected error validating taints: '%s', params: '%v'", err.Error(), p)
			}
		}
	}
}

func helpTestTaintsForV(t *testing.T, version string, shouldApply bool) {
	cluster := &Cluster{Spec: ClusterSpec{KubernetesVersion: version}}
	ig := &InstanceGroup{Spec: InstanceGroupSpec{Role: InstanceGroupRoleMaster, Taints: []string{"foo", "bar", "baz"}}}
	c, err := BuildKubeletConfigSpec(cluster, ig)

	var expTaints []string

	if shouldApply {
		expTaints = []string{"foo", "bar", "baz"}

		if c.RegisterSchedulable == nil || !*c.RegisterSchedulable {
			t.Fatalf("Expected RegisterSchedulable == &true, got %v", c.RegisterSchedulable)
		}

		if !aEqual(expTaints, c.Taints) {
			t.Fatalf("Expected taints %v, got %v", expTaints, c.Taints)
		}
	} else if err == nil || err.Error() != taintValidationError {
		t.Fatalf("Received an unexpected error: '%s'", err.Error())
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
