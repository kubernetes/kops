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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// +kops:fitask
type SecurityGroupRule struct {
	ID        *string
	Name      *string
	Lifecycle fi.Lifecycle

	SecurityGroup *SecurityGroup
	CIDR          *string
	IPv6CIDR      *string
	PrefixList    *string
	Protocol      *string

	// FromPort is the lower-bound (inclusive) of the port-range
	FromPort *int64
	// ToPort is the upper-bound (inclusive) of the port-range
	ToPort      *int64
	SourceGroup *SecurityGroup

	Egress *bool

	Tags map[string]string
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

	request := &ec2.DescribeSecurityGroupRulesInput{
		Filters: []*ec2.Filter{
			awsup.NewEC2Filter("group-id", *e.SecurityGroup.ID),
		},
	}

	response, err := cloud.EC2().DescribeSecurityGroupRules(request)
	if err != nil {
		return nil, fmt.Errorf("error listing SecurityGroup: %v", err)
	}

	if response == nil || len(response.SecurityGroupRules) == 0 {
		return nil, nil
	}

	var foundRule *ec2.SecurityGroupRule

	for _, rule := range response.SecurityGroupRules {
		if e.matches(rule) {
			foundRule = rule
			break
		}
	}

	if foundRule != nil {
		actual := &SecurityGroupRule{
			ID:            foundRule.SecurityGroupRuleId,
			Name:          e.Name,
			SecurityGroup: &SecurityGroup{ID: e.SecurityGroup.ID},
			FromPort:      foundRule.FromPort,
			ToPort:        foundRule.ToPort,
			Protocol:      foundRule.IpProtocol,
			Egress:        e.Egress,

			Tags: intersectTags(foundRule.Tags, e.Tags),
		}

		if aws.StringValue(actual.Protocol) == "-1" {
			actual.Protocol = nil
		}

		if fi.StringValue(actual.Protocol) != "icmpv6" {
			if fi.Int64Value(actual.FromPort) == int64(-1) {
				actual.FromPort = nil
			}
			if fi.Int64Value(actual.ToPort) == int64(-1) {
				actual.ToPort = nil
			}
		}

		if e.CIDR != nil {
			actual.CIDR = e.CIDR
		}
		if e.IPv6CIDR != nil {
			actual.IPv6CIDR = e.IPv6CIDR
		}
		if e.PrefixList != nil {
			actual.PrefixList = e.PrefixList
		}
		if e.SourceGroup != nil {
			actual.SourceGroup = &SecurityGroup{ID: e.SourceGroup.ID}
		}

		// Avoid spurious changes
		actual.Lifecycle = e.Lifecycle

		e.ID = actual.ID

		return actual, nil
	}
	return nil, nil
}

func (e *SecurityGroupRule) SetCidrOrPrefix(cidr string) {
	if strings.HasPrefix(cidr, "pl-") {
		e.PrefixList = &cidr
	} else if utils.IsIPv6CIDR(cidr) {
		e.IPv6CIDR = &cidr
	} else {
		e.CIDR = &cidr
	}
}

