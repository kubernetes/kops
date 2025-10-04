/*
Copyright 2024 The Kubernetes Authors.

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
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2API is an interface to satisfy the ec2.Client API.
// More details about this pattern: https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/unit-testing.html
type EC2API interface {
	AuthorizeSecurityGroupIngress(ctx context.Context, params *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	CreateRoute(ctx context.Context, params *ec2.CreateRouteInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error)
	CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteRoute(ctx context.Context, params *ec2.DeleteRouteInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteOutput, error)
	DeleteSecurityGroup(ctx context.Context, params *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	DeleteTags(ctx context.Context, params *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
	DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFuns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeInstanceTopology(ctx context.Context, params *ec2.DescribeInstanceTopologyInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceTopologyOutput, error)
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	ModifyInstanceAttribute(ctx context.Context, params *ec2.ModifyInstanceAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error)
	RevokeSecurityGroupIngress(ctx context.Context, params *ec2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)
}

// awsSdkEC2 is an implementation of the EC2 interface, backed by aws-sdk-go
type awsSdkEC2 struct {
	ec2 EC2API
}

// Implementation of EC2.Instances
func (s *awsSdkEC2) DescribeInstances(ctx context.Context, request *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) ([]ec2types.Instance, error) {
	// Instances are paged
	results := []ec2types.Instance{}
	var nextToken *string
	requestTime := time.Now()

	if request.MaxResults == nil && len(request.InstanceIds) == 0 {
		// MaxResults must be set in order for pagination to work
		// MaxResults cannot be set with InstanceIds
		request.MaxResults = aws.Int32(1000)
	}

	for {
		response, err := s.ec2.DescribeInstances(ctx, request)
		if err != nil {
			recordAWSMetric("describe_instance", 0, err)
			return nil, fmt.Errorf("error listing AWS instances: %q", err)
		}

		for _, reservation := range response.Reservations {
			results = append(results, reservation.Instances...)
		}

		nextToken = response.NextToken
		if aws.ToString(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_instance", timeTaken, nil)
	return results, nil
}

func (s *awsSdkEC2) DescribeInstanceTopology(ctx context.Context, input *ec2.DescribeInstanceTopologyInput, optFns ...func(*ec2.Options)) ([]ec2types.InstanceTopology, error) {
	var topologies []ec2types.InstanceTopology

	paginator := ec2.NewDescribeInstanceTopologyPaginator(s.ec2, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		topologies = append(topologies, output.Instances...)
	}

	return topologies, nil
}

// DescribeNetworkInterfaces describes network interface provided in the input.
func (s *awsSdkEC2) DescribeNetworkInterfaces(ctx context.Context, input *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	requestTime := time.Now()
	resp, err := s.ec2.DescribeNetworkInterfaces(ctx, input)
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_network_interfaces", timeTaken, err)
	return resp, err
}

// Implements EC2.DescribeSecurityGroups
func (s *awsSdkEC2) DescribeSecurityGroups(ctx context.Context, request *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) ([]ec2types.SecurityGroup, error) {
	// Security groups are paged
	results := []ec2types.SecurityGroup{}
	var nextToken *string
	requestTime := time.Now()
	for {
		response, err := s.ec2.DescribeSecurityGroups(ctx, request)
		if err != nil {
			recordAWSMetric("describe_security_groups", 0, err)
			return nil, fmt.Errorf("error listing AWS security groups: %q", err)
		}

		results = append(results, response.SecurityGroups...)

		nextToken = response.NextToken
		if aws.ToString(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_security_groups", timeTaken, nil)
	return results, nil
}

func (s *awsSdkEC2) DescribeSubnets(ctx context.Context, request *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) ([]ec2types.Subnet, error) {
	// Subnets are not paged
	response, err := s.ec2.DescribeSubnets(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing AWS subnets: %q", err)
	}
	return response.Subnets, nil
}

func (s *awsSdkEC2) DescribeAvailabilityZones(ctx context.Context, request *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) ([]ec2types.AvailabilityZone, error) {
	// AZs are not paged
	response, err := s.ec2.DescribeAvailabilityZones(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing AWS availability zones: %q", err)
	}
	return response.AvailabilityZones, err
}

func (s *awsSdkEC2) CreateSecurityGroup(ctx context.Context, request *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
	return s.ec2.CreateSecurityGroup(ctx, request)
}

func (s *awsSdkEC2) DeleteSecurityGroup(ctx context.Context, request *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	return s.ec2.DeleteSecurityGroup(ctx, request)
}

func (s *awsSdkEC2) AuthorizeSecurityGroupIngress(ctx context.Context, request *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	return s.ec2.AuthorizeSecurityGroupIngress(ctx, request)
}

func (s *awsSdkEC2) RevokeSecurityGroupIngress(ctx context.Context, request *ec2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	return s.ec2.RevokeSecurityGroupIngress(ctx, request)
}

func (s *awsSdkEC2) CreateTags(ctx context.Context, request *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	requestTime := time.Now()
	resp, err := s.ec2.CreateTags(ctx, request)
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("create_tags", timeTaken, err)
	return resp, err
}

func (s *awsSdkEC2) DeleteTags(ctx context.Context, request *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error) {
	requestTime := time.Now()
	resp, err := s.ec2.DeleteTags(ctx, request)
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("delete_tags", timeTaken, err)
	return resp, err
}

func (s *awsSdkEC2) DescribeRouteTables(ctx context.Context, request *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) ([]ec2types.RouteTable, error) {
	results := []ec2types.RouteTable{}
	var nextToken *string
	requestTime := time.Now()
	for {
		response, err := s.ec2.DescribeRouteTables(ctx, request)
		if err != nil {
			recordAWSMetric("describe_route_tables", 0, err)
			return nil, fmt.Errorf("error listing AWS route tables: %q", err)
		}

		results = append(results, response.RouteTables...)

		nextToken = response.NextToken
		if aws.ToString(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_route_tables", timeTaken, nil)
	return results, nil
}

func (s *awsSdkEC2) CreateRoute(ctx context.Context, request *ec2.CreateRouteInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error) {
	return s.ec2.CreateRoute(ctx, request)
}

func (s *awsSdkEC2) DeleteRoute(ctx context.Context, request *ec2.DeleteRouteInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteOutput, error) {
	return s.ec2.DeleteRoute(ctx, request)
}

func (s *awsSdkEC2) ModifyInstanceAttribute(ctx context.Context, request *ec2.ModifyInstanceAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error) {
	return s.ec2.ModifyInstanceAttribute(ctx, request)
}

func (s *awsSdkEC2) DescribeVpcs(ctx context.Context, request *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return s.ec2.DescribeVpcs(ctx, request)
}
