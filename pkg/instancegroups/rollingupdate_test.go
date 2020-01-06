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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
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

func getTestSetup() (*RollingUpdateCluster, *awsup.MockAWSCloud, *kopsapi.Cluster) {
	k8sClient := fake.NewSimpleClientset()

	mockcloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockAutoscaling := &mockautoscaling.MockAutoscaling{}
	mockcloud.MockAutoscaling = mockAutoscaling
	mockcloud.MockEC2 = mockAutoscaling.GetEC2Shim(mockcloud.MockEC2)

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

func makeGroup(groups map[string]*cloudinstances.CloudInstanceGroup, k8sClient kubernetes.Interface, cloud awsup.AWSCloud, name string, role kopsapi.InstanceGroupRole, count int, needUpdate int) {
	fakeClient := k8sClient.(*fake.Clientset)

	groups[name] = &cloudinstances.CloudInstanceGroup{
		HumanName: name,
		InstanceGroup: &kopsapi.InstanceGroup{
			ObjectMeta: v1meta.ObjectMeta{
				Name: name,
			},
			Spec: kopsapi.InstanceGroupSpec{
				Role: role,
			},
		},
		Raw: &autoscaling.Group{AutoScalingGroupName: aws.String("asg-" + name)},
	}
	cloud.Autoscaling().CreateAutoScalingGroup(&autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(name),
		DesiredCapacity:      aws.Int64(int64(count)),
		MinSize:              aws.Int64(1),
		MaxSize:              aws.Int64(5),
	})

	var instanceIds []*string
	for i := 0; i < count; i++ {
		id := name + string('a'+i)
		var node *v1.Node
		if role != kopsapi.InstanceGroupRoleBastion {
			node = &v1.Node{
				ObjectMeta: v1meta.ObjectMeta{Name: id + ".local"},
			}
			_ = fakeClient.Tracker().Add(node)
		}
		member := cloudinstances.CloudInstanceGroupMember{
			ID:                 id,
			Node:               node,
			CloudInstanceGroup: groups[name],
		}
		if i < needUpdate {
			groups[name].NeedUpdate = append(groups[name].NeedUpdate, &member)
		} else {
			groups[name].Ready = append(groups[name].Ready, &member)
		}
		instanceIds = append(instanceIds, aws.String(id))
	}
	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String(name),
		InstanceIds:          instanceIds,
	})
}

func getGroups(k8sClient kubernetes.Interface, cloud awsup.AWSCloud) map[string]*cloudinstances.CloudInstanceGroup {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, k8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 0)
	makeGroup(groups, k8sClient, cloud, "node-2", kopsapi.InstanceGroupRoleNode, 3, 0)
	makeGroup(groups, k8sClient, cloud, "master-1", kopsapi.InstanceGroupRoleMaster, 2, 0)
	makeGroup(groups, k8sClient, cloud, "bastion-1", kopsapi.InstanceGroupRoleBastion, 1, 0)
	return groups
}

func getGroupsAllNeedUpdate(k8sClient kubernetes.Interface, cloud awsup.AWSCloud) map[string]*cloudinstances.CloudInstanceGroup {
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, k8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 3)
	makeGroup(groups, k8sClient, cloud, "node-2", kopsapi.InstanceGroupRoleNode, 3, 3)
	makeGroup(groups, k8sClient, cloud, "master-1", kopsapi.InstanceGroupRoleMaster, 2, 2)
	makeGroup(groups, k8sClient, cloud, "bastion-1", kopsapi.InstanceGroupRoleBastion, 1, 1)
	return groups
}

func TestRollingUpdateAllNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
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
	c, cloud, cluster := getTestSetup()

	c.CloudOnly = true
	c.ClusterValidator = &assertNotCalledClusterValidator{T: t}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assert.Empty(t, c.K8sClient.(*fake.Clientset).Actions())

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateAllNeedUpdateNoFailOnValidate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.FailOnValidate = false
	c.ClusterValidator = &failingClusterValidator{}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	asgGroups, _ := cloud.Autoscaling().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	for _, group := range asgGroups.AutoScalingGroups {
		assert.Emptyf(t, group.Instances, "Not all instances terminated in group %s", group.AutoScalingGroupName)
	}
}

func TestRollingUpdateNoneNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()
	groups := getGroups(c.K8sClient, cloud)

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assert.Empty(t, c.K8sClient.(*fake.Clientset).Actions())

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateNoneNeedUpdateWithForce(t *testing.T) {
	c, cloud, cluster := getTestSetup()
	groups := getGroups(c.K8sClient, cloud)

	c.Force = true

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
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

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateUnknownRole(t *testing.T) {
	c, cloud, cluster := getTestSetup()
	groups := getGroups(c.K8sClient, cloud)

	groups["node-1"].InstanceGroup.Spec.Role = "Unknown"

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateAllNeedUpdateFailsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateAllNeedUpdateErrorsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateNodes1NeedsUpdateFailsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failingClusterValidator{}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 3)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
}

func TestRollingUpdateNodes1NeedsUpdateErrorsValidation(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &erroringClusterValidator{}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 3)
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
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "master-1",
		ReturnError: false,
	}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
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

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
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

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 3)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.Error(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 2)
}

func TestRollingUpdateClusterErrorsValidationAfterOneNode(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	c.ClusterValidator = &failAfterOneNodeClusterValidator{
		Cloud:       cloud,
		Group:       "node-1",
		ReturnError: true,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 3)
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
	c, cloud, cluster := getTestSetup()

	// This should only take a few milliseconds,
	// but we have to pad to allow for random delays (e.g. GC)
	// TODO: Replace with a virtual clock?
	c.ValidationTimeout = 1 * time.Second

	c.ClusterValidator = &flappingClusterValidator{
		T:     t,
		Cloud: cloud,
	}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
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
	c, cloud, cluster := getTestSetup()

	// This should only take a few milliseconds,
	// but we have to pad to allow for random delays (e.g. GC)
	// TODO: Replace with a virtual clock?
	c.ValidationTimeout = 1 * time.Second

	c.ClusterValidator = &failThreeTimesClusterValidator{}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 0)
	assertGroupInstanceCount(t, cloud, "node-2", 0)
	assertGroupInstanceCount(t, cloud, "master-1", 0)
	assertGroupInstanceCount(t, cloud, "bastion-1", 0)
}

func TestRollingUpdateTaintAllButOneNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 2)
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

func TestRollingUpdateMaxSurgeIgnoredForMaster(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	two := intstr.FromInt(2)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxSurge: &two,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "master-1", kopsapi.InstanceGroupRoleMaster, 3, 2)
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
			deleted[a.GetName()] = true
			cordoned = ""
		case testingclient.ListAction:
			// Don't care
		default:
			t.Errorf("unexpected action %v", a)
		}
	}

	assertGroupInstanceCount(t, cloud, "master-1", 1)
}

func TestRollingUpdateDisabled(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	zero := intstr.FromInt(0)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxUnavailable: &zero,
	}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

func TestRollingUpdateDisabledCloudonly(t *testing.T) {
	c, cloud, cluster := getTestSetup()
	c.CloudOnly = true

	zero := intstr.FromInt(0)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxUnavailable: &zero,
	}

	groups := getGroupsAllNeedUpdate(c.K8sClient, cloud)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 3)
	assertGroupInstanceCount(t, cloud, "node-2", 3)
	assertGroupInstanceCount(t, cloud, "master-1", 2)
	assertGroupInstanceCount(t, cloud, "bastion-1", 1)
}

// The concurrent update tests attempt to induce the following expected update sequence:
//
// (Only for surging "all need update" test, to verify the toe-dipping behavior)
// Request validate (8)            -->
//                                 <-- validated
// Detach instance                 -->
// Request validate (7)            -->
//                                 <-- validated
// Detach instance                 -->
// (end only for surging "all need update" tests)
// (Only for surging "all but one need update" test)
// Request validate (7)            -->
//                                 <-- validated
// Detach instance                 -->
// Detach instance                 -->
// (end only for surging "all but one need update" test)
// (Only for non-surging "all need update" tests, to verify the toe-dipping behavior)
// Request validate (7)            -->
//                                 <-- validated
// Request terminate 1 node (7)    -->
//                                 <-- 1 node terminated, 6 left
// (end only for non-surging "all need update" tests)
// Request validate (6)            -->
//                                 <-- validated
// Request terminate 2 nodes (6,5) -->
//                                 <-- 1 node terminated (5), 5 left
// Request validate (4)            -->
//                                 <-- 1 node terminated (6), 4 left
//                                 <-- validated
// Request terminate 2 nodes (4,3) -->
//                                 <-- 1 node terminated (3), 3 left
// Request validate (2)            -->
//                                 <-- validated
// Request terminate 1 node (2)    -->
//                                 <-- 1 node terminated (2), 2 left
// Request validate (1)            -->
//                                 <-- 1 node terminated (4), 1 left
//                                 <-- validated
// Request terminate 1 node (1)    -->
//                                 <-- 1 node terminated, 0 left
// Request validate (0)            -->
//                                 <-- validated

