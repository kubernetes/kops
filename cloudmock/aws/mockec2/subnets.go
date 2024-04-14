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

package mockec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

type subnetInfo struct {
	main ec2types.Subnet
}

func (m *MockEC2) FindSubnet(id string) *ec2types.Subnet {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	subnet := m.subnets[id]
	if subnet == nil {
		return nil
	}

	copy := subnet.main
	copy.Tags = m.getTags(ec2types.ResourceTypeSubnet, id)
	return &copy
}

func (m *MockEC2) SubnetIds() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var ids []string
	for id := range m.subnets {
		ids = append(ids, id)
	}
	return ids
}

func (m *MockEC2) CreateSubnetWithId(request *ec2.CreateSubnetInput, id string) (*ec2.CreateSubnetOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	subnet := &ec2types.Subnet{
		SubnetId:         s(id),
		VpcId:            request.VpcId,
		CidrBlock:        request.CidrBlock,
		AvailabilityZone: request.AvailabilityZone,
		PrivateDnsNameOptionsOnLaunch: &ec2types.PrivateDnsNameOptionsOnLaunch{
			EnableResourceNameDnsAAAARecord: aws.Bool(false),
			EnableResourceNameDnsARecord:    aws.Bool(false),
			HostnameType:                    ec2types.HostnameTypeIpName,
		},
	}

	if request.Ipv6CidrBlock != nil {
		subnet.Ipv6CidrBlockAssociationSet = []ec2types.SubnetIpv6CidrBlockAssociation{
			{
				AssociationId: aws.String("subnet-cidr-assoc-ipv6-" + id),
				Ipv6CidrBlock: request.Ipv6CidrBlock,
				Ipv6CidrBlockState: &ec2types.SubnetCidrBlockState{
					State: ec2types.SubnetCidrBlockStateCodeAssociated,
				},
			},
		}
	}

	if m.subnets == nil {
		m.subnets = make(map[string]*subnetInfo)
	}
	m.subnets[*subnet.SubnetId] = &subnetInfo{
		main: *subnet,
	}

	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeSubnet)...)

	response := &ec2.CreateSubnetOutput{
		Subnet: subnet,
	}
	return response, nil
}

func (m *MockEC2) CreateSubnet(ctx context.Context, request *ec2.CreateSubnetInput, optFns ...func(*ec2.Options)) (*ec2.CreateSubnetOutput, error) {
	klog.Infof("CreateSubnet: %v", request)

	id := m.allocateId("subnet")
	return m.CreateSubnetWithId(request, id)
}

func (m *MockEC2) DescribeSubnets(ctx context.Context, request *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeSubnets: %v", request)

	if len(request.SubnetIds) != 0 {
		request.Filters = append(request.Filters, ec2types.Filter{Name: s("subnet-id"), Values: request.SubnetIds})
	}

	var subnets []ec2types.Subnet

	for id, subnet := range m.subnets {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "vpc-id":
				if *subnet.main.VpcId == filter.Values[0] {
					match = true
				}
			case "subnet-id":
				for _, v := range filter.Values {
					if *subnet.main.SubnetId == v {
						match = true
					}
				}
			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2types.ResourceTypeSubnet, *subnet.main.SubnetId, filter)
				} else {
					return nil, fmt.Errorf("unknown filter name: %q", *filter.Name)
				}
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := subnet.main
		copy.Tags = m.getTags(ec2types.ResourceTypeSubnet, id)
		subnets = append(subnets, copy)
	}

	response := &ec2.DescribeSubnetsOutput{
		Subnets: subnets,
	}

	return response, nil
}

func (m *MockEC2) AssociateRouteTable(ctx context.Context, request *ec2.AssociateRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.AssociateRouteTableOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AuthorizeSecurityGroupIngress: %v", request)

	if aws.ToString(request.SubnetId) == "" {
		return nil, fmt.Errorf("SubnetId not specified")
	}
	if aws.ToString(request.RouteTableId) == "" {
		return nil, fmt.Errorf("RouteTableId not specified")
	}

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	subnet := m.subnets[*request.SubnetId]
	if subnet == nil {
		return nil, fmt.Errorf("Subnet not found")
	}
	rt := m.RouteTables[*request.RouteTableId]
	if rt == nil {
		return nil, fmt.Errorf("RouteTable not found")
	}

	associationID := m.allocateId("rta")

	rt.Associations = append(rt.Associations, ec2types.RouteTableAssociation{
		RouteTableId:            rt.RouteTableId,
		SubnetId:                subnet.main.SubnetId,
		RouteTableAssociationId: &associationID,
	})
	// TODO: More fields
	// // Indicates whether this is the main route table.
	// Main *bool `locationName:"main" type:"boolean"`

	// TODO: We need to fold permissions

	response := &ec2.AssociateRouteTableOutput{
		AssociationId: &associationID,
	}
	return response, nil
}

func (m *MockEC2) DeleteSubnet(ctx context.Context, request *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteSubnet: %v", request)

	id := aws.ToString(request.SubnetId)
	o := m.subnets[id]
	if o == nil {
		return nil, fmt.Errorf("Subnet %q not found", id)
	}
	delete(m.subnets, id)

	return &ec2.DeleteSubnetOutput{}, nil
}

func (m *MockEC2) ModifySubnetAttribute(ctx context.Context, request *ec2.ModifySubnetAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifySubnetAttributeOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	subnet := m.subnets[*request.SubnetId]
	if request.AssignIpv6AddressOnCreation != nil {
		subnet.main.AssignIpv6AddressOnCreation = request.AssignIpv6AddressOnCreation.Value
	}
	if request.EnableResourceNameDnsAAAARecordOnLaunch != nil {
		subnet.main.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord = request.EnableResourceNameDnsAAAARecordOnLaunch.Value
	}
	if request.EnableResourceNameDnsARecordOnLaunch != nil {
		subnet.main.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord = request.EnableResourceNameDnsARecordOnLaunch.Value
	}
	if len(request.PrivateDnsHostnameTypeOnLaunch) > 0 {
		subnet.main.PrivateDnsNameOptionsOnLaunch.HostnameType = request.PrivateDnsHostnameTypeOnLaunch
	}
	return &ec2.ModifySubnetAttributeOutput{}, nil
}
