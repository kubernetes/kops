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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func (m *MockEC2) AddRouteTable(rt *ec2.RouteTable) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.RouteTables == nil {
		m.RouteTables = make(map[string]*ec2.RouteTable)
	}
	for _, tag := range rt.Tags {
		m.addTag(*rt.RouteTableId, tag)
	}
	rt.Tags = nil
	m.RouteTables[*rt.RouteTableId] = rt
}

func (m *MockEC2) DescribeRouteTablesRequest(*ec2.DescribeRouteTablesInput) (*request.Request, *ec2.DescribeRouteTablesOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeRouteTablesWithContext(aws.Context, *ec2.DescribeRouteTablesInput, ...request.Option) (*ec2.DescribeRouteTablesOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeRouteTables(request *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeRouteTables: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun not implemented")
	}

	if len(request.RouteTableIds) != 0 {
		request.Filters = append(request.Filters, &ec2.Filter{Name: s("route-table-id"), Values: request.RouteTableIds})
	}

	response := &ec2.DescribeRouteTablesOutput{}
	for _, rt := range m.RouteTables {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "route-table-id":
				for _, v := range filter.Values {
					if *rt.RouteTableId == *v {
						match = true
					}
				}
			case "association.subnet-id":
				for _, a := range rt.Associations {
					for _, v := range filter.Values {
						if aws.StringValue(a.SubnetId) == *v {
							match = true
						}
					}
				}
			default:
				match = m.hasTag(ec2.ResourceTypeRouteTable, *rt.RouteTableId, filter)
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
		copy.Tags = m.getTags(ec2.ResourceTypeRouteTable, *rt.RouteTableId)
		response.RouteTables = append(response.RouteTables, &copy)
	}

	return response, nil
}

func (m *MockEC2) CreateRouteTable(request *ec2.CreateRouteTableInput) (*ec2.CreateRouteTableOutput, error) {
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

	rt := &ec2.RouteTable{
		RouteTableId: s(id),
		VpcId:        request.VpcId,
	}

	if m.RouteTables == nil {
		m.RouteTables = make(map[string]*ec2.RouteTable)
	}
	m.RouteTables[id] = rt

	copy := *rt
	response := &ec2.CreateRouteTableOutput{
		RouteTable: &copy,
	}
	return response, nil
}

func (m *MockEC2) CreateRouteTableWithContext(aws.Context, *ec2.CreateRouteTableInput, ...request.Option) (*ec2.CreateRouteTableOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) CreateRouteTableRequest(*ec2.CreateRouteTableInput) (*request.Request, *ec2.CreateRouteTableOutput) {
	panic("Not implemented")
}

func (m *MockEC2) CreateRoute(request *ec2.CreateRouteInput) (*ec2.CreateRouteOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateRoute: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	rt := m.RouteTables[aws.StringValue(request.RouteTableId)]
	if rt == nil {
		return nil, fmt.Errorf("RouteTable not found")
	}

	r := &ec2.Route{
		DestinationCidrBlock:        request.DestinationCidrBlock,
		DestinationIpv6CidrBlock:    request.DestinationIpv6CidrBlock,
		EgressOnlyInternetGatewayId: request.EgressOnlyInternetGatewayId,
		GatewayId:                   request.GatewayId,
		InstanceId:                  request.InstanceId,
		NatGatewayId:                request.NatGatewayId,
		NetworkInterfaceId:          request.NetworkInterfaceId,
		VpcPeeringConnectionId:      request.VpcPeeringConnectionId,
	}

	rt.Routes = append(rt.Routes, r)
	return &ec2.CreateRouteOutput{Return: aws.Bool(true)}, nil
}
func (m *MockEC2) CreateRouteWithContext(aws.Context, *ec2.CreateRouteInput, ...request.Option) (*ec2.CreateRouteOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) CreateRouteRequest(*ec2.CreateRouteInput) (*request.Request, *ec2.CreateRouteOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteRouteTable(request *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteRouteTable: %v", request)

	id := aws.StringValue(request.RouteTableId)
	o := m.RouteTables[id]
	if o == nil {
		return nil, fmt.Errorf("RouteTable %q not found", id)
	}
	delete(m.RouteTables, id)

	return &ec2.DeleteRouteTableOutput{}, nil
}
func (m *MockEC2) DeleteRouteTableWithContext(aws.Context, *ec2.DeleteRouteTableInput, ...request.Option) (*ec2.DeleteRouteTableOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DeleteRouteTableRequest(*ec2.DeleteRouteTableInput) (*request.Request, *ec2.DeleteRouteTableOutput) {
	panic("Not implemented")
}
