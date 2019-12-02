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

package instancegroups

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	testingclient "k8s.io/client-go/testing"
	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

const (
	cordonPatch = "{\"spec\":{\"unschedulable\":true}}"
	taintPatch  = "{\"spec\":{\"taints\":[{\"effect\":\"PreferNoSchedule\",\"key\":\"kops.k8s.io/scheduled-for-update\"}]}}"
)

func getTestSetup() (*RollingUpdateCluster, awsup.AWSCloud, *kopsapi.Cluster, map[string]*cloudinstances.CloudInstanceGroup) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}
	setUpCloud(mockcloud)

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:                   mockcloud,
		MasterInterval:          1 * time.Millisecond,
		NodeInterval:            1 * time.Millisecond,
		BastionInterval:         1 * time.Millisecond,
		Force:                   false,
		K8sClient:               k8sClient,
		ClusterValidator:        &successfulClusterValidator{},
		FailOnValidate:          true,
		ValidateTickDuration:    1 * time.Millisecond,
		ValidateSuccessDuration: 5 * time.Millisecond,
	}

	return c, mockcloud, cluster, getGroups(k8sClient)
}

func setUpCloud(cloud awsup.AWSCloud) {
	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("node-1"),
		DesiredCapacity:      aws.Int64(3),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("node-1"),
		InstanceIds:          []*string{aws.String("node-1a"), aws.String("node-1b"), aws.String("node-1c")},
	})

	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("node-2"),
		DesiredCapacity:      aws.Int64(3),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("node-2"),
		InstanceIds:          []*string{aws.String("node-2a"), aws.String("node-2b"), aws.String("node-2c")},
	})

	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("master-1"),
		DesiredCapacity:      aws.Int64(2),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("master-1"),
		InstanceIds:          []*string{aws.String("master-1a"), aws.String("master-1b")},
	})

	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("bastion-1"),
		DesiredCapacity:      aws.Int64(1),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("bastion-1"),
		InstanceIds:          []*string{aws.String("bastion-1a")},
	})
}

type successfulClusterValidator struct{}

func (*successfulClusterValidator) Validate() (*validation.ValidationCluster, error) {
	return &validation.ValidationCluster{}, nil
}

type failingClusterValidator struct{}

func (*failingClusterValidator) Validate() (*validation.ValidationCluster, error) {
	return &validation.ValidationCluster{
		Failures: []*validation.ValidationError{
			{
				Kind:    "testing",
				Name:    "testingfailure",
				Message: "testing failure",
			},
		},
	}, nil
}

type erroringClusterValidator struct{}

func (*erroringClusterValidator) Validate() (*validation.ValidationCluster, error) {
	return nil, errors.New("testing validation error")
}

type assertNotCalledClusterValidator struct {
	T *testing.T
}

func (v *assertNotCalledClusterValidator) Validate() (*validation.ValidationCluster, error) {
	v.T.Fatal("validator called unexpectedly")
	return nil, errors.New("validator called unexpectedly")
}

func makeGroup(groups map[string]*cloudinstances.CloudInstanceGroup, k8sClient *fake.Clientset, name string, role kopsapi.InstanceGroupRole, count int) {
	groups[name] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: name,
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: role,
			},
		},
	}
	for i := 0; i < count; i++ {
		id := name + string('a'+i)
		var node *v1.Node
		if role != kopsapi.InstanceGroupRoleBastion {
			node = &v1.Node{
				ObjectMeta: v1meta.ObjectMeta{Name: id + ".local"},
			}
			_ = k8sClient.Tracker().Add(node)
		}
		groups[name].Ready = append(groups[name].Ready, &cloudinstances.CloudInstanceGroupMember{
			ID:   id,
			Node: node,
		})
	}
}

func getGroups(k8sClient *fake.Clientset) map[string]*cloudinstances.CloudInstanceGroup {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, k8sClient, "node-1", kopsapi.InstanceGroupRoleNode, 3)
	makeGroup(groups, k8sClient, "node-2", kopsapi.InstanceGroupRoleNode, 3)
	makeGroup(groups, k8sClient, "master-1", kopsapi.InstanceGroupRoleMaster, 2)
	makeGroup(groups, k8sClient, "bastion-1", kopsapi.InstanceGroupRoleBastion, 1)
	return groups
}

func markNeedUpdate(group *cloudinstances.CloudInstanceGroup, nodeIds ...string) {
	for _, nodeId := range nodeIds {
		var newReady []*cloudinstances.CloudInstanceGroupMember
		found := false
		for _, member := range group.Ready {
			if member.ID == nodeId {
				group.NeedUpdate = append(group.NeedUpdate, member)
				found = true
			} else {
				newReady = append(newReady, member)
			}
		}
		group.Ready = newReady
		if !found {
			panic(fmt.Sprintf("didn't find nodeId %s in ready list", nodeId))
		}
	}
}

