/*
Copyright 2022 The Kubernetes Authors.

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

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/apimachinery/pkg/util/intstr"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
)

func TestWarmPoolOnlyRoll(t *testing.T) {
	c, cloud := getTestSetup()

	groupName := "warmPoolOnly"
	instanceID := "node-1"

	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	makeGroup(groups, c.K8sClient, cloud, groupName, kopsapi.InstanceGroupRoleNode, 0, 0)

	group := groups[groupName]
	group.MinSize = 0
	group.MaxSize = 10

	maxSurge := intstr.FromString("25%")

	group.InstanceGroup.Spec.RollingUpdate = &kopsapi.RollingUpdate{
		MaxSurge: &maxSurge,
	}

	cloud.Autoscaling().AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: &groupName,
		InstanceIds:          []*string{&instanceID},
	})

	instance, err := group.NewCloudInstance("node-1", cloudinstances.CloudInstanceStatusNeedsUpdate, nil)
	if err != nil {
		t.Fatalf("could not create cloud instance: %v", err)
	}

	instance.State = cloudinstances.WarmPool

	{
		err := c.rollingUpdateInstanceGroup(group, 0*time.Second)
		if err != nil {
			t.Fatalf("could not roll instance group: %v", err)
		}
	}
}
