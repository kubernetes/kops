/*
Copyright 2021 The Kubernetes Authors.

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

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// Here we have three nodes that are up to date, while three warm nodes need updating.
// Only the initial cluster validation should be run
func TestRollingUpdateOnlyWarmPoolNodes(t *testing.T) {
	c, cloud := getTestSetup()
	k8sClient := c.K8sClient
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroupWithWarmPool(groups, k8sClient, cloud, "node-1", kops.InstanceGroupRoleNode, 3, 0, 3, 3)

	validator := &countingValidator{}
	c.ClusterValidator = validator

	assert.Equal(t, 3, len(groups["node-1"].NeedUpdate), "number of nodes needing update")

	err := c.RollingUpdate(groups, &kops.InstanceGroupList{})
	assert.NoError(t, err, "rolling update")
	assert.Equal(t, 1, validator.numValidations, "number of validations")
}

func TestRollingWarmPoolBeforeJoinedNodes(t *testing.T) {
	c, cloud := getTestSetup()
	k8sClient := c.K8sClient
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroupWithWarmPool(groups, k8sClient, cloud, "node-1", kops.InstanceGroupRoleNode, 3, 3, 3, 3)

	warmPoolBeforeJoinedNodesTest := &warmPoolBeforeJoinedNodesTest{
		EC2API: cloud.MockEC2,
		t:      t,
	}
	cloud.MockEC2 = warmPoolBeforeJoinedNodesTest

	err := c.RollingUpdate(groups, &kops.InstanceGroupList{})

	assert.NoError(t, err, "rolling update")

	assert.Equal(t, 6, warmPoolBeforeJoinedNodesTest.numTerminations, "Number of terminations")
}

type countingValidator struct {
	numValidations int
}

func (c *countingValidator) Validate() (*validation.ValidationCluster, error) {
	c.numValidations++
	return &validation.ValidationCluster{}, nil
}

func makeGroupWithWarmPool(groups map[string]*cloudinstances.CloudInstanceGroup, k8sClient kubernetes.Interface, cloud *awsup.MockAWSCloud, name string, role kops.InstanceGroupRole, count int, needUpdate int, warmCount int, warmNeedUpdate int) {
	makeGroup(groups, k8sClient, cloud, name, role, count, needUpdate)

	group := groups[name]

	wpInstances := []*autoscaling.Instance{}
	warmStoppedState := autoscaling.LifecycleStateWarmedStopped
	for i := 0; i < warmCount; i++ {
		id := name + "-wp-" + string(rune('a'+i))
		instance := &autoscaling.Instance{
			InstanceId:     &id,
			LifecycleState: &warmStoppedState,
		}
		wpInstances = append(wpInstances, instance)

		cm, _ := group.NewCloudInstance(id, cloudinstances.CloudInstanceStatusNeedsUpdate, nil)
		cm.State = cloudinstances.WarmPool

	}

	// There is no API to write to warm pools, so we need to cheat.
	mockASG := cloud.MockAutoscaling.(*mockautoscaling.MockAutoscaling)
	mockASG.WarmPoolInstances[name] = wpInstances
}

type warmPoolBeforeJoinedNodesTest struct {
	ec2iface.EC2API
	t               *testing.T
	numTerminations int
}

func (t *warmPoolBeforeJoinedNodesTest) TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	t.numTerminations++

	return t.EC2API.TerminateInstances(input)
}