type concurrentTest struct {
	ec2iface.EC2API
	t                       *testing.T
	mutex                   sync.Mutex
	surge                   int
	terminationRequestsLeft int
	previousValidation      int
	validationChan          chan bool
	terminationChan         chan bool
	detached                map[string]bool
}

func (c *concurrentTest) Validate() (*validation.ValidationCluster, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.detached) < c.surge {
		assert.Greater(c.t, c.previousValidation, 7, "previous validation")
		c.previousValidation--
		return &validation.ValidationCluster{}, nil
	}

	terminationRequestsLeft := c.terminationRequestsLeft
	switch terminationRequestsLeft {
	case 7, 6, 0:
		assert.Equal(c.t, terminationRequestsLeft+1, c.previousValidation, "previous validation")
	case 5, 3:
		c.t.Errorf("unexpected call to Validate with %d termination requests left", terminationRequestsLeft)
	case 4:
		assert.Equal(c.t, 6, c.previousValidation, "previous validation")
		select {
		case c.terminationChan <- true:
		default:
			c.t.Error("terminationChan is full")
		}
		c.mutex.Unlock()
		select {
		case <-c.validationChan:
		case <-time.After(1 * time.Second):
			c.t.Error("timed out reading from validationChan")
		}
		c.mutex.Lock()
	case 2:
		assert.Equal(c.t, 4, c.previousValidation, "previous validation")
	case 1:
		assert.Equal(c.t, 2, c.previousValidation, "previous validation")
		select {
		case c.terminationChan <- true:
		default:
			c.t.Error("terminationChan is full")
		}
		c.mutex.Unlock()
		select {
		case <-c.validationChan:
		case <-time.After(1 * time.Second):
			c.t.Error("timed out reading from validationChan")
		}
		c.mutex.Lock()
	}
	c.previousValidation = terminationRequestsLeft

	return &validation.ValidationCluster{}, nil
}

func (c *concurrentTest) TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	if input.DryRun != nil && *input.DryRun {
		return &ec2.TerminateInstancesOutput{}, nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, id := range input.InstanceIds {
		assert.Equal(c.t, c.surge, len(c.detached), "Number of detached instances")
		if c.detached[*id] {
			assert.LessOrEqual(c.t, c.terminationRequestsLeft, c.surge, "Deleting detached instances last")
		}

		terminationRequestsLeft := c.terminationRequestsLeft
		c.terminationRequestsLeft--
		switch terminationRequestsLeft {
		case 7, 2, 1:
			assert.Equal(c.t, terminationRequestsLeft, c.previousValidation, "previous validation")
		case 6, 4:
			assert.Equal(c.t, terminationRequestsLeft, c.previousValidation, "previous validation")
			c.mutex.Unlock()
			select {
			case <-c.terminationChan:
			case <-time.After(1 * time.Second):
				c.t.Error("timed out reading from terminationChan")
			}
			c.mutex.Lock()
			go c.delayThenWakeValidation()
		case 5, 3:
			assert.Equal(c.t, terminationRequestsLeft+1, c.previousValidation, "previous validation")
		}
	}
	return c.EC2API.TerminateInstances(input)
}

func (c *concurrentTest) delayThenWakeValidation() {
	time.Sleep(20 * time.Millisecond) // NodeInterval plus some
	select {
	case c.validationChan <- true:
	default:
		c.t.Error("validationChan is full")
	}
}

func (c *concurrentTest) AssertComplete() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	assert.Equal(c.t, 0, c.previousValidation, "last validation")
}

func newConcurrentTest(t *testing.T, cloud *awsup.MockAWSCloud, numSurge int, allNeedUpdate bool) *concurrentTest {
	test := concurrentTest{
		EC2API:                  cloud.MockEC2,
		t:                       t,
		surge:                   numSurge,
		terminationRequestsLeft: 6,
		validationChan:          make(chan bool),
		terminationChan:         make(chan bool),
		detached:                map[string]bool{},
	}
	if numSurge == 0 && allNeedUpdate {
		test.terminationRequestsLeft = 7
	}
	if numSurge == 0 {
		test.previousValidation = test.terminationRequestsLeft + 1
	} else if allNeedUpdate {
		test.previousValidation = 9
	} else {
		test.previousValidation = 8
	}
	return &test
}

func TestRollingUpdateMaxUnavailableAllNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	concurrentTest := newConcurrentTest(t, cloud, 0, true)
	c.ValidateSuccessDuration = 0
	c.ClusterValidator = concurrentTest
	cloud.MockEC2 = concurrentTest

	two := intstr.FromInt(2)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxUnavailable: &two,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 7, 7)

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 0)
	concurrentTest.AssertComplete()
}

