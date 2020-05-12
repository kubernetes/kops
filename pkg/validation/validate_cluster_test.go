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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type MockCloud struct {
	awsup.MockAWSCloud
	Groups map[string]*cloudinstances.CloudInstanceGroup

	t                      *testing.T
	expectedCluster        *kopsapi.Cluster
	expectedInstanceGroups []kopsapi.InstanceGroup
}

var _ fi.Cloud = (*MockCloud)(nil)

func BuildMockCloud(t *testing.T, groups map[string]*cloudinstances.CloudInstanceGroup, expectedCluster *kopsapi.Cluster, expectedInstanceGroups []kopsapi.InstanceGroup) *MockCloud {
	m := MockCloud{
		MockAWSCloud:           *awsup.BuildMockAWSCloud("us-east-1", "abc"),
		Groups:                 groups,
		t:                      t,
		expectedCluster:        expectedCluster,
		expectedInstanceGroups: expectedInstanceGroups,
	}
	return &m
}

func (c *MockCloud) GetCloudGroups(cluster *kopsapi.Cluster, instancegroups []*kopsapi.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	assert.Equal(c.t, c.expectedCluster, cluster, "cluster")

	var igs = make([]kopsapi.InstanceGroup, 0, len(instancegroups))
	for _, ig := range instancegroups {
		igs = append(igs, *ig)
	}
	assert.ElementsMatch(c.t, c.expectedInstanceGroups, igs)

	// TODO assert nodes contains all the nodes in the mock kubernetes.Interface?

	return c.Groups, nil
}

func testValidate(t *testing.T, groups map[string]*cloudinstances.CloudInstanceGroup, objects []runtime.Object) (*ValidationCluster, error) {
	cluster := &kopsapi.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "testcluster.k8s.local"},
	}

	if len(groups) == 0 {
		groups = make(map[string]*cloudinstances.CloudInstanceGroup)
		groups["master-1"] = &cloudinstances.CloudInstanceGroup{
			InstanceGroup: &kopsapi.InstanceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "master-1",
				},
				Spec: kopsapi.InstanceGroupSpec{
					Role: kopsapi.InstanceGroupRoleMaster,
				},
			},
			MinSize: 1,
			Ready: []*cloudinstances.CloudInstanceGroupMember{
				{
					ID: "i-00001",
					Node: &v1.Node{
						ObjectMeta: metav1.ObjectMeta{Name: "master-1a"},
						Status: v1.NodeStatus{
							Conditions: []v1.NodeCondition{
								{Type: "Ready", Status: v1.ConditionTrue},
							},
						},
					},
				},
			},
		}
	}

	instanceGroups := make([]kopsapi.InstanceGroup, 0, len(groups))
	objects = append([]runtime.Object(nil), objects...)
	for _, g := range groups {
		instanceGroups = append(instanceGroups, *g.InstanceGroup)
		for _, member := range g.Ready {
			node := member.Node
			if node != nil {
				objects = append(objects, node)
			}
		}
		for _, member := range g.NeedUpdate {
			node := member.Node
			if node != nil {
				objects = append(objects, node)
			}
		}
	}

	mockcloud := BuildMockCloud(t, groups, cluster, instanceGroups)

	validator, err := NewClusterValidator(cluster, mockcloud, &kopsapi.InstanceGroupList{Items: instanceGroups}, fake.NewSimpleClientset(objects...))
	if err != nil {
		return nil, err
	}
	return validator.Validate()
}

func Test_ValidateCloudGroupMissing(t *testing.T) {
	cluster := &kopsapi.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "testcluster.k8s.local"},
	}
	instanceGroups := []kopsapi.InstanceGroup{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleNode,
			},
		},
	}

	mockcloud := BuildMockCloud(t, nil, cluster, instanceGroups)

	validator, err := NewClusterValidator(cluster, mockcloud, &kopsapi.InstanceGroupList{Items: instanceGroups}, fake.NewSimpleClientset())
	require.NoError(t, err)
	v, err := validator.Validate()
	require.NoError(t, err)
	if !assert.Len(t, v.Failures, 1) ||
		!assert.Equal(t, &ValidationError{
			Kind:    "InstanceGroup",
			Name:    "node-1",
			Message: "InstanceGroup \"node-1\" is missing from the cloud provider",
		}, v.Failures[0]) {
		printDebug(t, v)
	}
}

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
		MinSize: 3,
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
							{Type: "Ready", Status: v1.ConditionTrue},
						},
					},
				},
			},
		},
	}

	v, err := testValidate(t, groups, nil)
	require.NoError(t, err)
	if !assert.Len(t, v.Failures, 1) ||
		!assert.Equal(t, &ValidationError{
			Kind:    "InstanceGroup",
			Name:    "node-1",
			Message: "InstanceGroup \"node-1\" did not have enough nodes 2 vs 3",
		}, v.Failures[0]) {
		printDebug(t, v)
	}
}

