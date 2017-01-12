/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"strings"
)

func (m *MockEC2) CreateSubnetRequest(*ec2.CreateSubnetInput) (*request.Request, *ec2.CreateSubnetOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) CreateSubnet(request *ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	glog.Infof("CreateSubnet: %v", request)

	m.subnetNumber++
	n := m.subnetNumber

	subnet := &ec2.Subnet{
		SubnetId:  s(fmt.Sprintf("subnet-%d", n)),
		VpcId:     request.VpcId,
		CidrBlock: request.CidrBlock,
	}
	m.Subnets = append(m.Subnets, subnet)
	response := &ec2.CreateSubnetOutput{
		Subnet: subnet,
	}
	return response, nil
}

func (m *MockEC2) DescribeSubnetsRequest(*ec2.DescribeSubnetsInput) (*request.Request, *ec2.DescribeSubnetsOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) DescribeSubnets(request *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	glog.Infof("DescribeSubnets: %v", request)

	var subnets []*ec2.Subnet

	for _, subnet := range m.Subnets {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeSubnet, *subnet.SubnetId, filter)
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

		copy := *subnet
		copy.Tags = m.getTags(ec2.ResourceTypeSubnet, *subnet.SubnetId)
		subnets = append(subnets, &copy)
	}

	response := &ec2.DescribeSubnetsOutput{
		Subnets: subnets,
	}

	return response, nil
}
