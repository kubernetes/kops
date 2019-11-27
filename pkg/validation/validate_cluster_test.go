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

package validation

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
)

func Test_ValidateNodesNotEnough(t *testing.T) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["node-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleNode,
			},
		},
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID: "i-00001",
				Node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1a"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: "Ready", Status: v1.ConditionTrue},
						},
					},
				},
			},
		},
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID: "i-00002",
				Node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1b"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: "Ready", Status: v1.ConditionFalse},
						},
					},
				},
			},
		},
	}

	{
		v := &ValidationCluster{}
		groups["node-1"].MinSize = 3
		v.validateNodes(groups)
		if len(v.Failures) != 2 {
			printDebug(t, v)
			t.Fatal("Too few nodes not caught")
		}
	}

	{
		groups["node-1"].MinSize = 2
		v := &ValidationCluster{}
		v.validateNodes(groups)
		if len(v.Failures) != 1 {
			printDebug(t, v)
			t.Fatal("Not ready node not caught")
		}
	}

	{
		groups["node-1"].NeedUpdate[0].Node.Status.Conditions[0].Status = v1.ConditionTrue
		v := &ValidationCluster{}
		v.validateNodes(groups)
		if len(v.Failures) != 0 {
			printDebug(t, v)
			t.Fatal("unexpected errors")
		}
	}
}

func Test_ValidateNoPodFailures(t *testing.T) {
	v := &ValidationCluster{}
	err := v.collectPodFailures(dummyPodClient(
		[]map[string]string{
			{
				"name":  "pod1",
				"ready": "true",
				"phase": string(v1.PodRunning),
			},
			{
				"name":  "job1",
				"ready": "false",
				"phase": string(v1.PodSucceeded),
			},
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(v.Failures) != 0 {
		fmt.Printf("failures: %+v\n", v.Failures)
		t.Fatal("no failures expected")
	}
}

func Test_ValidatePodFailure(t *testing.T) {
	v := &ValidationCluster{}
	err := v.collectPodFailures(dummyPodClient(
		[]map[string]string{
			{
				"name":  "pod1",
				"ready": "false",
				"phase": string(v1.PodRunning),
			},
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(v.Failures) != 1 || v.Failures[0].Name != "kube-system/pod1" {
		printDebug(t, v)
		t.Fatal("pod1 failure expected")
	}
}

func printDebug(t *testing.T, v *ValidationCluster) {
	t.Logf("cluster - %d failures", len(v.Failures))
	for _, fail := range v.Failures {
		t.Logf("  failure: %+v", fail)
	}
}

func dummyPodClient(pods []map[string]string) kubernetes.Interface {
	return fake.NewSimpleClientset(makePodList(pods))
}

func dummyPod(podMap map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podMap["name"],
			Namespace: "kube-system",
		},
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodPhase(podMap["phase"]),
			ContainerStatuses: []v1.ContainerStatus{
				{
					Ready: podMap["ready"] == "true",
				},
			},
		},
	}
}

// MakePodList constructs api.PodList from a list of pod attributes
func makePodList(pods []map[string]string) *v1.PodList {
	var list v1.PodList
	for _, pod := range pods {
		list.Items = append(list.Items, dummyPod(pod))
	}
	return &list
}

func Test_ValidateBastionNodes(t *testing.T) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["ig1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ig1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleNode,
			},
		},
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "i-00001",
				Node: nil,
			},
		},
	}

	// When an instancegroup's nodes are not ready, that is an error
	{
		v := &ValidationCluster{}
		groups["ig1"].InstanceGroup.Spec.Role = kopsapi.InstanceGroupRoleNode
		v.validateNodes(groups)
		if len(v.Failures) != 1 {
			printDebug(t, v)
			t.Fatal("Nodes are expected to join cluster")
		} else if v.Failures[0].Message != "machine \"i-00001\" has not yet joined cluster" {
			printDebug(t, v)
			t.Fatalf("unexpected validation failure: %+v", v.Failures[0])
		}
	}

	// Except for a bastion instancegroup - those are not expected to join as nodes
	{
		v := &ValidationCluster{}
		groups["ig1"].InstanceGroup.Spec.Role = kopsapi.InstanceGroupRoleBastion
		v.validateNodes(groups)
		if len(v.Failures) != 0 {
			printDebug(t, v)
			t.Fatal("Bastion nodes are not expected to join cluster")
		}
	}

}