func Test_ValidateDetachedNodesDontCount(t *testing.T) {
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
		MinSize: 2,
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
							{Type: "Ready", Status: v1.ConditionTrue},
						},
					},
				},
				Detached: true,
			},
		},
	}

	v, err := testValidate(t, groups, nil)
	require.NoError(t, err)
	if !assert.Len(t, v.Failures, 1) ||
		!assert.Equal(t, &ValidationError{
			Kind:    "InstanceGroup",
			Name:    "node-1",
			Message: "InstanceGroup \"node-1\" did not have enough nodes 1 vs 2",
		}, v.Failures[0]) {
		printDebug(t, v)
	}
}

func Test_ValidateNodeNotReady(t *testing.T) {
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
		MinSize: 2,
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

	v, err := testValidate(t, groups, nil)
	require.NoError(t, err)
	if !assert.Len(t, v.Failures, 1) ||
		!assert.Equal(t, &ValidationError{
			Kind:    "Node",
			Name:    "node-1b",
			Message: "node \"node-1b\" is not ready",
		}, v.Failures[0]) {
		printDebug(t, v)
	}
}

func Test_ValidateMastersNotEnough(t *testing.T) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["node-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleMaster,
			},
		},
		MinSize: 3,
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID: "i-00001",
				Node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "master-1a"},
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
					ObjectMeta: metav1.ObjectMeta{Name: "master-1b"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: "Ready", Status: v1.ConditionTrue},
						},
					},
				},
			},
		},
	}

	v, err := testValidate(t, groups, nil)
	require.NoError(t, err)
	if !assert.Len(t, v.Failures, 1) ||
		!assert.Equal(t, &ValidationError{
			Kind:    "InstanceGroup",
			Name:    "master-1",
			Message: "InstanceGroup \"master-1\" did not have enough nodes 2 vs 3",
		}, v.Failures[0]) {
		printDebug(t, v)
	}
}

func Test_ValidateMasterNotReady(t *testing.T) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["node-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleMaster,
			},
		},
		MinSize: 2,
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID: "i-00001",
				Node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "master-1a"},
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
					ObjectMeta: metav1.ObjectMeta{Name: "master-1b"},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: "Ready", Status: v1.ConditionFalse},
						},
					},
				},
			},
		},
	}

	v, err := testValidate(t, groups, nil)
	require.NoError(t, err)
	if !assert.Len(t, v.Failures, 1) ||
		!assert.Equal(t, &ValidationError{
			Kind:    "Node",
			Name:    "master-1b",
			Message: "master \"master-1b\" is not ready",
		}, v.Failures[0]) {
		printDebug(t, v)
	}
}

