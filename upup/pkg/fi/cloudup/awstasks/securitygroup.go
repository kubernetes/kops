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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type SecurityGroup struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID          *string
	Description *string
	VPC         *VPC

	RemoveExtraRules []string

	// Shared is set if this is a shared security group (one we don't create or own)
	Shared *bool

	Tags map[string]string
}

var (
	_ fi.CompareWithID            = &SecurityGroup{}
	_ fi.CloudupProducesDeletions = &SecurityGroup{}
)

func (e *SecurityGroup) CompareWithID() *string {
	return e.ID
}

// OrderSecurityGroupsById implements sort.Interface for []SecurityGroup, based on ID
type OrderSecurityGroupsById []*SecurityGroup

func (a OrderSecurityGroupsById) Len() int      { return len(a) }
func (a OrderSecurityGroupsById) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderSecurityGroupsById) Less(i, j int) bool {
	return fi.ValueOf(a[i].ID) < fi.ValueOf(a[j].ID)
}

func (e *SecurityGroup) Find(c *fi.CloudupContext) (*SecurityGroup, error) {
	sg, err := e.findEc2(c)
	if err != nil {
		return nil, err
	}
	if sg == nil {
		return nil, nil
	}
	actual := &SecurityGroup{
		ID:          sg.GroupId,
		Name:        sg.GroupName,
		Description: sg.Description,
		VPC:         &VPC{ID: sg.VpcId},
		Tags:        intersectTags(sg.Tags, e.Tags),
	}

	klog.V(2).Infof("found matching SecurityGroup %q", *actual.ID)
	e.ID = actual.ID

	actual.RemoveExtraRules = e.RemoveExtraRules

	// Prevent spurious comparison failures
	actual.Shared = e.Shared
	actual.Lifecycle = e.Lifecycle
	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (e *SecurityGroup) findEc2(c *fi.CloudupContext) (*ec2types.SecurityGroup, error) {
	ctx := c.Context()
	cloud := awsup.GetCloud(c)
	request := &ec2.DescribeSecurityGroupsInput{}

	if fi.ValueOf(e.ID) != "" {
		// Find by ID.
		request.GroupIds = []string{fi.ValueOf(e.ID)}
	} else if fi.ValueOf(e.Name) != "" && e.VPC != nil && e.VPC.ID != nil {
		// Find by filters (name and VPC ID).
		filters := cloud.BuildFilters(e.Name)
		filters = append(filters, awsup.NewEC2Filter("vpc-id", *e.VPC.ID))
		filters = append(filters, awsup.NewEC2Filter("group-name", *e.Name))
		request.Filters = filters

	} else {
		// No reason to try.
		return nil, nil
	}

	response, err := cloud.EC2().DescribeSecurityGroups(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing SecurityGroups: %v", err)
	}
	if response == nil || len(response.SecurityGroups) == 0 {
		return nil, nil
	}

	if len(response.SecurityGroups) != 1 {
		return nil, fmt.Errorf("found multiple SecurityGroups matching tags")
	}
	sg := response.SecurityGroups[0]
	return &sg, nil
}

func (e *SecurityGroup) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *SecurityGroup) ShouldCreate(a, e, changes *SecurityGroup) (bool, error) {
	if fi.ValueOf(e.Shared) {
		return false, nil
	}
	return true, nil
}

func (_ *SecurityGroup) CheckChanges(a, e, changes *SecurityGroup) error {
	if a != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil && !fi.ValueOf(e.Shared) {
			return fi.CannotChangeField("Name")
		}
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}
	return nil
}

