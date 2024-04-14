/*
Copyright 2017 The Kubernetes Authors.

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

func (m *MockEC2) FindEgressOnlyInternetGateway(id string) *ec2types.EgressOnlyInternetGateway {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	internetGateway := m.EgressOnlyInternetGateways[id]
	if internetGateway == nil {
		return nil
	}

	copy := *internetGateway
	copy.Tags = m.getTags(ec2types.ResourceTypeEgressOnlyInternetGateway, id)
	return &copy
}

func (m *MockEC2) EgressOnlyInternetGatewayIds() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var ids []string
	for id := range m.EgressOnlyInternetGateways {
		ids = append(ids, id)
	}
	return ids
}

func (m *MockEC2) CreateEgressOnlyInternetGateway(ctx context.Context, request *ec2.CreateEgressOnlyInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.CreateEgressOnlyInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateEgressOnlyInternetGateway: %v", request)

	id := m.allocateId("eigw")
	tags := tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeEgressOnlyInternetGateway)

	eigw := &ec2types.EgressOnlyInternetGateway{
		EgressOnlyInternetGatewayId: s(id),
		Attachments: []ec2types.InternetGatewayAttachment{
			{
				VpcId: request.VpcId,
			},
		},
		Tags: tags,
	}

	if m.EgressOnlyInternetGateways == nil {
		m.EgressOnlyInternetGateways = make(map[string]*ec2types.EgressOnlyInternetGateway)
	}
	m.EgressOnlyInternetGateways[id] = eigw

	m.addTags(id, tags...)

	response := &ec2.CreateEgressOnlyInternetGatewayOutput{
		EgressOnlyInternetGateway: eigw,
	}
	return response, nil
}

func (m *MockEC2) DescribeEgressOnlyInternetGateways(ctx context.Context, request *ec2.DescribeEgressOnlyInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeEgressOnlyInternetGatewaysOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeEgressOnlyInternetGateways: %v", request)

	var internetGateways []ec2types.EgressOnlyInternetGateway

	if len(request.EgressOnlyInternetGatewayIds) != 0 {
		request.Filters = append(request.Filters, ec2types.Filter{Name: s("egress-only-internet-gateway-id"), Values: request.EgressOnlyInternetGatewayIds})
	}

	for id, internetGateway := range m.EgressOnlyInternetGateways {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "internet-gateway-id":
				for _, v := range filter.Values {
					if id == v {
						match = true
					}
				}

			case "attachment.vpc-id":
				for _, v := range filter.Values {
					if internetGateway.Attachments != nil {
						for _, attachment := range internetGateway.Attachments {
							if *attachment.VpcId == v {
								match = true
							}
						}
					}
				}

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2types.ResourceTypeEgressOnlyInternetGateway, id, filter)
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

		copy := *internetGateway
		copy.Tags = m.getTags(ec2types.ResourceTypeEgressOnlyInternetGateway, id)
		internetGateways = append(internetGateways, copy)
	}

	response := &ec2.DescribeEgressOnlyInternetGatewaysOutput{
		EgressOnlyInternetGateways: internetGateways,
	}

	return response, nil
}

func (m *MockEC2) DeleteEgressOnlyInternetGateway(ctx context.Context, request *ec2.DeleteEgressOnlyInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteEgressOnlyInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteEgressOnlyInternetGateway: %v", request)

	id := aws.ToString(request.EgressOnlyInternetGatewayId)
	o := m.EgressOnlyInternetGateways[id]
	if o == nil {
		return nil, fmt.Errorf("EgressOnlyInternetGateway %q not found", id)
	}
	delete(m.EgressOnlyInternetGateways, id)

	return &ec2.DeleteEgressOnlyInternetGatewayOutput{}, nil
}
