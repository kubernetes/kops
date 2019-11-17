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

	if len(instanceGroups) == 0 {
		instanceGroups = []kopsapi.InstanceGroup{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "master-1",
				},
				Spec: kopsapi.InstanceGroupSpec{
					Role: kopsapi.InstanceGroupRoleMaster,
				},
			},
		}
	}

	mockcloud := BuildMockCloud(t, groups, cluster, instanceGroups)

	validator, err := NewClusterValidator(cluster, mockcloud, &kopsapi.InstanceGroupList{Items: instanceGroups}, fake.NewSimpleClientset(objects...))
	if err != nil {
		return nil, err
	}
	return validator.Validate()
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

	t.Run("too few nodes", func(t *testing.T) {
		groups["node-1"].MinSize = 3
		v, err := testValidate(t, groups, nil)
		require.NoError(t, err)
		if !assert.Len(t, v.Failures, 2) {
			printDebug(t, v)
		}
	})

	t.Run("not ready node", func(t *testing.T) {
		groups["node-1"].MinSize = 2
		v, err := testValidate(t, groups, nil)
		require.NoError(t, err)
		if !assert.Len(t, v.Failures, 1) {
			printDebug(t, v)
		}
	})

	t.Run("unexpected errors", func(t *testing.T) {
		groups["node-1"].NeedUpdate[0].Node.Status.Conditions[0].Status = v1.ConditionTrue
		v, err := testValidate(t, groups, nil)
		require.NoError(t, err)
		if !assert.Empty(t, v.Failures) {
			printDebug(t, v)
		}
	})
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
	v, err := testValidate(t, nil, makePodList(
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

	require.NoError(t, err)
	assert.Empty(t, v.Failures)
}

func Test_ValidatePodFailure(t *testing.T) {
	for _, tc := range []struct {
		name     string
		phase    v1.PodPhase
		expected ValidationError
	}{
		{
			name:  "pending",
			phase: v1.PodPending,
			expected: ValidationError{
				Kind:    "Pod",
				Name:    "kube-system/pod1",
				Message: "kube-system pod \"pod1\" is pending",
			},
		},
		{
			name:  "notready",
			phase: v1.PodRunning,
			expected: ValidationError{
				Kind:    "Pod",
				Name:    "kube-system/pod1",
				Message: "kube-system pod \"pod1\" is not ready (container1,container2)",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := testValidate(t, nil, makePodList(
				[]map[string]string{
					{
						"name":  "pod1",
						"ready": "false",
						"phase": string(tc.phase),
					},
				},
			))

			require.NoError(t, err)
			if !assert.Len(t, v.Failures, 1) ||
				!assert.Equal(t, &tc.expected, v.Failures[0]) {
				printDebug(t, v)
			}
		})
	}
}

func printDebug(t *testing.T, v *ValidationCluster) {
	t.Logf("cluster - %d failures", len(v.Failures))
	for _, fail := range v.Failures {
		t.Logf("  failure: %+v", fail)
	}
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
					Name:  "container1",
					Ready: podMap["ready"] == "true",
				},
				{
					Name:  "container2",
					Ready: podMap["ready"] == "true",
				},
			},
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
