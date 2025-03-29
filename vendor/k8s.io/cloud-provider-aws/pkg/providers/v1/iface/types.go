package iface

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

// EC2 is an abstraction over AWS', to allow mocking/other implementations
// Note that the DescribeX functions return a list, so callers don't need to deal with paging
// TODO: Should we rename this to AWS (EBS & ELB are not technically part of EC2)
type EC2 interface {
	// Query EC2 for instances matching the filter
	DescribeInstances(request *ec2.DescribeInstancesInput) ([]*ec2.Instance, error)

	DescribeSecurityGroups(request *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error)

	CreateSecurityGroup(*ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)
	DeleteSecurityGroup(request *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error)

	AuthorizeSecurityGroupIngress(*ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupIngress(*ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error)

	DescribeSubnets(*ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error)

	DescribeAvailabilityZones(request *ec2.DescribeAvailabilityZonesInput) ([]*ec2.AvailabilityZone, error)

	CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)
	DeleteTags(input *ec2.DeleteTagsInput) (*ec2.DeleteTagsOutput, error)

	DescribeRouteTables(request *ec2.DescribeRouteTablesInput) ([]*ec2.RouteTable, error)
	CreateRoute(request *ec2.CreateRouteInput) (*ec2.CreateRouteOutput, error)
	DeleteRoute(request *ec2.DeleteRouteInput) (*ec2.DeleteRouteOutput, error)

	ModifyInstanceAttribute(request *ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error)

	DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error)

	DescribeNetworkInterfaces(input *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error)
}
