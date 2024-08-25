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
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// awsSdkEC2 is an implementation of the EC2 interface, backed by aws-sdk-go
type awsSdkEC2 struct {
	ec2 ec2iface.EC2API
}

// Implementation of EC2.Instances
func (s *awsSdkEC2) DescribeInstances(request *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	// Instances are paged
	results := []*ec2.Instance{}
	var nextToken *string
	requestTime := time.Now()

	if request.MaxResults == nil && len(request.InstanceIds) == 0 {
		// MaxResults must be set in order for pagination to work
		// MaxResults cannot be set with InstanceIds
		request.MaxResults = aws.Int64(1000)
	}

	for {
		response, err := s.ec2.DescribeInstances(request)
		if err != nil {
			recordAWSMetric("describe_instance", 0, err)
			return nil, fmt.Errorf("error listing AWS instances: %q", err)
		}

		for _, reservation := range response.Reservations {
			results = append(results, reservation.Instances...)
		}

		nextToken = response.NextToken
		if aws.StringValue(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_instance", timeTaken, nil)
	return results, nil
}

// DescribeNetworkInterfaces describes network interface provided in the input.
func (s *awsSdkEC2) DescribeNetworkInterfaces(input *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	requestTime := time.Now()
	resp, err := s.ec2.DescribeNetworkInterfaces(input)
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_network_interfaces", timeTaken, err)
	return resp, err
}

// Implements EC2.DescribeSecurityGroups
func (s *awsSdkEC2) DescribeSecurityGroups(request *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	// Security groups are paged
	results := []*ec2.SecurityGroup{}
	var nextToken *string
	requestTime := time.Now()
	for {
		response, err := s.ec2.DescribeSecurityGroups(request)
		if err != nil {
			recordAWSMetric("describe_security_groups", 0, err)
			return nil, fmt.Errorf("error listing AWS security groups: %q", err)
		}

		results = append(results, response.SecurityGroups...)

		nextToken = response.NextToken
		if aws.StringValue(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_security_groups", timeTaken, nil)
	return results, nil
}

func (s *awsSdkEC2) DescribeSubnets(request *ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error) {
	// Subnets are not paged
	response, err := s.ec2.DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error listing AWS subnets: %q", err)
	}
	return response.Subnets, nil
}

func (s *awsSdkEC2) DescribeAvailabilityZones(request *ec2.DescribeAvailabilityZonesInput) ([]*ec2.AvailabilityZone, error) {
	// AZs are not paged
	response, err := s.ec2.DescribeAvailabilityZones(request)
	if err != nil {
		return nil, fmt.Errorf("error listing AWS availability zones: %q", err)
	}
	return response.AvailabilityZones, err
}

func (s *awsSdkEC2) CreateSecurityGroup(request *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	return s.ec2.CreateSecurityGroup(request)
}

func (s *awsSdkEC2) DeleteSecurityGroup(request *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	return s.ec2.DeleteSecurityGroup(request)
}

func (s *awsSdkEC2) AuthorizeSecurityGroupIngress(request *ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	return s.ec2.AuthorizeSecurityGroupIngress(request)
}

func (s *awsSdkEC2) RevokeSecurityGroupIngress(request *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	return s.ec2.RevokeSecurityGroupIngress(request)
}

func (s *awsSdkEC2) CreateTags(request *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	requestTime := time.Now()
	resp, err := s.ec2.CreateTags(request)
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("create_tags", timeTaken, err)
	return resp, err
}

func (s *awsSdkEC2) DeleteTags(request *ec2.DeleteTagsInput) (*ec2.DeleteTagsOutput, error) {
	requestTime := time.Now()
	resp, err := s.ec2.DeleteTags(request)
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("delete_tags", timeTaken, err)
	return resp, err
}

func (s *awsSdkEC2) DescribeRouteTables(request *ec2.DescribeRouteTablesInput) ([]*ec2.RouteTable, error) {
	results := []*ec2.RouteTable{}
	var nextToken *string
	requestTime := time.Now()
	for {
		response, err := s.ec2.DescribeRouteTables(request)
		if err != nil {
			recordAWSMetric("describe_route_tables", 0, err)
			return nil, fmt.Errorf("error listing AWS route tables: %q", err)
		}

		results = append(results, response.RouteTables...)

		nextToken = response.NextToken
		if aws.StringValue(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	timeTaken := time.Since(requestTime).Seconds()
	recordAWSMetric("describe_route_tables", timeTaken, nil)
	return results, nil
}

func (s *awsSdkEC2) CreateRoute(request *ec2.CreateRouteInput) (*ec2.CreateRouteOutput, error) {
	return s.ec2.CreateRoute(request)
}

func (s *awsSdkEC2) DeleteRoute(request *ec2.DeleteRouteInput) (*ec2.DeleteRouteOutput, error) {
	return s.ec2.DeleteRoute(request)
}

func (s *awsSdkEC2) ModifyInstanceAttribute(request *ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error) {
	return s.ec2.ModifyInstanceAttribute(request)
}

func (s *awsSdkEC2) DescribeVpcs(request *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	return s.ec2.DescribeVpcs(request)
}
