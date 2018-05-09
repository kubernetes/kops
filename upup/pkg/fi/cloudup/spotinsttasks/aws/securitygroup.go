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

package aws

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=SecurityGroup
type SecurityGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID          *string
	Description *string
	VPC         *VPC

	RemoveExtraRules []string

	// Shared is set if this is a shared security group (one we don't create or own)
	Shared *bool
}

var _ fi.CompareWithID = &SecurityGroup{}
var _ fi.ProducesDeletions = &SecurityGroup{}

func (e *SecurityGroup) CompareWithID() *string {
	return e.ID
}

// OrderSecurityGroupsById implements sort.Interface for []SecurityGroup, based on ID
type OrderSecurityGroupsById []*SecurityGroup

func (a OrderSecurityGroupsById) Len() int      { return len(a) }
func (a OrderSecurityGroupsById) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderSecurityGroupsById) Less(i, j int) bool {
	return fi.StringValue(a[i].ID) < fi.StringValue(a[j].ID)
}

func (e *SecurityGroup) Find(c *fi.Context) (*SecurityGroup, error) {
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
	}

	glog.V(2).Infof("found matching SecurityGroup %q", *actual.ID)
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

func (e *SecurityGroup) findEc2(c *fi.Context) (*ec2.SecurityGroup, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	var vpcID *string
	if e.VPC != nil {
		vpcID = e.VPC.ID
	}

	if vpcID == nil {
		return nil, nil
	}

	request := &ec2.DescribeSecurityGroupsInput{}

	if fi.StringValue(e.ID) != "" {
		request.GroupIds = []*string{e.ID}
	} else {
		filters := cloud.BuildFilters(e.Name)
		filters = append(filters, awsup.NewEC2Filter("vpc-id", *vpcID))
		filters = append(filters, awsup.NewEC2Filter("group-name", *e.Name))

		request.Filters = filters
	}

	response, err := cloud.EC2().DescribeSecurityGroups(request)
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
	return sg, nil
}

func (e *SecurityGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *SecurityGroup) ShouldCreate(a, e, changes *SecurityGroup) (bool, error) {
	if fi.BoolValue(e.Shared) {
		return false, nil
	}
	return true, nil
}

func (_ *SecurityGroup) CheckChanges(a, e, changes *SecurityGroup) error {
	if a != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}
	return nil
}

func (_ *SecurityGroup) Render(t *spotinst.Target, a, e, changes *SecurityGroup) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Do we want to do any verification of the security group?
		return nil
	}

	if a == nil {
		glog.V(2).Infof("Creating SecurityGroup with Name:%q VPC:%q", *e.Name, *e.VPC.ID)

		request := &ec2.CreateSecurityGroupInput{
			VpcId:       e.VPC.ID,
			GroupName:   e.Name,
			Description: e.Description,
		}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateSecurityGroup(request)
		if err != nil {
			return fmt.Errorf("error creating SecurityGroup: %v", err)
		}

		e.ID = response.GroupId
	}

	return t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.ID, t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).BuildTags(e.Name))
}

type deleteSecurityGroupRule struct {
	groupID    *string
	permission *ec2.IpPermission
	egress     bool
}

var _ fi.Deletion = &deleteSecurityGroupRule{}

