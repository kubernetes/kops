package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"github.com/kopeio/aws-controller/pkg/kope/utils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=SecurityGroup
type SecurityGroup struct {
	Name *string

	ID          *string
	Description *string
	VPC         *VPC

	RemoveExtraRules *bool
}

var _ fi.CompareWithID = &SecurityGroup{}
var _ fi.ProducesDeletions = &SecurityGroup{}

func (e *SecurityGroup) CompareWithID() *string {
	return e.ID
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

	return actual, nil
}

func (e *SecurityGroup) findEc2(c *fi.Context) (*ec2.SecurityGroup, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

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

		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2.DescribeSecurityGroups(request)
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

func (_ *SecurityGroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SecurityGroup) error {
	if a == nil {
		glog.V(2).Infof("Creating SecurityGroup with Name:%q VPC:%q", *e.Name, *e.VPC.ID)

		request := &ec2.CreateSecurityGroupInput{
			VpcId:       e.VPC.ID,
			GroupName:   e.Name,
			Description: e.Description,
		}

		response, err := t.Cloud.EC2.CreateSecurityGroup(request)
		if err != nil {
			return fmt.Errorf("error creating SecurityGroup: %v", err)
		}

		e.ID = response.GroupId
	}

	return t.AddAWSTags(*e.ID, t.Cloud.BuildTags(e.Name))
}

type terraformSecurityGroup struct {
	Name        *string            `json:"name"`
	VPCID       *terraform.Literal `json:"vpc_id"`
	Description *string            `json:"description"`
	Tags        map[string]string  `json:"tags,omitempty"`
}

func (_ *SecurityGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *SecurityGroup) error {
	cloud := t.Cloud.(*awsup.AWSCloud)

	tf := &terraformSecurityGroup{
		Name:        e.Name,
		VPCID:       e.VPC.TerraformLink(),
		Description: e.Description,
		Tags:        cloud.BuildTags(e.Name),
	}

	return t.RenderResource("aws_security_group", *e.Name, tf)
}

func (e *SecurityGroup) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_security_group", *e.Name, "id")
}

type deleteSecurityGroupRule struct {
	groupID    *string
	permission *ec2.IpPermission
	egress     bool
}

var _ fi.Deletion = &deleteSecurityGroupRule{}

func (d *deleteSecurityGroupRule) Delete(t fi.Target) error {
	glog.V(2).Infof("deleting security group permission: %v", utils.DebugString(d.permission))

	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	if d.egress {
		request := &ec2.RevokeSecurityGroupEgressInput{
			GroupId: d.groupID,
		}
		request.IpPermissions = []*ec2.IpPermission{d.permission}

		glog.V(2).Infof("Calling EC2 RevokeSecurityGroupEgress")
		_, err := awsTarget.Cloud.EC2.RevokeSecurityGroupEgress(request)
		if err != nil {
			return fmt.Errorf("error revoking SecurityGroupEgress: %v", err)
		}
	} else {
		request := &ec2.RevokeSecurityGroupIngressInput{
			GroupId: d.groupID,
		}
		request.IpPermissions = []*ec2.IpPermission{d.permission}

		glog.V(2).Infof("Calling EC2 RevokeSecurityGroupIngress")
		_, err := awsTarget.Cloud.EC2.RevokeSecurityGroupIngress(request)
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
	//permissionString := utils.DebugString(d.permission)
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

	if fi.BoolValue(e.RemoveExtraRules) != true {
		return nil, nil
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