func TestRollingUpdateMaxUnavailableAllButOneNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	concurrentTest := newConcurrentTest(t, cloud, 0, false)
	c.ValidateSuccessDuration = 0
	c.ClusterValidator = concurrentTest
	cloud.MockEC2 = concurrentTest

	two := intstr.FromInt(2)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxUnavailable: &two,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 7, 6)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 1)
	concurrentTest.AssertComplete()
}

func TestRollingUpdateMaxUnavailableAllNeedUpdateMaster(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	concurrentTest := newConcurrentTest(t, cloud, 0, true)
	c.ValidateSuccessDuration = 0
	c.ClusterValidator = concurrentTest
	cloud.MockEC2 = concurrentTest

	two := intstr.FromInt(2)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxUnavailable: &two,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "master-1", kopsapi.InstanceGroupRoleMaster, 7, 7)

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "master-1", 0)
	concurrentTest.AssertComplete()
}

type concurrentTestAutoscaling struct {
	autoscalingiface.AutoScalingAPI
	ConcurrentTest *concurrentTest
}

func (m *concurrentTestAutoscaling) DetachInstances(input *autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	m.ConcurrentTest.mutex.Lock()
	defer m.ConcurrentTest.mutex.Unlock()

	assert.Equal(m.ConcurrentTest.t, "node-1", *input.AutoScalingGroupName)
	assert.False(m.ConcurrentTest.t, *input.ShouldDecrementDesiredCapacity)

	for _, id := range input.InstanceIds {
		assert.Less(m.ConcurrentTest.t, len(m.ConcurrentTest.detached), m.ConcurrentTest.surge, "Number of detached instances")
		assert.False(m.ConcurrentTest.t, m.ConcurrentTest.detached[*id], *id+" already detached")
		m.ConcurrentTest.detached[*id] = true
	}
	return &autoscaling.DetachInstancesOutput{}, nil
}

type ec2IgnoreTags struct {
	ec2iface.EC2API
}

// CreateTags ignores tagging of instances done by the AWS fi.Cloud implementation of DetachInstance()
func (e *ec2IgnoreTags) CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	return &ec2.CreateTagsOutput{}, nil
}

func TestRollingUpdateMaxSurgeAllNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	concurrentTest := newConcurrentTest(t, cloud, 2, true)
	c.ValidateSuccessDuration = 0
	c.ClusterValidator = concurrentTest
	cloud.MockAutoscaling = &concurrentTestAutoscaling{
		AutoScalingAPI: cloud.MockAutoscaling,
		ConcurrentTest: concurrentTest,
	}
	cloud.MockEC2 = &ec2IgnoreTags{EC2API: concurrentTest}

	two := intstr.FromInt(2)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxSurge: &two,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 6, 6)

	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 0)
	concurrentTest.AssertComplete()
}

func TestRollingUpdateMaxSurgeAllButOneNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	concurrentTest := newConcurrentTest(t, cloud, 2, false)
	c.ValidateSuccessDuration = 0
	c.ClusterValidator = concurrentTest
	cloud.MockAutoscaling = &concurrentTestAutoscaling{
		AutoScalingAPI: cloud.MockAutoscaling,
		ConcurrentTest: concurrentTest,
	}
	cloud.MockEC2 = &ec2IgnoreTags{EC2API: concurrentTest}

	two := intstr.FromInt(2)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxSurge: &two,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 7, 6)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 1)
	concurrentTest.AssertComplete()
}

type countDetach struct {
	autoscalingiface.AutoScalingAPI
	Count int
}

func (c *countDetach) DetachInstances(input *autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	c.Count += len(input.InstanceIds)
	return &autoscaling.DetachInstancesOutput{}, nil
}

func TestRollingUpdateMaxSurgeGreaterThanNeedUpdate(t *testing.T) {
	c, cloud, cluster := getTestSetup()

	countDetach := &countDetach{AutoScalingAPI: cloud.MockAutoscaling}
	cloud.MockAutoscaling = countDetach
	cloud.MockEC2 = &ec2IgnoreTags{EC2API: cloud.MockEC2}

	ten := intstr.FromInt(10)
	cluster.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxSurge: &ten,
	}

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, "node-1", kopsapi.InstanceGroupRoleNode, 3, 2)
	err := c.RollingUpdate(groups, cluster, &kopsapi.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")

	assertGroupInstanceCount(t, cloud, "node-1", 1)
	assert.Equal(t, 2, countDetach.Count)
}

// TODO tests for surging when instances start off already detached

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