func markAllNeedUpdate(groups map[string]*cloudinstances.CloudInstanceGroup) {
	markNeedUpdate(groups["node-1"], "node-1a", "node-1b", "node-1c")
	markNeedUpdate(groups["node-2"], "node-2a", "node-2b", "node-2c")
	markNeedUpdate(groups["master-1"], "master-1a", "master-1b")
	markNeedUpdate(groups["bastion-1"], "bastion-1a")
}

func TestRollingUpdateAllNeedUpdate(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	cordoned := ""
	tainted := map[string]bool{}
	deleted := map[string]bool{}
	for _, action := range c.K8sClient.(*fake.Clientset).Actions() {
		switch a := action.(type) {
		case testingclient.PatchAction:
			if string(a.GetPatch()) == cordonPatch {
				assertCordon(t, a)
				assert.Equal(t, "", cordoned, "at most one node cordoned at a time")
				assert.True(t, tainted[a.GetName()], "node", a.GetName(), "tainted")
				cordoned = a.GetName()
			} else {
				assertTaint(t, a)
				assert.Equal(t, "", cordoned, "not tainting while node cordoned")
				assert.False(t, tainted[a.GetName()], "node", a.GetName(), "already tainted")
				tainted[a.GetName()] = true
			}
		case testingclient.DeleteAction:
			assert.Equal(t, "nodes", a.GetResource().Resource)
			assert.Equal(t, cordoned, a.GetName(), "node was cordoned before delete")
			assert.False(t, deleted[a.GetName()], "node", a.GetName(), "already deleted")
			if !strings.HasPrefix(a.GetName(), "master-") {
				assert.True(t, deleted["master-1a.local"], "master-1a was deleted before node", a.GetName())
				assert.True(t, deleted["master-1b.local"], "master-1b was deleted before node", a.GetName())
			}
			deleted[a.GetName()] = true
			cordoned = ""
		case testingclient.ListAction:
			// Don't care
		default:
			t.Errorf("unexpected action %v", a)
		}
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateAllNeedUpdateCloudonly(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.CloudOnly = true
	c.ClusterValidator = &assertNotCalledClusterValidator{T: t}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assert.Empty(t, c.K8sClient.(*fake.Clientset).Actions())

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateAllNeedUpdateNoFailOnValidate(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.FailOnValidate = false
	c.ClusterValidator = &failingClusterValidator{}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateNoneNeedUpdate(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assert.Empty(t, c.K8sClient.(*fake.Clientset).Actions())

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateNoneNeedUpdateWithForce(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.Force = true

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateEmptyGroup(t *testing.T) {
	c, cloud, _, _ := getTestSetup()

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	err := c.RollingUpdate(groups, &kopsapi.Cluster{}, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateUnknownRole(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	groups["node-1"].InstanceGroup.Spec.Role = "Unknown"

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateAllNeedUpdateFailsValidation(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateAllNeedUpdateErrorsValidation(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateNodes1NeedsUpdateFailsValidation(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	markNeedUpdate(groups["node-1"], "node-1a", "node-1b", "node-1c")
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
}

func TestRollingUpdateNodes1NeedsUpdateErrorsValidation(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	markNeedUpdate(groups["node-1"], "node-1a", "node-1b", "node-1c")
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
}

type failAfterOneNodeClusterValidator struct {
	Cloud       awsup.AWSCloud
	Group       string
	ReturnError bool
}

func (v *failAfterOneNodeClusterValidator) Validate() (*validation.ValidationCluster, error) {
	asgGroups, _ := v.Cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(v.Group)},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if int64(len(group.Instances)) < *group.DesiredCapacity {
			if v.ReturnError {
				return nil, errors.New("testing validation error")
			}
			return &validation.ValidationCluster{
				Failures: []*validation.ValidationError{
					{
						Kind:    "testing",
						Name:    "testingfailure",
						Message: "testing failure",
					},
				},
			}, nil
		}
	}
	return &validation.ValidationCluster{}, nil
}

func TestRollingUpdateClusterFailsValidationAfterOneMaster(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "master-1",
		ReturnError: false,
	}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 1)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateClusterErrorsValidationAfterOneMaster(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "master-1",
		ReturnError: true,
	}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 1)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateClusterFailsValidationAfterOneNode(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "node-1",
		ReturnError: false,
	}

	markNeedUpdate(groups["node-1"], "node-1a", "node-1b", "node-1c")
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
}

func TestRollingUpdateClusterErrorsValidationAfterOneNode(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "node-1",
		ReturnError: true,
	}

	markNeedUpdate(groups["node-1"], "node-1a", "node-1b", "node-1c")
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
}

type flappingClusterValidator struct {
	T               *testing.T
	Cloud           awsup.AWSCloud
	invocationCount int
}

func (v *flappingClusterValidator) Validate() (*validation.ValidationCluster, error) {
	asgGroups, _ := v.Cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("master-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		switch len(group.Instances) {
		case 1:
			return &validation.ValidationCluster{}, nil
		case 0:
			assert.GreaterOrEqual(v.T, v.invocationCount, 7, "validator invocation count")
		}
	}

	v.invocationCount++
	switch v.invocationCount {
	case 1, 3, 5:
		return &validation.ValidationCluster{
			Failures: []*validation.ValidationError{
				{
					Kind:    "testing",
					Name:    "testingfailure",
					Message: "testing failure",
				},
			},
		}, nil
	}
	return &validation.ValidationCluster{}, nil
}

func TestRollingUpdateFlappingValidation(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	// This should only take a few milliseconds,
	// but we have to pad to allow for random delays (e.g. GC)
	// TODO: Replace with a virtual clock?
	c.ValidationTimeout = 1 * time.Second

	c.ClusterValidator = &flappingClusterValidator{
		T:     t,
		Cloud: cloud,
	}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 0)
	assertGroupInstanceCount(t, cloud, "node-2", 0)
	assertGroupInstanceCount(t, cloud, "master-1", 0)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

type failThreeTimesClusterValidator struct {
	invocationCount int
}

func (v *failThreeTimesClusterValidator) Validate() (*validation.ValidationCluster, error) {
	v.invocationCount++
	if v.invocationCount <= 3 {
		return &validation.ValidationCluster{
			Failures: []*validation.ValidationError{
				{
					Kind:    "testing",
					Name:    "testingfailure",
					Message: "testing failure",
				},
			},
		}, nil
	}
	return &validation.ValidationCluster{}, nil
}

func TestRollingUpdateValidatesAfterBastion(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	// This should only take a few milliseconds,
	// but we have to pad to allow for random delays (e.g. GC)
	// TODO: Replace with a virtual clock?
	c.ValidationTimeout = 1 * time.Second

	c.ClusterValidator = &failThreeTimesClusterValidator{}

	markAllNeedUpdate(groups)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 0)
	assertGroupInstanceCount(t, cloud, "node-2", 0)
	assertGroupInstanceCount(t, cloud, "master-1", 0)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateTaintAllButOneNeedUpdate(t *testing.T) {
	c, cloud, cluster, groups := getTestSetup()

	markNeedUpdate(groups["node-1"], "node-1a", "node-1b")
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	cordoned := ""
	tainted := map[string]bool{}
	deleted := map[string]bool{}
	for _, action := range c.K8sClient.(*fake.Clientset).Actions() {
		switch a := action.(type) {
		case testingclient.PatchAction:
			if string(a.GetPatch()) == cordonPatch {
				assertCordon(t, a)
				assert.Equal(t, "", cordoned, "at most one node cordoned at a time")
				cordoned = a.GetName()
			} else {
				assertTaint(t, a)
				assert.False(t, tainted[a.GetName()], "node", a.GetName(), "already tainted")
				tainted[a.GetName()] = true
			}
		case testingclient.DeleteAction:
			assert.Equal(t, "nodes", a.GetResource().Resource)
			assert.Equal(t, cordoned, a.GetName(), "node was cordoned before delete")
			assert.Len(t, tainted, 2, "all nodes tainted before any delete")
			assert.False(t, deleted[a.GetName()], "node", a.GetName(), "already deleted")
			deleted[a.GetName()] = true
			cordoned = ""
		case testingclient.ListAction:
			// Don't care
		default:
			t.Errorf("unexpected action %v", a)
		}
	}

	assertGroupInstanceCount(t, cloud, "node-1", 1)
}

func assertCordon(t *testing.T, action testingclient.PatchAction) {
	assert.Equal(t, "nodes", action.GetResource().Resource)
	assert.Equal(t, cordonPatch, string(action.GetPatch()))
}

func assertTaint(t *testing.T, action testingclient.PatchAction) {
	assert.Equal(t, "nodes", action.GetResource().Resource)
	assert.Equal(t, taintPatch, string(action.GetPatch()))
}

func assertGroupInstanceCount(t *testing.T, cloud awsup.AWSCloud, groupName string, expected int) {
	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(groupName)},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Lenf(t, group.Instances, expected, "%s instances", groupName)
	}
}
