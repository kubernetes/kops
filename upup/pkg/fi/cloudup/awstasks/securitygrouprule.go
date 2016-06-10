package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=SecurityGroupRule
type SecurityGroupRule struct {
	Name *string

	SecurityGroup *SecurityGroup
	CIDR          *string
	Protocol      *string
	FromPort      *int64
	ToPort        *int64
	SourceGroup   *SecurityGroup

	Egress *bool
}

func (e *SecurityGroupRule) Find(c *fi.Context) (*SecurityGroupRule, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	if e.SecurityGroup == nil || e.SecurityGroup.ID == nil {
		return nil, nil
	}

	request := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			awsup.NewEC2Filter("group-id", *e.SecurityGroup.ID),
		},
	}

	response, err := cloud.EC2.DescribeSecurityGroups(request)
	if err != nil {
		return nil, fmt.Errorf("error listing SecurityGroup: %v", err)
	}

	if response == nil || len(response.SecurityGroups) == 0 {
		return nil, nil
	}

	if len(response.SecurityGroups) != 1 {
		glog.Fatalf("found multiple security groups for id=%s", *e.SecurityGroup.ID)
	}
	sg := response.SecurityGroups[0]
	//glog.V(2).Info("found existing security group")

	var foundRule *ec2.IpPermission

	matchProtocol := "-1" // Wildcard
	if e.Protocol != nil {
		matchProtocol = *e.Protocol
	}

	ipPermissions := sg.IpPermissions
	if fi.BoolValue(e.Egress) {
		ipPermissions = sg.IpPermissionsEgress
	}

	for _, rule := range ipPermissions {
		if aws.Int64Value(rule.FromPort) != aws.Int64Value(e.FromPort) {
			continue
		}
		if aws.Int64Value(rule.ToPort) != aws.Int64Value(e.ToPort) {
			continue
		}
		if aws.StringValue(rule.IpProtocol) != matchProtocol {
			continue
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
				continue
			}
		}

		if e.SourceGroup != nil {
			// TODO: Only if len 1?
			match := false
			for _, spec := range rule.UserIdGroupPairs {
				if aws.StringValue(spec.GroupId) == *e.SourceGroup.ID {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		foundRule = rule
		break
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
		return actual, nil
	}

	return nil, nil
}

func (e *SecurityGroupRule) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *SecurityGroupRule) CheckChanges(a, e, changes *SecurityGroupRule) error {
	if a == nil {
		if e.SecurityGroup == nil {
			return fi.RequiredField("SecurityGroup")
		}
	}
	return nil
}

func (_ *SecurityGroupRule) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SecurityGroupRule) error {
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

		if fi.BoolValue(e.Egress) {
			request := &ec2.AuthorizeSecurityGroupEgressInput{
				GroupId: e.SecurityGroup.ID,
			}
			request.IpPermissions = []*ec2.IpPermission{ipPermission}

			glog.V(2).Infof("Calling EC2 AuthorizeSecurityGroupEgress")
			_, err := t.Cloud.EC2.AuthorizeSecurityGroupEgress(request)
			if err != nil {
				return fmt.Errorf("error creating SecurityGroupEgress: %v", err)
			}
		} else {
			request := &ec2.AuthorizeSecurityGroupIngressInput{
				GroupId: e.SecurityGroup.ID,
			}
			request.IpPermissions = []*ec2.IpPermission{ipPermission}

			glog.V(2).Infof("Calling EC2 AuthorizeSecurityGroupIngress")
			_, err := t.Cloud.EC2.AuthorizeSecurityGroupIngress(request)
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
