/*
Copyright 2016 The Kubernetes Authors.

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
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"

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

func TestRollingUpdateAllNeedUpdate(t *testing.T) {
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
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
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

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
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
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		Force:           false,
		K8sClient:       k8sClient,
	}

	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

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

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
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
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		Force:           true,
		K8sClient:       k8sClient,
	}
	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

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

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
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
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		K8sClient:       k8sClient,
		Force:           false,
	}
	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	err := c.RollingUpdate(groups, &kopsapi.Cluster{}, &kopsapi.InstanceGroupList{})
	if err != nil {
		t.Errorf("Error on rolling update: %v", err)
	}

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
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
		Cloud:           mockcloud,
		MasterInterval:  1 * time.Millisecond,
		NodeInterval:    1 * time.Millisecond,
		BastionInterval: 1 * time.Millisecond,
		Force:           false,
		K8sClient:       k8sClient,
	}
	cloud := c.Cloud.(awsup.AWSCloud)
	setUpCloud(c)

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

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

	asgGroups, _ = cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
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
