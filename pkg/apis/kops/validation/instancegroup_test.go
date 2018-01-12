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

package validation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"strings"
	"testing"
)

func TestDefaultTaintsEnforcedBefore160(t *testing.T) {
	type param struct {
		ver       string
		role      kops.InstanceGroupRole
		taints    []string
		shouldErr bool
	}

	params := []param{
		{"1.5.0", kops.InstanceGroupRoleNode, []string{kops.TaintNoScheduleMaster15}, true},
		{"1.5.1", kops.InstanceGroupRoleNode, nil, false},
		{"1.5.2", kops.InstanceGroupRoleNode, []string{}, false},
		{"1.6.0", kops.InstanceGroupRoleNode, []string{kops.TaintNoScheduleMaster15}, false},
		{"1.6.1", kops.InstanceGroupRoleNode, []string{"Foo"}, false},
	}

	for _, p := range params {
		cluster := &kops.Cluster{Spec: kops.ClusterSpec{KubernetesVersion: p.ver}}
		ig := &kops.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: kops.InstanceGroupSpec{
				Taints: p.taints,
				Role:   p.role,
			},
		}

		err := CrossValidateInstanceGroup(ig, cluster, false)
		if p.shouldErr {
			if err == nil {
				t.Fatal("Expected error building kubelet config, received nil.")
			} else if !strings.Contains(err.Error(), "User-specified taints are not supported before kubernetes version 1.6.0") {
				t.Fatalf("Received an unexpected error validating taints: '%s'", err.Error())
			}
		} else {
			if err != nil {
				t.Fatalf("Received an unexpected error validating taints: '%s', params: '%v'", err.Error(), p)
			}
		}
	}
}
