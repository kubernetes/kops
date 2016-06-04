package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
)

type SecurityGroupIngress struct {
	SecurityGroup *SecurityGroup
	CIDR          *string
	Protocol      *string
	FromPort      *int64
	ToPort        *int64
	SourceGroup   *SecurityGroup
}

func (e *SecurityGroupIngress) String() string {
	return fi.TaskAsString(e)
}

//func (s *SecurityGroupIngress) Key() string {
//	key := s.SecurityGroup.Key()
//	if s.Protocol != nil {
//		key += "-" + *s.Protocol
//	}
//	if s.FromPort != nil {
//		key += "-" + strconv.FormatInt(*s.FromPort, 10)
//	}
//	if s.ToPort != nil {
//		key += "-" + strconv.FormatInt(*s.ToPort, 10)
//	}
//	if s.CIDR != nil {
//		key += "-" + *s.CIDR
//	}
//	if s.SourceGroup != nil {
//		key += "-" + s.SourceGroup.Key()
//	}
//	return key
//}

func (e *SecurityGroupIngress) Find(c *fi.Context) (*SecurityGroupIngress, error) {
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

	for _, rule := range sg.IpPermissions {
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
		actual := &SecurityGroupIngress{
			SecurityGroup: &SecurityGroup{ID: e.SecurityGroup.ID},
			FromPort:      foundRule.FromPort,
			ToPort:        foundRule.ToPort,
			Protocol:      foundRule.IpProtocol,
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

func (e *SecurityGroupIngress) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *SecurityGroupIngress) CheckChanges(a, e, changes *SecurityGroupIngress) error {
	if a == nil {
		if e.SecurityGroup == nil {
			return fi.RequiredField("SecurityGroup")
		}
	}
	return nil
}

func (_ *SecurityGroupIngress) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *SecurityGroupIngress) error {
	if a == nil {
		request := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: e.SecurityGroup.ID,
		}

		protocol := e.Protocol
		if protocol == nil {
			protocol = aws.String("-1")
		}

		if e.SourceGroup != nil {
			request.IpPermissions = []*ec2.IpPermission{
				{
					IpProtocol: protocol,
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						{
							GroupId: e.SourceGroup.ID,
						},
					},
					FromPort: e.FromPort,
					ToPort:   e.ToPort,
				},
			}
		} else {
			request.IpPermissions = []*ec2.IpPermission{
				{
					IpProtocol: protocol,
					FromPort:   e.FromPort,
					ToPort:     e.ToPort,
					IpRanges: []*ec2.IpRange{
						{CidrIp: e.CIDR},
					},
				},
			}
		}

		glog.V(2).Infof("Calling EC2 AuthorizeSecurityGroupIngress")
		_, err := t.Cloud.EC2.AuthorizeSecurityGroupIngress(request)
		if err != nil {
			return fmt.Errorf("error creating SecurityGroupIngress: %v", err)
		}
	}

	// No tags on ingress rules (there are tags on the group though)

	return nil
}
