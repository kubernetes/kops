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
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func getTestSetup() (*RollingUpdateCluster, awsup.AWSCloud, *kopsapi.Cluster) {
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

	return c, mockcloud, cluster
}

func setUpCloud(cloud awsup.AWSCloud) {
	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("node-1"),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("node-1"),
		InstanceIds:          []*string{aws.String("node-1a"), aws.String("node-1b")},
	})

	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("node-2"),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("node-2"),
		InstanceIds:          []*string{aws.String("node-2a"), aws.String("node-2b")},
	})

	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("master-1"),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("master-1"),
		InstanceIds:          []*string{aws.String("master-1a"), aws.String("master-1b")},
	})

	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("bastion-1"),
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

func getGroups() map[string]*cloudinstances.CloudInstanceGroup {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["node-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: "node-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleNode,
			},
		},
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "node-1a",
				Node: &v1.Node{},
			},
			{
				ID:   "node-1b",
				Node: &v1.Node{},
			},
		},
	}
	groups["node-2"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: "node-2",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleNode,
			},
		},
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "node-2a",
				Node: &v1.Node{},
			},
			{
				ID:   "node-2b",
				Node: &v1.Node{},
			},
		},
	}
	groups["master-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: "master-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleMaster,
			},
		},
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "master-1a",
				Node: &v1.Node{},
			},
			{
				ID:   "master-1b",
				Node: &v1.Node{},
			},
		},
	}
	groups["bastion-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: "bastion-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: kopsapi.InstanceGroupRoleBastion,
			},
		},
		Ready: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "bastion-1a",
				Node: &v1.Node{},
			},
		},
	}
	return groups
}

func markNeedUpdate(group *cloudinstances.CloudInstanceGroup, nodeIds ...string) {
	for _, nodeId := range nodeIds {
		found := false
		for _, member := range group.Ready {
			if member.ID == nodeId {
				group.NeedUpdate = append(group.NeedUpdate, &cloudinstances.CloudInstanceGroupMember{
					ID:                 member.ID,
					Node:               member.Node,
					CloudInstanceGroup: member.CloudInstanceGroup,
				})
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Sprintf("didn't find nodeId %s in ready list", nodeId))
		}
	}
}

func getGroupsNodes1NeedsUpdate() map[string]*cloudinstances.CloudInstanceGroup {
	groups := getGroups()
	markNeedUpdate(groups["node-1"], "node-1a", "node-1b")
	return groups
}

func getGroupsAllNeedUpdate() map[string]*cloudinstances.CloudInstanceGroup {
	groups := getGroups()
	markNeedUpdate(groups["node-1"], "node-1a", "node-1b")
	markNeedUpdate(groups["node-2"], "node-2a", "node-2b")
	markNeedUpdate(groups["master-1"], "master-1a", "master-1b")
	markNeedUpdate(groups["bastion-1"], "bastion-1a")
	return groups
}

func TestRollingUpdateAllNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateAllNeedUpdateCloudonly(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.CloudOnly = true
	c.ClusterValidator = &assertNotCalledClusterValidator{T: t}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateAllNeedUpdateNoFailOnValidate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.FailOnValidate = false
	c.ClusterValidator = &failingClusterValidator{}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateNoneNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	err := c.RollingUpdate(getGroups(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateNoneNeedUpdateWithForce(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.Force = true

	err := c.RollingUpdate(getGroups(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateEmptyGroup(t *testing.T) {
	c, cloud, _ := getTestSetup()

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	err := c.RollingUpdate(groups, &kopsapi.Cluster{}, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateUnknownRole(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	groups := getGroups()
	groups["node-1"].InstanceGroup.Spec.Role = "Unknown"

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateAllNeedUpdateFailsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateAllNeedUpdateErrorsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateNodes1NeedsUpdateFailsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
}

func TestRollingUpdateNodes1NeedsUpdateErrorsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
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
		if len(group.Instances) < 2 {
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
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "master-1",
		ReturnError: false,
	}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 1)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateClusterErrorsValidationAfterOneMaster(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "master-1",
		ReturnError: true,
	}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 1)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateClusterFailsValidationAfterOneNode(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "node-1",
		ReturnError: false,
	}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 1)
}

func TestRollingUpdateClusterErrorsValidationAfterOneNode(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "node-1",
		ReturnError: true,
	}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 1)
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
	c, cloud, cluster := getTestSetup()

	// This should only take a few milliseconds,
	// but we have to pad to allow for random delays (e.g. GC)
	// TODO: Replace with a virtual clock?
	c.ValidationTimeout = 1 * time.Second

	c.ClusterValidator = &flappingClusterValidator{
		T:     t,
		Cloud: cloud,
	}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
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
	c, cloud, cluster := getTestSetup()

	// This should only take a few milliseconds,
	// but we have to pad to allow for random delays (e.g. GC)
	// TODO: Replace with a virtual clock?
	c.ValidationTimeout = 1 * time.Second

	c.ClusterValidator = &failThreeTimesClusterValidator{}

	err := c.RollingUpdate(getGroupsAllNeedUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 0)
	assertGroupInstanceCount(t, cloud, "node-2", 0)
	assertGroupInstanceCount(t, cloud, "master-1", 0)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func assertGroupInstanceCount(t *testing.T, cloud awsup.AWSCloud, groupName string, expected int) {
	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(groupName)},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Lenf(t, group.Instances, expected, "%s instances", groupName)
	}
}