func Test_ValidateMasterStaticPods(t *testing.T) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["node-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleMaster,
			},
		},
		MinSize: 1,
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID: "i-00001",
				Node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "master-1a",
						Labels: map[string]string{"kubernetes.io/role": "master"},
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{
								Address: "1.2.3.4",
							},
						},
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
					ObjectMeta: metav1.ObjectMeta{
						Name:   "master-1b",
						Labels: map[string]string{"kubernetes.io/role": "master"},
					},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: "Ready", Status: v1.ConditionTrue},
						},
						Addresses: []v1.NodeAddress{
							{
								Address: "5.6.7.8",
							},
						},
					},
				},
			},
			{
				ID: "i-00003",
				Node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "master-1c",
						Labels: map[string]string{"kubernetes.io/role": "master"},
					},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{Type: "Ready", Status: v1.ConditionFalse},
						},
						Addresses: []v1.NodeAddress{
							{
								Address: "9.10.11.12",
							},
						},
					},
				},
			},
		},
	}

	var podList []map[string]string
	expectedFailures := []*ValidationError{
		{
			Kind:    "Node",
			Name:    "master-1c",
			Message: "master \"master-1c\" is not ready",
		},
	}

	for i, pod := range []string{
		"kube-apiserver",
		"kube-controller-manager",
		"kube-scheduler",
	} {
		podList = append(podList, []map[string]string{
			{
				"name":              fmt.Sprintf("pod-a-%d", i),
				"ready":             "true",
				"k8s-app":           pod,
				"phase":             string(v1.PodRunning),
				"priorityClassName": "system-cluster-critical",
				"hostip":            "1.2.3.4",
			},
			{
				"name":              fmt.Sprintf("pod-b-%d", i),
				"namespace":         "other",
				"ready":             "true",
				"k8s-app":           pod,
				"phase":             string(v1.PodRunning),
				"priorityClassName": "system-cluster-critical",
				"hostip":            "5.6.7.8",
			},
		}...)
		expectedFailures = append(expectedFailures, &ValidationError{
			Kind:    "Node",
			Name:    "master-1b",
			Message: "master \"master-1b\" is missing " + pod + " pod",
		})
	}

	v, err := testValidate(t, groups, makePodList(podList))
	require.NoError(t, err)
	if !assert.ElementsMatch(t, v.Failures, expectedFailures) {
		printDebug(t, v)
	}
}

func Test_ValidateNoComponentFailures(t *testing.T) {
	v, err := testValidate(t, nil, []runtime.Object{
		&v1.ComponentStatus{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testcomponent",
			},
			Conditions: []v1.ComponentCondition{
				{
					Status: v1.ConditionTrue,
				},
			},
		},
	})

	require.NoError(t, err)
	assert.Empty(t, v.Failures)
}

func Test_ValidateComponentFailure(t *testing.T) {
	for _, status := range []v1.ConditionStatus{
		v1.ConditionFalse,
		v1.ConditionUnknown,
	} {
		t.Run(string(status), func(t *testing.T) {
			v, err := testValidate(t, nil, []runtime.Object{
				&v1.ComponentStatus{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testcomponent",
					},
					Conditions: []v1.ComponentCondition{
						{
							Status: status,
						},
					},
				},
			})

			require.NoError(t, err)
			if !assert.Len(t, v.Failures, 1) ||
				!assert.Equal(t, &ValidationError{
					Kind:    "ComponentStatus",
					Name:    "testcomponent",
					Message: "component \"testcomponent\" is unhealthy",
				}, v.Failures[0]) {
				printDebug(t, v)
			}
		})
	}
}

func Test_ValidateNoPodFailures(t *testing.T) {
	testpods := []map[string]string{}

	for _, phase := range []v1.PodPhase{
		v1.PodPending,
		v1.PodRunning,
		v1.PodSucceeded,
		v1.PodFailed,
		v1.PodUnknown,
	} {
		for _, priority := range []string{"", "otherPriority"} {
			testpods = append(testpods, []map[string]string{
				{
					"name":              fmt.Sprintf("ready-%s-%s", priority, string(phase)),
					"namespace":         "kube-system",
					"priorityClassName": priority,
					"ready":             "true",
					"phase":             string(phase),
				},
				{
					"name":              fmt.Sprintf("notready-%s-%s", priority, string(phase)),
					"namespace":         "kube-system",
					"priorityClassName": priority,
					"ready":             "false",
					"phase":             string(phase),
				},
			}...)
		}
	}

	for _, namespace := range []string{"kube-system", "otherNamespace"} {
		for _, priority := range []string{"node", "cluster"} {
			testpods = append(testpods, []map[string]string{
				{
					"name":              fmt.Sprintf("ready-%s-%s", priority, namespace),
					"namespace":         namespace,
					"priorityClassName": fmt.Sprintf("system-%s-critical", priority),
					"ready":             "true",
					"phase":             string(v1.PodRunning),
				},
				{
					"name":              fmt.Sprintf("notready-%s-%s", priority, namespace),
					"namespace":         namespace,
					"priorityClassName": fmt.Sprintf("system-%s-critical", priority),
					"ready":             "false",
					"phase":             string(v1.PodSucceeded),
				},
			}...)
		}
	}

	v, err := testValidate(t, nil, makePodList(testpods))

	require.NoError(t, err)
	if !assert.Empty(t, v.Failures) {
		printDebug(t, v)
	}
}

