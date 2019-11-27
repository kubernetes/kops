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

package awstasks

import (
	"fmt"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=SecurityGroupRule
type SecurityGroupRule struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	SecurityGroup *SecurityGroup
	CIDR          *string
	Protocol      *string

	// FromPort is the lower-bound (inclusive) of the port-range
	FromPort *int64
	// ToPort is the upper-bound (inclusive) of the port-range
	ToPort      *int64
	SourceGroup *SecurityGroup

	Egress *bool
}

func (e *SecurityGroupRule) Find(c *fi.Context) (*SecurityGroupRule, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.SecurityGroup == nil || e.SecurityGroup.ID == nil {
		return nil, nil
	}

	if e.SourceGroup != nil && e.SourceGroup.ID == nil {
		klog.V(4).Infof("Skipping find of SecurityGroupRule %s, because SourceGroup was not found", fi.StringValue(e.Name))
		return nil, nil
	}

	request := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			awsup.NewEC2Filter("group-id", *e.SecurityGroup.ID),
		},
	}

	response, err := cloud.EC2().DescribeSecurityGroups(request)
	if err != nil {
		return nil, fmt.Errorf("error listing SecurityGroup: %v", err)
	}

	if response == nil || len(response.SecurityGroups) == 0 {
		return nil, nil
	}

	if len(response.SecurityGroups) != 1 {
		klog.Fatalf("found multiple security groups for id=%s", *e.SecurityGroup.ID)
	}
	sg := response.SecurityGroups[0]
	//klog.V(2).Info("found existing security group")

	var foundRule *ec2.IpPermission

	ipPermissions := sg.IpPermissions
	if fi.BoolValue(e.Egress) {
		ipPermissions = sg.IpPermissionsEgress
	}

	for _, rule := range ipPermissions {
		if e.matches(rule) {
			foundRule = rule
			break
		}
	}

	if foundRule != nil {
		actual := &SecurityGroupRule{
			Name:          e.Name,
			SecurityGroup: &SecurityGroup{ID: e.SecurityGroup.ID},
			FromPort:      foundRule.FromPort,
			ToPort:        foundRule.ToPort,
			Protocol:      foundRule.IpProtocol,
			Egress:        e.Egress,
		}

		if aws.StringValue(actual.Protocol) == "-1" {
			actual.Protocol = nil
		}
		if e.CIDR != nil {
			actual.CIDR = e.CIDR
		}
		if e.SourceGroup != nil {
			actual.SourceGroup = &SecurityGroup{ID: e.SourceGroup.ID}
		}

		// Avoid spurious changes
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *SecurityGroupRule) matches(rule *ec2.IpPermission) bool {
	if aws.Int64Value(rule.FromPort) != aws.Int64Value(e.FromPort) {
		return false
	}
	if aws.Int64Value(rule.ToPort) != aws.Int64Value(e.ToPort) {
		return false
	}

	matchProtocol := "-1" // Wildcard
	if e.Protocol != nil {
		matchProtocol = *e.Protocol
	}
	if aws.StringValue(rule.IpProtocol) != matchProtocol {
		return false
	}

	if e.CIDR != nil {
		// TODO: Only if len 1?
		match := false
		for _, ipRange := range rule.IpRanges {
			if aws.StringValue(ipRange.CidrIp) == *e.CIDR {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	if e.SourceGroup != nil {
		// TODO: Only if len 1?
		match := false
		for _, spec := range rule.UserIdGroupPairs {
			if e.SourceGroup == nil {
				continue
			}

			if e.SourceGroup.ID == nil {
				klog.Warningf("SourceGroup had nil ID: %v", e.SourceGroup)
				continue
			}

			if aws.StringValue(spec.GroupId) == *e.SourceGroup.ID {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func (e *SecurityGroupRule) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *SecurityGroupRule) CheckChanges(a, e, changes *SecurityGroupRule) error {
	if a == nil {
		if e.SecurityGroup == nil {
			return field.Required(field.NewPath("SecurityGroup"), "")
		}
	}

	if e.FromPort != nil && e.Protocol == nil {
		return field.Required(field.NewPath("Protocol"), "Protocol must be specified with FromPort")
	}
	if e.ToPort != nil && e.Protocol == nil {
		return field.Required(field.NewPath("Protocol"), "Protocol must be specified with ToPort")
	}

	return nil
}

// Description returns a human readable summary of the security group rule
func (e *SecurityGroupRule) Description() string {
	var description []string

	if e.Protocol != nil {
		description = append(description, fmt.Sprintf("protocol=%s", *e.Protocol))
	}

	if e.FromPort != nil {
		description = append(description, fmt.Sprintf("fromPort=%d", *e.FromPort))
	}

	if e.ToPort != nil {
		description = append(description, fmt.Sprintf("toPort=%d", *e.ToPort))
	}

	if e.SourceGroup != nil {
		description = append(description, fmt.Sprintf("sourceGroup=%s", fi.StringValue(e.SourceGroup.ID)))
	}

	if e.CIDR != nil {
		description = append(description, fmt.Sprintf("cidr=%s", *e.CIDR))
	}

	return strings.Join(description, " ")
}

func (_ *SecurityGroupRule) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SecurityGroupRule) error {
	name := fi.StringValue(e.Name)

	if a == nil {
		protocol := e.Protocol
		if protocol == nil {
			protocol = aws.String("-1")
		}

		ipPermission := &ec2.IpPermission{
			IpProtocol: protocol,
			FromPort:   e.FromPort,
			ToPort:     e.ToPort,
		}

		if e.SourceGroup != nil {
			ipPermission.UserIdGroupPairs = []*ec2.UserIdGroupPair{
				{
					GroupId: e.SourceGroup.ID,
				},
			}
		} else {
			// Default to 0.0.0.0/0 ?
			ipPermission.IpRanges = []*ec2.IpRange{
				{CidrIp: e.CIDR},
			}
		}

		description := e.Description()

		if fi.BoolValue(e.Egress) {
			request := &ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: e.SecurityGroup.ID,
			}
			request.IpPermissions = []*ec2.IpPermission{ipPermission}

			klog.V(2).Infof("%s: Calling EC2 AuthorizeSecurityGroupEgress (%s)", name, description)
			_, err := t.Cloud.EC2().AuthorizeSecurityGroupEgress(request)
			if err != nil {
				return fmt.Errorf("error creating SecurityGroupEgress: %v", err)
			}
		} else {
			request := &ec2.AuthorizeSecurityGroupIngressInput{
				GroupId: e.SecurityGroup.ID,
			}
			request.IpPermissions = []*ec2.IpPermission{ipPermission}

			klog.V(2).Infof("%s: Calling EC2 AuthorizeSecurityGroupIngress (%s)", name, description)
			_, err := t.Cloud.EC2().AuthorizeSecurityGroupIngress(request)
			if err != nil {
				return fmt.Errorf("error creating SecurityGroupIngress: %v", err)
			}
		}

	}

	// No tags on security group rules (there are tags on the group though)

	return nil
}

type terraformSecurityGroupIngress struct {
	Type *string `json:"type"`

	SecurityGroup *terraform.Literal `json:"security_group_id"`
	SourceGroup   *terraform.Literal `json:"source_security_group_id,omitempty"`

	FromPort *int64 `json:"from_port,omitempty"`
	ToPort   *int64 `json:"to_port,omitempty"`

	Protocol   *string  `json:"protocol,omitempty"`
	CIDRBlocks []string `json:"cidr_blocks,omitempty"`
}

func (_ *SecurityGroupRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SecurityGroupRule) error {
	tf := &terraformSecurityGroupIngress{
		Type:          fi.String("ingress"),
		SecurityGroup: e.SecurityGroup.TerraformLink(),
		FromPort:      e.FromPort,
		ToPort:        e.ToPort,
		Protocol:      e.Protocol,
	}
	if fi.BoolValue(e.Egress) {
		tf.Type = fi.String("egress")
	}

	if e.Protocol == nil {
		tf.Protocol = fi.String("-1")
		tf.FromPort = fi.Int64(0)
		tf.ToPort = fi.Int64(0)
	}

	if tf.FromPort == nil {
		// FromPort is required by tf
		tf.FromPort = fi.Int64(0)
	}
	if tf.ToPort == nil {
		// ToPort is required by tf
		tf.ToPort = fi.Int64(65535)
	}

	if e.SourceGroup != nil {
		tf.SourceGroup = e.SourceGroup.TerraformLink()
	}

	if e.CIDR != nil {
		tf.CIDRBlocks = append(tf.CIDRBlocks, *e.CIDR)
	}
	return t.RenderResource("aws_security_group_rule", *e.Name, tf)
}

type cloudformationSecurityGroupIngress struct {
	SecurityGroup *cloudformation.Literal `json:"GroupId,omitempty"`
	SourceGroup   *cloudformation.Literal `json:"SourceSecurityGroupId,omitempty"`

	FromPort *int64 `json:"FromPort,omitempty"`
	ToPort   *int64 `json:"ToPort,omitempty"`

	Protocol *string `json:"IpProtocol,omitempty"`
	CidrIp   *string `json:"CidrIp,omitempty"`
}

func (_ *SecurityGroupRule) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *SecurityGroupRule) error {
	cfType := "AWS::EC2::SecurityGroupIngress"
	if fi.BoolValue(e.Egress) {
		cfType = "AWS::EC2::SecurityGroupEgress"
	}

	tf := &cloudformationSecurityGroupIngress{
		SecurityGroup: e.SecurityGroup.CloudformationLink(),
		FromPort:      e.FromPort,
		ToPort:        e.ToPort,
		Protocol:      e.Protocol,
	}

	if e.Protocol == nil {
		tf.Protocol = fi.String("-1")
		tf.FromPort = fi.Int64(0)
		tf.ToPort = fi.Int64(0)
	}

	if tf.FromPort == nil {
		// FromPort is required by tf
		tf.FromPort = fi.Int64(0)
	}
	if tf.ToPort == nil {
		// ToPort is required by tf
		tf.ToPort = fi.Int64(65535)
	}

	if e.SourceGroup != nil {
		tf.SourceGroup = e.SourceGroup.CloudformationLink()
	}

	if e.CIDR != nil {
		tf.CidrIp = e.CIDR
	}

	return t.RenderResource(cfType, *e.Name, tf)
}