func (d *deleteSecurityGroupRule) Delete(t fi.Target) error {
	glog.V(2).Infof("deleting security group permission: %v", fi.DebugAsJsonString(d.permission))

	target, ok := t.(*spotinst.Target)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	if d.egress {
		request := &ec2.RevokeSecurityGroupEgressInput{
			GroupId: d.groupID,
		}
		request.IpPermissions = []*ec2.IpPermission{d.permission}

		glog.V(2).Infof("Calling EC2 RevokeSecurityGroupEgress")
		_, err := target.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().RevokeSecurityGroupEgress(request)
		if err != nil {
			return fmt.Errorf("error revoking SecurityGroupEgress: %v", err)
		}
	} else {
		request := &ec2.RevokeSecurityGroupIngressInput{
			GroupId: d.groupID,
		}
		request.IpPermissions = []*ec2.IpPermission{d.permission}

		glog.V(2).Infof("Calling EC2 RevokeSecurityGroupIngress")
		_, err := target.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().RevokeSecurityGroupIngress(request)
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
	s := fi.StringValue(d.groupID) + ":"
	p := d.permission
	if aws.Int64Value(p.FromPort) != 0 {
		s += fmt.Sprintf(" port=%d", aws.Int64Value(p.FromPort))
		if aws.Int64Value(p.ToPort) != aws.Int64Value(p.FromPort) {
			s += fmt.Sprintf("-%d", aws.Int64Value(p.ToPort))
		}
	}
	if aws.StringValue(p.IpProtocol) != "-1" {
		s += fmt.Sprintf(" protocol=%s", aws.StringValue(p.IpProtocol))
	}
	for _, ug := range p.UserIdGroupPairs {
		s += fmt.Sprintf(" group=%s", aws.StringValue(ug.GroupId))
	}
	for _, r := range p.IpRanges {
		s += fmt.Sprintf(" ip=%s", aws.StringValue(r.CidrIp))
	}
	//permissionString := fi.DebugAsJsonString(d.permission)
	//s += permissionString

	return s
}

func expandPermissions(sgID *string, permission *ec2.IpPermission, egress bool) []*ec2.IpPermission {
	var rules []*ec2.IpPermission

	master := &ec2.IpPermission{
		FromPort:   permission.FromPort,
		ToPort:     permission.ToPort,
		IpProtocol: permission.IpProtocol,
	}

	for _, ipRange := range permission.IpRanges {
		a := &ec2.IpPermission{}
		*a = *master
		a.IpRanges = []*ec2.IpRange{ipRange}
		rules = append(rules, a)
	}

	for _, ug := range permission.UserIdGroupPairs {
		a := &ec2.IpPermission{}
		*a = *master
		a.UserIdGroupPairs = []*ec2.UserIdGroupPair{ug}
		rules = append(rules, a)
	}

	if len(rules) == 0 {
		// If there are no group or cidr restrictions, it is just a generic rule
		rules = append(rules, master)
	}

	return rules
}

func (e *SecurityGroup) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	var removals []fi.Deletion

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

	var ingress []*ec2.IpPermission
	for _, permission := range sg.IpPermissions {
		rules := expandPermissions(sg.GroupId, permission, false)
		ingress = append(ingress, rules...)
	}

	for _, permission := range ingress {
		// Because of #478, we can't remove all non-matching security groups
		// Instead we consider only certain rules to be 'in-scope'
		// (in the model, we typically consider only rules on port 22 and 443)
		match := false
		for _, rule := range rules {
			if rule.Matches(permission) {
				glog.V(2).Infof("permission matches rule %s: %v", rule, permission)
				match = true
				break
			}
		}
		if !match {
			glog.V(4).Infof("Ignoring security group permission %q (did not match removal rules)", permission)
			continue
		}
		found := false
		for _, t := range c.AllTasks() {
			er, ok := t.(*SecurityGroupRule)
			if !ok {
				continue
			}

			if er.SourceGroup != nil && er.SourceGroup.ID == nil {
				glog.V(4).Infof("Deletion skipping find of SecurityGroupRule %s, because SourceGroup was not found", fi.StringValue(er.Name))
				return nil, nil
			}

			if er.matches(permission) {
				found = true
			}
		}
		if !found {
			removals = append(removals, &deleteSecurityGroupRule{
				groupID:    sg.GroupId,
				permission: permission,
				egress:     false,
			})
		}
	}

	var egress []*ec2.IpPermission
	for _, permission := range sg.IpPermissionsEgress {
		rules := expandPermissions(sg.GroupId, permission, true)
		egress = append(egress, rules...)
	}
	for _, permission := range egress {
		found := false
		for _, t := range c.AllTasks() {
			er, ok := t.(*SecurityGroupRule)
			if !ok {
				continue
			}
			if er.matches(permission) {
				found = true
			}
		}
		if !found {
			removals = append(removals, &deleteSecurityGroupRule{
				groupID:    sg.GroupId,
				permission: permission,
				egress:     true,
			})
		}
	}

	return removals, nil
}

// RemovalRule is a rule that filters the permissions we should remove
type RemovalRule interface {
	Matches(permission *ec2.IpPermission) bool
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
			port, err := strconv.Atoi(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("cannot parse rule %q", rule)
			}

			return &PortRemovalRule{Port: port}, nil
		} else {
			return nil, fmt.Errorf("cannot parse rule %q", rule)
		}
	}
	return nil, fmt.Errorf("cannot parse rule %q", rule)
}

type PortRemovalRule struct {
	Port int
}

var _ RemovalRule = &PortRemovalRule{}

func (r *PortRemovalRule) String() string {
	return fi.DebugAsJsonString(r)
}

func (r *PortRemovalRule) Matches(permission *ec2.IpPermission) bool {
	// Check if port matches
	if permission.FromPort == nil || *permission.FromPort != int64(r.Port) {
		return false
	}
	if permission.ToPort == nil || *permission.ToPort != int64(r.Port) {
		return false
	}
	return true
}