func (_ *SecurityGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SecurityGroup) error {
	ctx := context.TODO()
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Do we want to do any verification of the security group?
		return nil
	}

	if a == nil {
		klog.V(2).Infof("Creating SecurityGroup with Name:%q VPC:%q", *e.Name, *e.VPC.ID)

		request := &ec2.CreateSecurityGroupInput{
			VpcId:             e.VPC.ID,
			GroupName:         e.Name,
			Description:       e.Description,
			TagSpecifications: awsup.EC2TagSpecification(ec2types.ResourceTypeSecurityGroup, e.Tags),
		}

		response, err := t.Cloud.EC2().CreateSecurityGroup(ctx, request)
		if err != nil {
			return fmt.Errorf("error creating SecurityGroup: %v", err)
		}

		e.ID = response.GroupId
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

type terraformSecurityGroup struct {
	Name        *string                  `cty:"name"`
	VPCID       *terraformWriter.Literal `cty:"vpc_id"`
	Description *string                  `cty:"description"`
	Tags        map[string]string        `cty:"tags"`
}

func (_ *SecurityGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SecurityGroup) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed
		return nil
	}

	tf := &terraformSecurityGroup{
		Name:        e.Name,
		VPCID:       e.VPC.TerraformLink(),
		Description: e.Description,
		Tags:        e.Tags,
	}

	return t.RenderResource("aws_security_group", *e.Name, tf)
}

func (e *SecurityGroup) TerraformLink() *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed
		if e.ID != nil {
			return terraformWriter.LiteralFromStringValue(*e.ID)
		} else {
			klog.Warningf("ID not set on shared subnet %v", e)
		}
	}

	return terraformWriter.LiteralProperty("aws_security_group", *e.Name, "id")
}

// deleteSecurityGroupRule tracks a securitygrouprule that we're going to delete
// It implements fi.CloudupDeletion
type deleteSecurityGroupRule struct {
	rule *ec2types.SecurityGroupRule
}

func buildDeleteSecurityGroupRule(rule ec2types.SecurityGroupRule) *deleteSecurityGroupRule {
	d := &deleteSecurityGroupRule{}
	d.rule = &rule
	return d
}

var _ fi.CloudupDeletion = &deleteSecurityGroupRule{}

func (d *deleteSecurityGroupRule) Delete(t fi.CloudupTarget) error {
	ctx := context.TODO()
	klog.V(2).Infof("deleting security group permission: %v", fi.DebugAsJsonString(d.rule))

	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	if aws.ToBool(d.rule.IsEgress) {
		request := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:              d.rule.GroupId,
			SecurityGroupRuleIds: []string{fi.ValueOf(d.rule.SecurityGroupRuleId)},
		}

		klog.V(2).Infof("Calling EC2 RevokeSecurityGroupEgress")
		_, err := awsTarget.Cloud.EC2().RevokeSecurityGroupEgress(ctx, request)
		if err != nil {
			return fmt.Errorf("error revoking SecurityGroupEgress: %v", err)
		}
	} else {
		request := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:              d.rule.GroupId,
			SecurityGroupRuleIds: []string{fi.ValueOf(d.rule.SecurityGroupRuleId)},
		}

		klog.V(2).Infof("Calling EC2 RevokeSecurityGroupIngress")
		_, err := awsTarget.Cloud.EC2().RevokeSecurityGroupIngress(ctx, request)
		if err != nil {
			return fmt.Errorf("error revoking SecurityGroupIngress: %v", err)
		}
	}

	return nil
}

func (d *deleteSecurityGroupRule) TaskName() string {
	return "SecurityGroupRule"
}

func (d *deleteSecurityGroupRule) Item() string {
	s := fi.ValueOf(d.rule.GroupId) + ":"
	p := d.rule
	if aws.ToInt32(p.FromPort) != 0 {
		s += fmt.Sprintf(" port=%d", aws.ToInt32(p.FromPort))
		if aws.ToInt32(p.ToPort) != aws.ToInt32(p.FromPort) {
			s += fmt.Sprintf("-%d", aws.ToInt32(p.ToPort))
		}
	}
	if aws.ToString(p.IpProtocol) != "-1" {
		s += fmt.Sprintf(" protocol=%s", aws.ToString(p.IpProtocol))
	}
	if p.ReferencedGroupInfo != nil {
		s += fmt.Sprintf(" group=%s", aws.ToString(p.ReferencedGroupInfo.GroupId))
	}
	s += fmt.Sprintf(" ip=%s", aws.ToString(p.CidrIpv4))
	s += fmt.Sprintf(" ipv6=%s", aws.ToString(p.CidrIpv6))
	// permissionString := fi.DebugAsJsonString(d.permission)
	// s += permissionString

	return s
}

