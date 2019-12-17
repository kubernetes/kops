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
	"testing"
	"time"

	"k8s.io/kops/pkg/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func setUpCloud(c *RollingUpdateCluster) {
	cloud := c.Cloud.(awsup.AWSCloud)
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

func TestRollingUpdateAllNeedUpdate(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

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

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "bastion-1a",
				Node: &v1.Node{},
			},
		},
	}

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	if err != nil {
		t.Errorf("Error on rolling update: %v", err)
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) > 0 {
			t.Error("Not all instances terminated")
		}
	}
}

func TestRollingUpdateNoneNeedUpdate(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

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

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

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

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	if err != nil {
		t.Errorf("Error on rolling update: %v", err)
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected 2 instances got: %v in %v", len(group.Instances), group)
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-2")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected 2 instances got: %v in %v", len(group.Instances), group)
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("master-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected 1 instance got: %v in %v", len(group.Instances), group)
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("bastion-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected 1 instance got: %v in %v", len(group.Instances), group)
		}
	}
}

func TestRollingUpdateNoneNeedUpdateWithForce(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:            mockcloud,
		MasterInterval:   1 * time.Millisecond,
		NodeInterval:     1 * time.Millisecond,
		BastionInterval:  1 * time.Millisecond,
		Force:            true,
		K8sClient:        k8sClient,
		ClusterValidator: &successfulClusterValidator{},
		FailOnValidate:   true,
	}
	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

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

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	if err != nil {
		t.Errorf("Error on rolling update: %v", err)
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) > 0 {
			t.Error("Not all instances terminated")
		}
	}
}

func TestRollingUpdateEmptyGroup(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	c := &RollingUpdateCluster{
		Cloud:            mockcloud,
		MasterInterval:   1 * time.Millisecond,
		NodeInterval:     1 * time.Millisecond,
		BastionInterval:  1 * time.Millisecond,
		K8sClient:        k8sClient,
		ClusterValidator: &successfulClusterValidator{},
		Force:            false,
		FailOnValidate:   true,
	}
	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	err := c.RollingUpdate(groups, &kopsapi.Cluster{}, &kopsapi.InstanceGroupList{})
	if err != nil {
		t.Errorf("Error on rolling update: %v", err)
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected 2 instances got: %v", len(group.Instances))
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-2")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected 2 instances got: %v", len(group.Instances))
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("master-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected 1 instances got: %v", len(group.Instances))
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("bastion-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected 1 instances got: %v", len(group.Instances))
		}
	}
}

func TestRollingUpdateUnknownRole(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

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
	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	groups["node-1"] = &cloudinstances.CloudInstanceGroup{
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: "node-1",
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: "Unknown",
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

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	if err == nil {
		t.Errorf("Error expected on rolling update: %v", err)
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected 2 instances got: %v", len(group.Instances))
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-2")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected 2 instances got: %v", len(group.Instances))
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("master-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected 1 instances got: %v", len(group.Instances))
		}
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("bastion-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected 1 instances got: %v", len(group.Instances))
		}
	}
}

func getGroupsNodes1NeedsUpdating() map[string]*cloudinstances.CloudInstanceGroup {
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
	return groups
}

func TestRollingUpdateClusterFailsValidation(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:            mockcloud,
		MasterInterval:   1 * time.Millisecond,
		NodeInterval:     1 * time.Millisecond,
		BastionInterval:  1 * time.Millisecond,
		Force:            false,
		K8sClient:        k8sClient,
		ClusterValidator: &failingClusterValidator{},
		FailOnValidate:   true,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdating(), cluster, &kopsapi.InstanceGroupList{})
	if err == nil {
		t.Error("Expected error from rolling update, got nil")
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected two instances in group got %v", len(group.Instances))
		}
	}
}

func TestRollingUpdateClusterErrorsValidation(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:            mockcloud,
		MasterInterval:   1 * time.Millisecond,
		NodeInterval:     1 * time.Millisecond,
		BastionInterval:  1 * time.Millisecond,
		Force:            false,
		K8sClient:        k8sClient,
		ClusterValidator: &erroringClusterValidator{},
		FailOnValidate:   true,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdating(), cluster, &kopsapi.InstanceGroupList{})
	if err == nil {
		t.Error("Expected error from rolling update, got nil")
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 2 {
			t.Errorf("Expected two instances in group got %v", len(group.Instances))
		}
	}
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
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		Force:           false,
		K8sClient:       k8sClient,
		ClusterValidator: &failAfterOneNodeClusterValidator{
			Cloud:       mockcloud,
			ReturnError: false,
		},
		FailOnValidate: true,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdating(), cluster, &kopsapi.InstanceGroupList{})
	if err == nil {
		t.Error("Expected error from rolling update, got nil")
	}

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected one instance in group got %v", len(group.Instances))
		}
	}
}

func TestRollingUpdateClusterErrorsValidationAfterOneNode(t *testing.T) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		Force:           false,
		K8sClient:       k8sClient,
		ClusterValidator: &failAfterOneNodeClusterValidator{
			Cloud:       mockcloud,
			ReturnError: true,
		},
		FailOnValidate: true,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	err := c.RollingUpdate(getGroupsNodes1NeedsUpdating(), cluster, &kopsapi.InstanceGroupList{})
	if err == nil {
		t.Error("Expected error from rolling update, got nil")
	}
	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String("node-1")},
	})
	for _, group := range asgGroups.AutoScalingGroups {
		if len(group.Instances) != 1 {
			t.Errorf("Expected one instance in group got %v", len(group.Instances))
		}
	}
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
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		Force:           false,
		K8sClient:       k8sClient,
		ClusterValidator: &flappingClusterValidator{
			T:     t,
			Cloud: mockcloud,
		},
		FailOnValidate:          true,
		ValidationTimeout:       200 * time.Second,
		ValidateTickDuration:    1 * time.Millisecond,
		ValidateSuccessDuration: 5 * time.Millisecond,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)
	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String("master-1"),
		InstanceIds:          []*string{aws.String("master-1b")},
	})

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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "bastion-1a",
				Node: &v1.Node{},
			},
		},
	}

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
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockcloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{}

	cluster := &kopsapi.Cluster{}
	cluster.Name = "test.k8s.local"

	c := &RollingUpdateCluster{
		Cloud:                mockcloud,
		MasterInterval:       1 * time.Millisecond,
		NodeInterval:         1 * time.Millisecond,
		BastionInterval:      1 * time.Millisecond,
		Force:                false,
		K8sClient:            k8sClient,
		ClusterValidator:     &failThreeTimesClusterValidator{},
		FailOnValidate:       true,
		ValidationTimeout:    1 * time.Second,
		ValidateTickDuration: 1 * time.Millisecond,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
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
		NeedUpdate: []*cloudinstances.CloudInstanceGroupMember{
			{
				ID:   "bastion-1a",
				Node: &v1.Node{},
			},
		},
	}

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
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
