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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

type subnetInfo struct {
	main ec2.Subnet
}

func (m *MockEC2) FindSubnet(id string) *ec2.Subnet {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	subnet := m.subnets[id]
	if subnet == nil {
		return nil
	}

	copy := subnet.main
	copy.Tags = m.getTags(ec2.ResourceTypeSubnet, id)
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

func (m *MockEC2) CreateSubnetRequest(*ec2.CreateSubnetInput) (*request.Request, *ec2.CreateSubnetOutput) {
	panic("Not implemented")
}

func (m *MockEC2) CreateSubnetWithContext(aws.Context, *ec2.CreateSubnetInput, ...request.Option) (*ec2.CreateSubnetOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) CreateSubnetWithId(request *ec2.CreateSubnetInput, id string) (*ec2.CreateSubnetOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	subnet := &ec2.Subnet{
		SubnetId:         s(id),
		VpcId:            request.VpcId,
		CidrBlock:        request.CidrBlock,
		AvailabilityZone: request.AvailabilityZone,
	}

	if m.subnets == nil {
		m.subnets = make(map[string]*subnetInfo)
	}
	m.subnets[*subnet.SubnetId] = &subnetInfo{
		main: *subnet,
	}

	response := &ec2.CreateSubnetOutput{
		Subnet: subnet,
	}
	return response, nil
}

func (m *MockEC2) CreateSubnet(request *ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	klog.Infof("CreateSubnet: %v", request)

	id := m.allocateId("subnet")
	return m.CreateSubnetWithId(request, id)
}

func (m *MockEC2) DescribeSubnetsRequest(*ec2.DescribeSubnetsInput) (*request.Request, *ec2.DescribeSubnetsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeSubnetsWithContext(aws.Context, *ec2.DescribeSubnetsInput, ...request.Option) (*ec2.DescribeSubnetsOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeSubnets(request *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeSubnets: %v", request)

	if len(request.SubnetIds) != 0 {
		request.Filters = append(request.Filters, &ec2.Filter{Name: s("subnet-id"), Values: request.SubnetIds})
	}

	var subnets []*ec2.Subnet

	for id, subnet := range m.subnets {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "vpc-id":
				if *subnet.main.VpcId == *filter.Values[0] {
					match = true
				}
			case "subnet-id":
				for _, v := range filter.Values {
					if *subnet.main.SubnetId == *v {
						match = true
					}
				}
			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeSubnet, *subnet.main.SubnetId, filter)
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
		copy.Tags = m.getTags(ec2.ResourceTypeSubnet, id)
		subnets = append(subnets, &copy)
	}

	response := &ec2.DescribeSubnetsOutput{
		Subnets: subnets,
	}

	return response, nil
}

func (m *MockEC2) AssociateRouteTable(request *ec2.AssociateRouteTableInput) (*ec2.AssociateRouteTableOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AuthorizeSecurityGroupIngress: %v", request)

	if aws.StringValue(request.SubnetId) == "" {
		return nil, fmt.Errorf("SubnetId not specified")
	}
	if aws.StringValue(request.RouteTableId) == "" {
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

	rt.Associations = append(rt.Associations, &ec2.RouteTableAssociation{
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
func (m *MockEC2) AssociateRouteTableWithContext(aws.Context, *ec2.AssociateRouteTableInput, ...request.Option) (*ec2.AssociateRouteTableOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) AssociateRouteTableRequest(*ec2.AssociateRouteTableInput) (*request.Request, *ec2.AssociateRouteTableOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteSubnet(request *ec2.DeleteSubnetInput) (*ec2.DeleteSubnetOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteSubnet: %v", request)

	id := aws.StringValue(request.SubnetId)
	o := m.subnets[id]
	if o == nil {
		return nil, fmt.Errorf("Subnet %q not found", id)
	}
	delete(m.subnets, id)

	return &ec2.DeleteSubnetOutput{}, nil
}

func (m *MockEC2) DeleteSubnetWithContext(aws.Context, *ec2.DeleteSubnetInput, ...request.Option) (*ec2.DeleteSubnetOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DeleteSubnetRequest(*ec2.DeleteSubnetInput) (*request.Request, *ec2.DeleteSubnetOutput) {
	panic("Not implemented")
}