func (e *SecurityGroupRule) matches(rule *ec2.SecurityGroupRule) bool {
	matchFromPort := int64(-1)
	if e.FromPort != nil {
		matchFromPort = *e.FromPort
	}
	if aws.Int64Value(rule.FromPort) != matchFromPort {
		return false
	}

	matchToPort := int64(-1)
	if e.ToPort != nil {
		matchToPort = *e.ToPort
	}
	if aws.Int64Value(rule.ToPort) != matchToPort {
		return false
	}

	matchProtocol := "-1" // Wildcard
	if e.Protocol != nil {
		matchProtocol = *e.Protocol
	}
	if aws.StringValue(rule.IpProtocol) != matchProtocol {
		return false
	}

	if fi.StringValue(e.CIDR) != fi.StringValue(rule.CidrIpv4) {
		return false
	}

	if fi.StringValue(e.IPv6CIDR) != fi.StringValue(rule.CidrIpv6) {
		return false
	}

	if fi.StringValue(e.PrefixList) != fi.StringValue(rule.PrefixListId) {
		return false
	}

	if e.SourceGroup != nil || rule.ReferencedGroupInfo != nil {
		if e.SourceGroup == nil || rule.ReferencedGroupInfo == nil {
			return false
		}
		if fi.StringValue(e.SourceGroup.ID) != fi.StringValue(rule.ReferencedGroupInfo.GroupId) {
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
		if e.CIDR != nil && e.IPv6CIDR != nil {
			return field.Forbidden(field.NewPath("CIDR/IPv6CIDR"), "Cannot set more than 1 CIDR or IPv6CIDR")
		}
		if e.PrefixList != nil && (e.CIDR != nil || e.IPv6CIDR != nil) {
			return field.Forbidden(field.NewPath("PrefixList"), "Cannot set PrefixList when CIDR or IPv6CIDR is set")
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

	if e.IPv6CIDR != nil {
		description = append(description, fmt.Sprintf("ipv6cidr=%s", *e.IPv6CIDR))
	}

	if e.PrefixList != nil {
		description = append(description, fmt.Sprintf("prefixList=%s", *e.PrefixList))
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
		} else if e.IPv6CIDR != nil {
			IPv6CIDR := e.IPv6CIDR
			ipPermission.Ipv6Ranges = []*ec2.Ipv6Range{
				{CidrIpv6: IPv6CIDR},
			}
		} else if e.CIDR != nil {
			CIDR := e.CIDR
			ipPermission.IpRanges = []*ec2.IpRange{
				{CidrIp: CIDR},
			}
		} else if e.PrefixList != nil {
			PrefixList := e.PrefixList
			ipPermission.PrefixListIds = []*ec2.PrefixListId{
				{PrefixListId: PrefixList},
			}
		} else {
			ipPermission.IpRanges = []*ec2.IpRange{
				{CidrIp: aws.String("0.0.0.0/0")},
			}
		}

		description := e.Description()

		if fi.BoolValue(e.Egress) {
			request := &ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: e.SecurityGroup.ID,
			}
			request.IpPermissions = []*ec2.IpPermission{ipPermission}
			request.TagSpecifications = awsup.EC2TagSpecification(ec2.ResourceTypeSecurityGroupRule, e.Tags)

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
			request.TagSpecifications = awsup.EC2TagSpecification(ec2.ResourceTypeSecurityGroupRule, e.Tags)

			klog.V(2).Infof("%s: Calling EC2 AuthorizeSecurityGroupIngress (%s)", name, description)
			_, err := t.Cloud.EC2().AuthorizeSecurityGroupIngress(request)
			if err != nil {
				return fmt.Errorf("error creating SecurityGroupIngress: %v", err)
			}
		}

	} else if changes.Tags != nil {
		return t.AddAWSTags(*a.ID, e.Tags)
	}

	// No tags on security group rules (there are tags on the group though)

	return nil
}

type terraformSecurityGroupIngress struct {
	Type *string `cty:"type"`

	SecurityGroup *terraformWriter.Literal `cty:"security_group_id"`
	SourceGroup   *terraformWriter.Literal `cty:"source_security_group_id"`

	FromPort *int64 `cty:"from_port"`
	ToPort   *int64 `cty:"to_port"`

	Protocol       *string  `cty:"protocol"`
	CIDRBlocks     []string `cty:"cidr_blocks"`
	IPv6CIDRBlocks []string `cty:"ipv6_cidr_blocks"`
	PrefixListIDs  []string `cty:"prefix_list_ids"`
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
	if e.IPv6CIDR != nil {
		tf.IPv6CIDRBlocks = append(tf.IPv6CIDRBlocks, *e.IPv6CIDR)
	}
	if e.PrefixList != nil {
		tf.PrefixListIDs = append(tf.PrefixListIDs, *e.PrefixList)
	}

	return t.RenderResource("aws_security_group_rule", *e.Name, tf)
}

type cloudformationSecurityGroupIngress struct {
	SecurityGroup *cloudformation.Literal `json:"GroupId,omitempty"`
	SourceGroup   *cloudformation.Literal `json:"SourceSecurityGroupId,omitempty"`

	FromPort *int64 `json:"FromPort,omitempty"`
	ToPort   *int64 `json:"ToPort,omitempty"`

	Protocol           *string `json:"IpProtocol,omitempty"`
	CidrIp             *string `json:"CidrIp,omitempty"`
	CidrIpv6           *string `json:"CidrIpv6,omitempty"`
	SourcePrefixListId *string `json:"SourcePrefixListId,omitempty"`
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
	if e.IPv6CIDR != nil {
		tf.CidrIpv6 = e.IPv6CIDR
	}
	if e.PrefixList != nil {
		tf.SourcePrefixListId = e.PrefixList
	}

	return t.RenderResource(cfType, *e.Name, tf)
}
