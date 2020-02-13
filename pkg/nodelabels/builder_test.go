/*
Copyright 2020 The Kubernetes Authors.

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

package nodelabels

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestBuildNodeLabels(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *kops.Cluster
		ig       *kops.InstanceGroup
		expected map[string]string
	}{
		{
			name: "RoleMaster",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.9.0",
					MasterKubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"master1": "master1",
							"master2": "master2",
						},
					},
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleMaster,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelMaster16: "",
				RoleLabelName15:   RoleMasterLabelValue15,
				"master1":         "master1",
				"master2":         "master2",
				"node1":           "override1",
				"node3":           "override3",
			},
		},
		{
			name: "RoleNode",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.9.0",
					MasterKubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"master1": "master1",
							"master2": "master2",
						},
					},
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "node1",
							"node2": "node2",
						},
					},
				},
			},
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleNode,
					Kubelet: &kops.KubeletConfigSpec{
						NodeLabels: map[string]string{
							"node1": "override1",
							"node3": "override3",
						},
					},
				},
			},
			expected: map[string]string{
				RoleLabelNode16: "",
				RoleLabelName15: RoleNodeLabelValue15,
				"node2":         "node2",
				"node1":         "override1",
				"node3":         "override3",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, _ := BuildNodeLabels(test.cluster, test.ig)
			if !reflect.DeepEqual(out, test.expected) {
				t.Fatalf("Actual result:\n%v\nExpect:\n%v", out, test.expected)
			}
		})
	}
}