func Test_ValidatePodFailure(t *testing.T) {
	for _, tc := range []struct {
		name     string
		phase    v1.PodPhase
		expected string
	}{
		{
			name:     "pending",
			phase:    v1.PodPending,
			expected: "pending",
		},
		{
			name:     "notready",
			phase:    v1.PodRunning,
			expected: "not ready (container1,container2)",
		},
		{
			name:     "unknown",
			phase:    v1.PodUnknown,
			expected: "unknown phase",
		},
	} {
		for _, priority := range []string{"node", "cluster"} {
			for _, namespace := range []string{"kube-system", "otherNamespace"} {
				t.Run(fmt.Sprintf("%s-%s-%s", tc.name, priority, namespace), func(t *testing.T) {
					v, err := testValidate(t, nil, makePodList(
						[]map[string]string{
							{
								"name":              "pod1",
								"namespace":         namespace,
								"priorityClassName": fmt.Sprintf("system-%s-critical", priority),
								"ready":             "false",
								"phase":             string(tc.phase),
							},
						},
					))
					expected := ValidationError{
						Kind:    "Pod",
						Name:    fmt.Sprintf("%s/pod1", namespace),
						Message: fmt.Sprintf("system-%s-critical pod \"pod1\" is %s", priority, tc.expected),
					}

					require.NoError(t, err)
					if !assert.Len(t, v.Failures, 1) ||
						!assert.Equal(t, &expected, v.Failures[0]) {
						printDebug(t, v)
					}
				})
			}
		}
	}
}

func printDebug(t *testing.T, v *ValidationCluster) {
	t.Logf("cluster - %d failures", len(v.Failures))
	for _, fail := range v.Failures {
		t.Logf("  failure: %+v", fail)
	}
}

func dummyPod(podMap map[string]string) v1.Pod {
	var labels map[string]string
	if podMap["k8s-app"] != "" {
		labels = map[string]string{"k8s-app": podMap["k8s-app"]}
	}
	namespace := podMap["namespace"]
	if namespace == "" {
		namespace = "kube-system"
	}
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podMap["name"],
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.PodSpec{
			PriorityClassName: podMap["priorityClassName"],
		},
		Status: v1.PodStatus{
			Phase: v1.PodPhase(podMap["phase"]),
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "container1",
					Ready: podMap["ready"] == "true",
				},
				{
					Name:  "container2",
					Ready: podMap["ready"] == "true",
				},
			},
			HostIP: podMap["hostip"],
		},
	}
}

// MakePodList constructs api.PodList from a list of pod attributes
func makePodList(pods []map[string]string) []runtime.Object {
	var list []runtime.Object
	for _, pod := range pods {
		p := dummyPod(pod)
		list = append(list, &p)
	}
	return list
}

func Test_ValidateBastionNodes(t *testing.T) {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["ig1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ig1",
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
	t.Run("instancegroup's nodes not ready", func(t *testing.T) {
		groups["ig1"].InstanceGroup.Spec.Role = kopsapi.InstanceGroupRoleNode
		v, err := testValidate(t, groups, nil)
		require.NoError(t, err)
		if !assert.Len(t, v.Failures, 1) {
			printDebug(t, v)
		} else if !assert.Equal(t, "machine \"i-00001\" has not yet joined cluster", v.Failures[0].Message) {
			printDebug(t, v)
		}
	})

	// Except for a bastion instancegroup - those are not expected to join as nodes
	t.Run("bastion instancegroup nodes not ready", func(t *testing.T) {
		groups["ig1"].InstanceGroup.Spec.Role = kopsapi.InstanceGroupRoleBastion
		v, err := testValidate(t, groups, nil)
		require.NoError(t, err)
		if !assert.Empty(t, v.Failures, "Bastion nodes are not expected to join cluster") {
			printDebug(t, v)
		}
	})

}