func (d *deleteSecurityGroupRule) DeferDeletion() bool {
	return true
}

func (e *SecurityGroup) FindDeletions(c *fi.CloudupContext) ([]fi.CloudupDeletion, error) {
	ctx := c.Context()
	var removals []fi.CloudupDeletion

	if len(e.RemoveExtraRules) == 0 {
		return nil, nil
	}

	var rules []RemovalRule
	for _, s := range e.RemoveExtraRules {
		rule, err := ParseRemovalRule(s)
		if err != nil {
			return nil, fmt.Errorf("cannot parse rule %q: %v", s, err)
		}
		rules = append(rules, rule)
	}

	sg, err := e.findEc2(c)
	if err != nil {
		return nil, err
	}
	if sg == nil {
		return nil, nil
	}

	cloud := awsup.GetCloud(c)

	request := &ec2.DescribeSecurityGroupRulesInput{
		Filters: []ec2types.Filter{
			awsup.NewEC2Filter("group-id", *e.ID),
		},
	}

	response, err := cloud.EC2().DescribeSecurityGroupRules(ctx, request)
	if err != nil {
		return nil, err
	}

	for _, permission := range response.SecurityGroupRules {
		// Because of #478, we can't remove all non-matching security groups
		// Instead we consider only certain rules to be 'in-scope'
		// (in the model, we typically consider only rules on port 22 and 443)
		match := false
		for _, rule := range rules {
			if rule.Matches(&permission) {
				klog.V(2).Infof("permission matches rule %s: %v", rule, permission)
				match = true
				break
			}
		}
		if !match {
			klog.V(4).Infof("Ignoring security group permission %+v (did not match removal rules)", permission)
			continue
		}
		found := false
		for _, t := range c.AllTasks() {
			er, ok := t.(*SecurityGroupRule)
			if !ok {
				continue
			}

			if er.SourceGroup != nil && er.SourceGroup.ID == nil {
				klog.V(4).Infof("Deletion skipping find of SecurityGroupRule %s, because SourceGroup was not found", fi.ValueOf(er.Name))
				return nil, nil
			}

			if er.matches(&permission) {
				found = true
			}
		}
		if !found {
			removals = append(removals, buildDeleteSecurityGroupRule(permission))
		}
	}

	return removals, nil
}

// RemovalRule is a rule that filters the permissions we should remove
type RemovalRule interface {
	Matches(permission *ec2types.SecurityGroupRule) bool
}

// ParseRemovalRule parses our removal rule DSL into a RemovalRule
func ParseRemovalRule(rule string) (RemovalRule, error) {
	rule = strings.TrimSpace(rule)
	tokens := strings.Split(rule, "=")

	// Simple little language:
	//   port=N matches rules that filter (only) by port=N
	//
	// Note this language is internal, so isn't required to be stable

	if len(tokens) == 2 {
		if tokens[0] == "port" {
			ports := strings.SplitN(tokens[1], ":", 2)
			fromPort, err := strconv.Atoi(ports[0])
			if err != nil {
				return nil, fmt.Errorf("cannot parse rule %q", rule)
			}
			toPort := fromPort
			if len(ports) > 1 {
				toPort, err = strconv.Atoi(ports[1])
				if err != nil {
					return nil, fmt.Errorf("cannot parse rule %q", rule)
				}
			}

			return &PortRemovalRule{
				FromPort: fromPort,
				ToPort:   toPort,
			}, nil
		} else {
			return nil, fmt.Errorf("cannot parse rule %q", rule)
		}
	}
	return nil, fmt.Errorf("cannot parse rule %q", rule)
}

type PortRemovalRule struct {
	FromPort int
	ToPort   int
}

var _ RemovalRule = &PortRemovalRule{}

func (r *PortRemovalRule) String() string {
	return fi.DebugAsJsonString(r)
}

func (r *PortRemovalRule) Matches(permission *ec2types.SecurityGroupRule) bool {
	// Check if port matches
	if permission.FromPort == nil || *permission.FromPort != int32(r.FromPort) {
		return false
	}
	if permission.ToPort == nil || *permission.ToPort != int32(r.ToPort) {
		return false
	}
	return true
}
