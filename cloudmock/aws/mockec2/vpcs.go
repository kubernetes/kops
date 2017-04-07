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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"strings"
)

type vpcInfo struct {
	main       ec2.Vpc
	attributes ec2.DescribeVpcAttributeOutput
}

func (m *MockEC2) FindVpc(id string) *ec2.Vpc {
	vpc := m.Vpcs[id]
	if vpc == nil {
		return nil
	}

	copy := vpc.main
	copy.Tags = m.getTags(ec2.ResourceTypeVpc, *vpc.main.VpcId)

	return &copy
}

func (m *MockEC2) CreateVpcRequest(*ec2.CreateVpcInput) (*request.Request, *ec2.CreateVpcOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) CreateVpc(request *ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	glog.Infof("CreateVpc: %v", request)

	m.vpcNumber++
	n := m.vpcNumber

	id := fmt.Sprintf("vpc-%d", n)

	vpc := &vpcInfo{
		main: ec2.Vpc{
			VpcId:     s(id),
			CidrBlock: request.CidrBlock,
			IsDefault: aws.Bool(false),
		},
		attributes: ec2.DescribeVpcAttributeOutput{
			EnableDnsHostnames: &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
			EnableDnsSupport:   &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
		},
	}

	if m.Vpcs == nil {
		m.Vpcs = make(map[string]*vpcInfo)
	}
	m.Vpcs[id] = vpc

	response := &ec2.CreateVpcOutput{
		Vpc: &vpc.main,
	}
	return response, nil
}

func (m *MockEC2) DescribeVpcsRequest(*ec2.DescribeVpcsInput) (*request.Request, *ec2.DescribeVpcsOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) DescribeVpcs(request *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	glog.Infof("DescribeVpcs: %v", request)

	var vpcs []*ec2.Vpc

	for _, vpc := range m.Vpcs {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeVpc, *vpc.main.VpcId, filter)
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

		copy := vpc.main
		copy.Tags = m.getTags(ec2.ResourceTypeVpc, *vpc.main.VpcId)
		vpcs = append(vpcs, &copy)
	}

	response := &ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}

	return response, nil
}

func (m *MockEC2) DescribeVpcAttributeRequest(*ec2.DescribeVpcAttributeInput) (*request.Request, *ec2.DescribeVpcAttributeOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeVpcAttribute(request *ec2.DescribeVpcAttributeInput) (*ec2.DescribeVpcAttributeOutput, error) {
	glog.Infof("DescribeVpcs: %v", request)

	vpc := m.Vpcs[*request.VpcId]
	if vpc == nil {
		return nil, fmt.Errorf("not found")
	}

	response := &ec2.DescribeVpcAttributeOutput{
		VpcId: vpc.main.VpcId,

		EnableDnsHostnames: vpc.attributes.EnableDnsHostnames,
		EnableDnsSupport:   vpc.attributes.EnableDnsSupport,
	}

	return response, nil
}

func (m *MockEC2) DescribeInternetGatewaysRequest(*ec2.DescribeInternetGatewaysInput) (*request.Request, *ec2.DescribeInternetGatewaysOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeInternetGateways(*ec2.DescribeInternetGatewaysInput) (*ec2.DescribeInternetGatewaysOutput, error) {
	return &ec2.DescribeInternetGatewaysOutput{
		InternetGateways: []*ec2.InternetGateway{{
			InternetGatewayId: aws.String("fake-ig"),
		}},
	}, nil
}
