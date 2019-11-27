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

package alitasks

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ess"
	"k8s.io/klog"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=ScalingGroup

type ScalingGroup struct {
	Name           *string
	Lifecycle      *fi.Lifecycle
	ScalingGroupId *string
	LoadBalancer   *LoadBalancer
	VSwitchs       []*VSwitch
	MinSize        *int
	MaxSize        *int
	Active         *bool
}

var _ fi.CompareWithID = &ScalingGroup{}

func (s *ScalingGroup) CompareWithID() *string {
	return s.ScalingGroupId
}

func (s *ScalingGroup) Find(c *fi.Context) (*ScalingGroup, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	describeScalingGroupsArgs := &ess.DescribeScalingGroupsArgs{
		RegionId:         common.Region(cloud.Region()),
		ScalingGroupName: common.FlattenArray{fi.StringValue(s.Name)},
	}

	groupList, _, err := cloud.EssClient().DescribeScalingGroups(describeScalingGroupsArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding scalingGroup: %v", err)
	}

	// Don't exist scalingGroup with specified ClusterTags or Name.
	if len(groupList) == 0 {
		return nil, nil
	}

	if len(groupList) > 1 {
		klog.V(4).Infof("The number of specified scalingGroup with the same name and ClusterTags exceeds 1, diskName:%q", *s.Name)
	}

	klog.V(2).Infof("found matching ScalingGroup with Name: %q", *s.Name)

	actual := &ScalingGroup{}
	actual.Name = fi.String(groupList[0].ScalingGroupName)
	actual.MinSize = fi.Int(groupList[0].MinSize)
	actual.MaxSize = fi.Int(groupList[0].MaxSize)
	actual.ScalingGroupId = fi.String(groupList[0].ScalingGroupId)
	actual.Active = fi.Bool(groupList[0].LifecycleState == ess.Active)

	actual.LoadBalancer = &LoadBalancer{
		LoadbalancerId: fi.String(groupList[0].LoadBalancerId),
	}

	if len(groupList[0].VSwitchIds.VSwitchId) != 0 {
		for _, vswitch := range groupList[0].VSwitchIds.VSwitchId {
			v := &VSwitch{
				VSwitchId: fi.String(vswitch),
			}
			actual.VSwitchs = append(actual.VSwitchs, v)
		}
	}

	// Ignore "system" fields
	s.ScalingGroupId = actual.ScalingGroupId
	s.Active = actual.Active
	actual.Lifecycle = s.Lifecycle
	return actual, nil

}

func (a *ScalingGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(a, c)
}

func (_ *ScalingGroup) CheckChanges(a, e, changes *ScalingGroup) error {
	if a == nil {
		if e.MaxSize == nil {
			return fi.RequiredField("MaxSize")
		}
		if e.MinSize == nil {
			return fi.RequiredField("MinSize")
		}
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *ScalingGroup) RenderALI(t *aliup.ALIAPITarget, a, e, changes *ScalingGroup) error {
	vswitchs := common.FlattenArray{}
	for _, vswitch := range e.VSwitchs {
		if vswitch.VSwitchId == nil {
			return errors.New("error updating scalingGroup, lack of VSwitchId")
		}
		vswitchs = append(vswitchs, fi.StringValue(vswitch.VSwitchId))
	}

	if a == nil {
		klog.V(2).Infof("Creating ScalingGroup with Name:%q", fi.StringValue(e.Name))

		createScalingGroupArgs := &ess.CreateScalingGroupArgs{
			ScalingGroupName: fi.StringValue(e.Name),
			RegionId:         common.Region(t.Cloud.Region()),
			MinSize:          e.MinSize,
			MaxSize:          e.MaxSize,
			VSwitchIds:       vswitchs,
		}

		if e.LoadBalancer != nil && e.LoadBalancer.LoadbalancerId != nil {
			loadBalancerIds := []string{fi.StringValue(e.LoadBalancer.LoadbalancerId)}
			loadBalancerId, _ := json.Marshal(loadBalancerIds)
			createScalingGroupArgs.LoadBalancerIds = string(loadBalancerId)
		}

		createScalingGroupResponse, err := t.Cloud.EssClient().CreateScalingGroup(createScalingGroupArgs)
		if err != nil {
			return fmt.Errorf("error creating autoscalingGroup: %v", err)
		}

		e.ScalingGroupId = fi.String(createScalingGroupResponse.ScalingGroupId)
		e.Active = fi.Bool(false)

	} else {
		//only support to update size
		if changes.MinSize != nil || changes.MaxSize != nil {
			klog.V(2).Infof("Modifying AutoscalingGroup with Name:%q", fi.StringValue(e.Name))

			modifyScalingGroupArgs := &ess.ModifyScalingGroupArgs{
				ScalingGroupId: fi.StringValue(a.ScalingGroupId),
				MinSize:        e.MinSize,
				MaxSize:        e.MaxSize,
			}
			_, err := t.Cloud.EssClient().ModifyScalingGroup(modifyScalingGroupArgs)
			if err != nil {
				return fmt.Errorf("error modifying autoscalingGroup: %v", err)
			}
		}
	}

	return nil
}

type terraformScalingGroup struct {
	Name    *string `json:"scaling_group_name,omitempty"`
	MaxSize *int    `json:"max_size,omitempty"`
	MinSize *int    `json:"min_size,omitempty"`

	VSwitchs     []*terraform.Literal `json:"vswitch_ids,omitempty"`
	LoadBalancer []*terraform.Literal `json:"loadbalancer_ids,omitempty"`
}

func (_ *ScalingGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ScalingGroup) error {
	tf := &terraformScalingGroup{
		Name:    e.Name,
		MinSize: e.MinSize,
		MaxSize: e.MaxSize,
	}

	if len(e.VSwitchs) != 0 {
		for _, s := range e.VSwitchs {
			tf.VSwitchs = append(tf.VSwitchs, s.TerraformLink())
		}
	}

	if e.LoadBalancer != nil {
		tf.LoadBalancer = append(tf.LoadBalancer, e.LoadBalancer.TerraformLink())
	}

	return t.RenderResource("alicloud_ess_scaling_group", *e.Name, tf)
}

func (a *ScalingGroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_ess_scaling_group", *a.Name, "id")
}
