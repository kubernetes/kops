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
		Cloud:            mockcloud,
		MasterInterval:   1 * time.Millisecond,
		NodeInterval:     1 * time.Millisecond,
		BastionInterval:  1 * time.Millisecond,
		Force:            false,
		K8sClient:        k8sClient,
		ClusterValidator: &successfulClusterValidator{},
		FailOnValidate:   true,
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
		InstanceIds:          []*string{aws.String("master-1a")},
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
	markNeedUpdate(groups["master-1"], "master-1a")
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

func TestRollingUpdateNoneNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	err := c.RollingUpdate(getGroups(), cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
	assertGroupInstanceCount(t, cloud, "node-2", 2)
	assertGroupInstanceCount(t, cloud, "master-1", 1)
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
	assertGroupInstanceCount(t, cloud, "master-1", 1)
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
	assertGroupInstanceCount(t, cloud, "master-1", 1)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateClusterFailsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
}

func TestRollingUpdateClusterErrorsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
}

type failAfterOneNodeClusterValidator struct {
	Cloud       awsup.AWSCloud
	ReturnError bool
}

func (v *failAfterOneNodeClusterValidator) Validate() (*validation.ValidationCluster, error) {
	asgGroups, _ := v.Cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
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

func TestRollingUpdateClusterFailsValidationAfterOneNode(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
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
		ReturnError: true,
	}

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdate(), cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 1)
}

func assertGroupInstanceCount(t *testing.T, cloud awsup.AWSCloud, groupName string, expected int) {
	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(groupName)},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Lenf(t, group.Instances, expected, "%s instances", groupName)
	}
}
