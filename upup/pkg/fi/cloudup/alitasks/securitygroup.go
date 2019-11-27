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
	"fmt"

	"k8s.io/klog"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=SecurityGroup
const SecurityResource = "securitygroup"

type SecurityGroup struct {
	Name            *string
	SecurityGroupId *string
	Lifecycle       *fi.Lifecycle
	VPC             *VPC
	Tags            map[string]string
}

var _ fi.CompareWithID = &SecurityGroup{}

func (s *SecurityGroup) CompareWithID() *string {
	return s.SecurityGroupId
}

func (s *SecurityGroup) Find(c *fi.Context) (*SecurityGroup, error) {
	if s.VPC == nil || s.VPC.ID == nil {
		klog.V(4).Infof("VPC / VPCId not found for %s, skipping Find", fi.StringValue(s.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)
	describeSecurityGroupsArgs := &ecs.DescribeSecurityGroupsArgs{
		RegionId: common.Region(cloud.Region()),
		VpcId:    fi.StringValue(s.VPC.ID),
	}

	securityGroupList, _, err := cloud.EcsClient().DescribeSecurityGroups(describeSecurityGroupsArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding SecurityGroups : %v", err)
	}

	// Don't exist securityGroup with specified  Name.
	if len(securityGroupList) == 0 {
		return nil, nil
	}

	actual := &SecurityGroup{}
	securityGroups := []ecs.SecurityGroupItemType{}

	// Find the securityGroup match the name and tags
	for _, securityGroup := range securityGroupList {
		if securityGroup.SecurityGroupName == fi.StringValue(s.Name) {
			securityGroups = append(securityGroups, securityGroup)
		}
	}

	for _, securityGroup := range securityGroups {
		resourceType := SecurityResource
		find := true
		tags, err := cloud.GetTags(securityGroup.SecurityGroupId, resourceType)
		if err != nil {
			return nil, fmt.Errorf("err finding SecurityGroups,%v", err)
		}

		if s.Tags != nil {
			for key, value := range s.Tags {
				if v, ok := tags[key]; !ok || v != value {
					find = false
				}
			}
		}

		if find {
			klog.V(2).Infof("found matching SecurityGroup with name: %q", *s.Name)
			actual.Name = fi.String(securityGroup.SecurityGroupName)
			actual.SecurityGroupId = fi.String(securityGroup.SecurityGroupId)
			// Ignore "system" fields
			actual.Lifecycle = s.Lifecycle
			actual.VPC = s.VPC
			actual.Tags = tags
			s.SecurityGroupId = actual.SecurityGroupId
			return actual, nil
		}
	}

	return nil, nil
}

func (s *SecurityGroup) Run(c *fi.Context) error {
	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}
	c.Cloud.(aliup.ALICloud).AddClusterTags(s.Tags)
	return fi.DefaultDeltaRunMethod(s, c)
}

func (_ *SecurityGroup) CheckChanges(a, e, changes *SecurityGroup) error {

	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}

	return nil
}

func (_ *SecurityGroup) RenderALI(t *aliup.ALIAPITarget, a, e, changes *SecurityGroup) error {

	if a == nil {
		klog.V(2).Infof("Creating SecurityGroup with Name:%q", fi.StringValue(e.Name))

		createSecurityGroupArgs := &ecs.CreateSecurityGroupArgs{
			RegionId:          common.Region(t.Cloud.Region()),
			SecurityGroupName: fi.StringValue(e.Name),
			VpcId:             fi.StringValue(e.VPC.ID),
		}

		securityGroupId, err := t.Cloud.EcsClient().CreateSecurityGroup(createSecurityGroupArgs)
		if err != nil {
			return fmt.Errorf("error creating securityGroup: %v", err)
		}
		e.SecurityGroupId = fi.String(securityGroupId)
	}

	resourceType := SecurityResource
	if changes != nil && changes.Tags != nil {
		if err := t.Cloud.CreateTags(*e.SecurityGroupId, resourceType, e.Tags); err != nil {
			return fmt.Errorf("error adding Tags to securityGroup: %v", err)
		}
	}

	if a != nil && (len(a.Tags) > 0) {
		klog.V(2).Infof("Modifying SecurityGroup with Name:%q", fi.StringValue(e.Name))

		tagsToDelete := e.getGroupTagsToDelete(a.Tags)
		if len(tagsToDelete) > 0 {
			if err := t.Cloud.RemoveTags(*e.SecurityGroupId, resourceType, tagsToDelete); err != nil {
				return fmt.Errorf("error removing Tags from ALI YunPan: %v", err)
			}
		}
	}

	return nil
}

func (s *SecurityGroup) getGroupTagsToDelete(currentTags map[string]string) map[string]string {
	tagsToDelete := map[string]string{}
	for k, v := range currentTags {
		if _, ok := s.Tags[k]; !ok {
			tagsToDelete[k] = v
		}
	}

	return tagsToDelete
}

type terraformSecurityGroup struct {
	Name  *string            `json:"name,omitempty"`
	VPCId *terraform.Literal `json:"vpc_id,omitempty"`
}

func (_ *SecurityGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SecurityGroup) error {
	tf := &terraformSecurityGroup{
		Name:  e.Name,
		VPCId: e.VPC.TerraformLink(),
	}

	return t.RenderResource("alicloud_security_group", *e.Name, tf)
}

func (l *SecurityGroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_security_group", *l.Name, "id")
}
