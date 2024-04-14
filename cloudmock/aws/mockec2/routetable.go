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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

func (m *MockEC2) AddRouteTable(rt *ec2types.RouteTable) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.RouteTables == nil {
		m.RouteTables = make(map[string]*ec2types.RouteTable)
	}

	m.addTags(*rt.RouteTableId, rt.Tags...)

	m.RouteTables[*rt.RouteTableId] = rt
}

func (m *MockEC2) DescribeRouteTables(ctx context.Context, request *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeRouteTables: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun not implemented")
	}

	if len(request.RouteTableIds) != 0 {
		request.Filters = append(request.Filters, ec2types.Filter{Name: s("route-table-id"), Values: request.RouteTableIds})
	}

	response := &ec2.DescribeRouteTablesOutput{}
	for _, rt := range m.RouteTables {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "route-table-id":
				for _, v := range filter.Values {
					if *rt.RouteTableId == v {
						match = true
					}
				}
			case "association.subnet-id":
				for _, a := range rt.Associations {
					for _, v := range filter.Values {
						if aws.ToString(a.SubnetId) == v {
							match = true
						}
					}
				}
			default:
				match = m.hasTag(ec2types.ResourceTypeRouteTable, *rt.RouteTableId, filter)
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := *rt
		copy.Tags = m.getTags(ec2types.ResourceTypeRouteTable, *rt.RouteTableId)
		response.RouteTables = append(response.RouteTables, copy)
	}

	return response, nil
}

func (m *MockEC2) CreateRouteTable(ctx context.Context, request *ec2.CreateRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteTableOutput, error) {
	klog.Infof("CreateRouteTable: %v", request)

	id := m.allocateId("rtb")
	return m.CreateRouteTableWithId(request, id)
}

func (m *MockEC2) CreateRouteTableWithId(request *ec2.CreateRouteTableInput, id string) (*ec2.CreateRouteTableOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	rt := &ec2types.RouteTable{
		RouteTableId: s(id),
		VpcId:        request.VpcId,
	}

	if m.RouteTables == nil {
		m.RouteTables = make(map[string]*ec2types.RouteTable)
	}
	m.RouteTables[id] = rt

	copy := *rt
	response := &ec2.CreateRouteTableOutput{
		RouteTable: &copy,
	}
	return response, nil
}

func (m *MockEC2) CreateRoute(ctx context.Context, request *ec2.CreateRouteInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateRoute: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	rt := m.RouteTables[aws.ToString(request.RouteTableId)]
	if rt == nil {
		return nil, fmt.Errorf("RouteTable not found")
	}

	r := ec2types.Route{
		DestinationCidrBlock:        request.DestinationCidrBlock,
		DestinationIpv6CidrBlock:    request.DestinationIpv6CidrBlock,
		EgressOnlyInternetGatewayId: request.EgressOnlyInternetGatewayId,
		GatewayId:                   request.GatewayId,
		InstanceId:                  request.InstanceId,
		NatGatewayId:                request.NatGatewayId,
		NetworkInterfaceId:          request.NetworkInterfaceId,
		TransitGatewayId:            request.TransitGatewayId,
		VpcPeeringConnectionId:      request.VpcPeeringConnectionId,
	}

	rt.Routes = append(rt.Routes, r)
	return &ec2.CreateRouteOutput{Return: aws.Bool(true)}, nil
}

func (m *MockEC2) DeleteRouteTable(ctx context.Context, request *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteRouteTable: %v", request)

	id := aws.ToString(request.RouteTableId)
	o := m.RouteTables[id]
	if o == nil {
		return nil, fmt.Errorf("RouteTable %q not found", id)
	}
	delete(m.RouteTables, id)

	return &ec2.DeleteRouteTableOutput{}, nil
}
