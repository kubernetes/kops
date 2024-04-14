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

func (m *MockEC2) FindInternetGateway(id string) *ec2types.InternetGateway {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	internetGateway := m.InternetGateways[id]
	if internetGateway == nil {
		return nil
	}

	copy := *internetGateway
	copy.Tags = m.getTags(ec2types.ResourceTypeInternetGateway, id)
	return &copy
}

func (m *MockEC2) InternetGatewayIds() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var ids []string
	for id := range m.InternetGateways {
		ids = append(ids, id)
	}
	return ids
}

func (m *MockEC2) CreateInternetGateway(ctx context.Context, request *ec2.CreateInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.CreateInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateInternetGateway: %v", request)

	id := m.allocateId("igw")
	tags := tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeInternetGateway)

	igw := &ec2types.InternetGateway{
		InternetGatewayId: s(id),
		Tags:              tags,
	}

	if m.InternetGateways == nil {
		m.InternetGateways = make(map[string]*ec2types.InternetGateway)
	}
	m.InternetGateways[id] = igw

	m.addTags(id, tags...)

	response := &ec2.CreateInternetGatewayOutput{
		InternetGateway: igw,
	}
	return response, nil
}

func (m *MockEC2) DescribeInternetGateways(ctx context.Context, request *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeInternetGateways: %v", request)

	var internetGateways []ec2types.InternetGateway

	if len(request.InternetGatewayIds) != 0 {
		request.Filters = append(request.Filters, ec2types.Filter{Name: s("internet-gateway-id"), Values: request.InternetGatewayIds})
	}

	for id, internetGateway := range m.InternetGateways {
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
					match = m.hasTag(ec2types.ResourceTypeInternetGateway, id, filter)
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
		copy.Tags = m.getTags(ec2types.ResourceTypeInternetGateway, id)
		internetGateways = append(internetGateways, copy)
	}

	response := &ec2.DescribeInternetGatewaysOutput{
		InternetGateways: internetGateways,
	}

	return response, nil
}

func (m *MockEC2) AttachInternetGateway(ctx context.Context, request *ec2.AttachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.AttachInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, internetGateway := range m.InternetGateways {
		if id == *request.InternetGatewayId {
			internetGateway.Attachments = append(internetGateway.Attachments,
				ec2types.InternetGatewayAttachment{
					VpcId: request.VpcId,
				})
			return &ec2.AttachInternetGatewayOutput{}, nil
		}
	}

	return nil, fmt.Errorf("InternetGateway not found")
}

func (m *MockEC2) DetachInternetGateway(ctx context.Context, request *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, igw := range m.InternetGateways {
		if id == *request.InternetGatewayId {
			found := false
			var newAttachments []ec2types.InternetGatewayAttachment
			for _, a := range igw.Attachments {
				if a.VpcId == request.VpcId {
					found = true
					continue
				}
				newAttachments = append(newAttachments, a)
			}

			if !found {
				return nil, fmt.Errorf("Attachment to VPC not found")
			}
			igw.Attachments = newAttachments

			return &ec2.DetachInternetGatewayOutput{}, nil
		}
	}

	return nil, fmt.Errorf("InternetGateway not found")
}

func (m *MockEC2) DeleteInternetGateway(ctx context.Context, request *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteInternetGateway: %v", request)

	id := aws.ToString(request.InternetGatewayId)
	o := m.InternetGateways[id]
	if o == nil {
		return nil, fmt.Errorf("InternetGateway %q not found", id)
	}
	delete(m.InternetGateways, id)

	return &ec2.DeleteInternetGatewayOutput{}, nil
}
