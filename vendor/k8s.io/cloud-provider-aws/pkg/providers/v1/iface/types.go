package iface

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2 is an abstraction over AWS', to allow mocking/other implementations
// Note that the DescribeX functions return a list, so callers don't need to deal with paging
// TODO: Should we rename this to AWS (EBS & ELB are not technically part of EC2)
type EC2 interface {
	// Query EC2 for instances matching the filter
	DescribeInstances(ctx context.Context, request *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) ([]ec2types.Instance, error)
	DescribeInstanceTopology(ctx context.Context, request *ec2.DescribeInstanceTopologyInput, optFns ...func(*ec2.Options)) ([]ec2types.InstanceTopology, error)

	DescribeSecurityGroups(ctx context.Context, request *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) ([]ec2types.SecurityGroup, error)

	CreateSecurityGroup(ctx context.Context, request *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	DeleteSecurityGroup(ctx context.Context, request *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)

	AuthorizeSecurityGroupIngress(ctx context.Context, request *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupIngress(ctx context.Context, request *ec2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)

	DescribeSubnets(ctx context.Context, request *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) ([]ec2types.Subnet, error)

	DescribeAvailabilityZones(ctx context.Context, request *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) ([]ec2types.AvailabilityZone, error)

	CreateTags(ctx context.Context, request *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, request *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)

	DescribeRouteTables(ctx context.Context, request *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) ([]ec2types.RouteTable, error)
	CreateRoute(ctx context.Context, request *ec2.CreateRouteInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error)
	DeleteRoute(ctx context.Context, request *ec2.DeleteRouteInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteOutput, error)

	ModifyInstanceAttribute(ctx context.Context, request *ec2.ModifyInstanceAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error)

	DescribeVpcs(ctx context.Context, input *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)

	DescribeNetworkInterfaces(ctx context.Context, input *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
}
