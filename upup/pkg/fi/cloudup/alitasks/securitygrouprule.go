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

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/klog"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=SecurityGroupRule

type SecurityGroupRule struct {
	Name          *string
	IpProtocol    *string
	SourceCidrIp  *string
	SecurityGroup *SecurityGroup
	SourceGroup   *SecurityGroup
	Lifecycle     *fi.Lifecycle
	PortRange     *string
	In            *bool
}

var _ fi.CompareWithID = &SecurityGroupRule{}

func (s *SecurityGroupRule) CompareWithID() *string {
	return s.Name
}

func (s *SecurityGroupRule) Find(c *fi.Context) (*SecurityGroupRule, error) {
	if s.SecurityGroup == nil || s.SecurityGroup.SecurityGroupId == nil {
		klog.V(4).Infof("SecurityGroup / SecurityGroupId not found for %s, skipping Find", fi.StringValue(s.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)
	var direction ecs.Direction

	if fi.BoolValue(s.In) {
		direction = ecs.DirectionIngress
	} else {
		direction = ecs.DirectionEgress
	}

	describeSecurityGroupAttributeArgs := &ecs.DescribeSecurityGroupAttributeArgs{
		RegionId:        common.Region(cloud.Region()),
		SecurityGroupId: fi.StringValue(s.SecurityGroup.SecurityGroupId),
		Direction:       direction,
	}

	describeResponse, err := cloud.EcsClient().DescribeSecurityGroupAttribute(describeSecurityGroupAttributeArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding SecurityGroupRules: %v", err)
	}

	if len(describeResponse.Permissions.Permission) == 0 {
		return nil, nil
	}

	actual := &SecurityGroupRule{}
	// Find securityGroupRule with specified ipProtocol, securityGroupId,SourceGroupId
	for _, securityGroupRule := range describeResponse.Permissions.Permission {

		if securityGroupRule.IpProtocol != ecs.IpProtocol(fi.StringValue(s.IpProtocol)) {
			continue
		}
		if s.SourceGroup != nil && securityGroupRule.SourceGroupId != fi.StringValue(s.SourceGroup.SecurityGroupId) {
			continue
		}
		if s.PortRange != nil && securityGroupRule.PortRange != fi.StringValue(s.PortRange) {
			continue
		}
		if s.SourceCidrIp != nil && securityGroupRule.SourceCidrIp != fi.StringValue(s.SourceCidrIp) {
			continue
		}

		klog.V(2).Infof("found matching SecurityGroupRule of securityGroup: %q", *s.SecurityGroup.SecurityGroupId)

		actual.PortRange = fi.String(securityGroupRule.PortRange)
		actual.SourceCidrIp = fi.String(securityGroupRule.SourceCidrIp)
		actual.IpProtocol = fi.String(string(securityGroupRule.IpProtocol))
		// Ignore "system" fields
		actual.Name = s.Name
		actual.SecurityGroup = s.SecurityGroup
		actual.Lifecycle = s.Lifecycle
		actual.In = s.In
		actual.SourceGroup = s.SourceGroup

		return actual, nil

	}

	return nil, nil
}

func (s *SecurityGroupRule) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, c)
}

func (_ *SecurityGroupRule) CheckChanges(a, e, changes *SecurityGroupRule) error {

	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.IpProtocol == nil {
			return fi.RequiredField("IpProtocol")
		}
		if e.PortRange == nil {
			return fi.RequiredField("PortRange")
		}
	} else {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *SecurityGroupRule) RenderALI(t *aliup.ALIAPITarget, a, e, changes *SecurityGroupRule) error {

	if a == nil {
		if fi.BoolValue(e.In) {
			klog.V(2).Infof("Creating SecurityGroupRule of SecurityGroup:%q", fi.StringValue(e.SecurityGroup.SecurityGroupId))

			authorizeSecurityGroupArgs := &ecs.AuthorizeSecurityGroupArgs{
				SecurityGroupId: fi.StringValue(e.SecurityGroup.SecurityGroupId),
				RegionId:        common.Region(t.Cloud.Region()),
				IpProtocol:      ecs.IpProtocol(fi.StringValue(e.IpProtocol)),
				PortRange:       fi.StringValue(e.PortRange),
			}

			if e.SourceGroup != nil && e.SourceGroup.SecurityGroupId != nil {
				authorizeSecurityGroupArgs.SourceGroupId = fi.StringValue(e.SourceGroup.SecurityGroupId)
			}

			if e.SourceCidrIp != nil {
				authorizeSecurityGroupArgs.SourceCidrIp = fi.StringValue(e.SourceCidrIp)
			}

			err := t.Cloud.EcsClient().AuthorizeSecurityGroup(authorizeSecurityGroupArgs)
			if err != nil {
				return fmt.Errorf("error creating securityGroupRule: %v", err)
			}

		} else {
			authorizeSecurityGroupEgressArgs := &ecs.AuthorizeSecurityGroupEgressArgs{
				SecurityGroupId: fi.StringValue(e.SecurityGroup.SecurityGroupId),
				RegionId:        common.Region(t.Cloud.Region()),
				IpProtocol:      ecs.IpProtocol(fi.StringValue(e.IpProtocol)),
				PortRange:       fi.StringValue(e.PortRange),
			}

			err := t.Cloud.EcsClient().AuthorizeSecurityGroupEgress(authorizeSecurityGroupEgressArgs)
			if err != nil {
				return fmt.Errorf("error creating securityGroupRule: %v", err)
			}
		}

	}
	return nil
}

type terraformSecurityGroupRole struct {
	Name            *string            `json:"name,omitempty"`
	Type            *string            `json:"type,omitempty"`
	IpProtocol      *string            `json:"ip_protocol,omitempty"`
	SourceCidrIp    *string            `json:"cidr_ip,omitempty"`
	SecurityGroupId *terraform.Literal `json:"security_group_id ,omitempty"`
	SourceGroupId   *terraform.Literal `json:"source_security_group_id  ,omitempty"`
	PortRange       *string            `json:"port_range,omitempty"`
}

func (_ *SecurityGroupRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SecurityGroupRule) error {

	tf := &terraformSecurityGroupRole{
		Name:            e.Name,
		IpProtocol:      e.IpProtocol,
		PortRange:       e.PortRange,
		SecurityGroupId: e.SecurityGroup.TerraformLink(),
	}

	if fi.BoolValue(e.In) {
		ruleType := "ingress"
		tf.Type = &ruleType

		if e.SourceGroup != nil {
			tf.SourceGroupId = e.SecurityGroup.TerraformLink()
		}

		if e.SourceCidrIp != nil {
			tf.SourceCidrIp = e.SourceCidrIp
		}
	} else {
		ruleType := "egress"
		tf.Type = &ruleType
	}

	return t.RenderResource("alicloud_security_group_rule", *e.Name, tf)
}

func (l *SecurityGroupRule) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_security_group_rule", *l.Name, "id")
}
