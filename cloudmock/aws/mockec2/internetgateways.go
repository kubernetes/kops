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
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func (m *MockEC2) FindInternetGateway(id string) *ec2.InternetGateway {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	internetGateway := m.InternetGateways[id]
	if internetGateway == nil {
		return nil
	}

	copy := *internetGateway
	copy.Tags = m.getTags(ec2.ResourceTypeInternetGateway, id)
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

func (m *MockEC2) CreateInternetGatewayRequest(*ec2.CreateInternetGatewayInput) (*request.Request, *ec2.CreateInternetGatewayOutput) {
	panic("Not implemented")
}

func (m *MockEC2) CreateInternetGatewayWithContext(aws.Context, *ec2.CreateInternetGatewayInput, ...request.Option) (*ec2.CreateInternetGatewayOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) CreateInternetGateway(request *ec2.CreateInternetGatewayInput) (*ec2.CreateInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateInternetGateway: %v", request)

	id := m.allocateId("igw")

	igw := &ec2.InternetGateway{
		InternetGatewayId: s(id),
	}

	if m.InternetGateways == nil {
		m.InternetGateways = make(map[string]*ec2.InternetGateway)
	}
	m.InternetGateways[id] = igw

	response := &ec2.CreateInternetGatewayOutput{
		InternetGateway: igw,
	}
	return response, nil
}

func (m *MockEC2) DescribeInternetGatewaysRequest(*ec2.DescribeInternetGatewaysInput) (*request.Request, *ec2.DescribeInternetGatewaysOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeInternetGatewaysWithContext(aws.Context, *ec2.DescribeInternetGatewaysInput, ...request.Option) (*ec2.DescribeInternetGatewaysOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeInternetGateways(request *ec2.DescribeInternetGatewaysInput) (*ec2.DescribeInternetGatewaysOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeInternetGateways: %v", request)

	var internetGateways []*ec2.InternetGateway

	if len(request.InternetGatewayIds) != 0 {
		request.Filters = append(request.Filters, &ec2.Filter{Name: s("internet-gateway-id"), Values: request.InternetGatewayIds})
	}

	for id, internetGateway := range m.InternetGateways {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "internet-gateway-id":
				for _, v := range filter.Values {
					if id == aws.StringValue(v) {
						match = true
					}
				}

			case "attachment.vpc-id":
				for _, v := range filter.Values {
					if internetGateway.Attachments != nil {
						for _, attachment := range internetGateway.Attachments {
							if *attachment.VpcId == *v {
								match = true
							}
						}
					}
				}

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeInternetGateway, id, filter)
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
		copy.Tags = m.getTags(ec2.ResourceTypeInternetGateway, id)
		internetGateways = append(internetGateways, &copy)
	}

	response := &ec2.DescribeInternetGatewaysOutput{
		InternetGateways: internetGateways,
	}

	return response, nil
}

func (m *MockEC2) AttachInternetGateway(request *ec2.AttachInternetGatewayInput) (*ec2.AttachInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, internetGateway := range m.InternetGateways {
		if id == *request.InternetGatewayId {
			internetGateway.Attachments = append(internetGateway.Attachments,
				&ec2.InternetGatewayAttachment{
					VpcId: request.VpcId,
				})
			return &ec2.AttachInternetGatewayOutput{}, nil
		}
	}

	return nil, fmt.Errorf("InternetGateway not found")

}

func (m *MockEC2) AttachInternetGatewayWithContext(aws.Context, *ec2.AttachInternetGatewayInput, ...request.Option) (*ec2.AttachInternetGatewayOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) AttachInternetGatewayRequest(*ec2.AttachInternetGatewayInput) (*request.Request, *ec2.AttachInternetGatewayOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DetachInternetGateway(request *ec2.DetachInternetGatewayInput) (*ec2.DetachInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, igw := range m.InternetGateways {
		if id == *request.InternetGatewayId {
			found := false
			var newAttachments []*ec2.InternetGatewayAttachment
			for _, a := range igw.Attachments {
				if aws.StringValue(a.VpcId) == aws.StringValue(request.VpcId) {
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

func (m *MockEC2) DetachInternetGatewayWithContext(aws.Context, *ec2.DetachInternetGatewayInput, ...request.Option) (*ec2.DetachInternetGatewayOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DetachInternetGatewayRequest(*ec2.DetachInternetGatewayInput) (*request.Request, *ec2.DetachInternetGatewayOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteInternetGateway(request *ec2.DeleteInternetGatewayInput) (*ec2.DeleteInternetGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteInternetGateway: %v", request)

	id := aws.StringValue(request.InternetGatewayId)
	o := m.InternetGateways[id]
	if o == nil {
		return nil, fmt.Errorf("InternetGateway %q not found", id)
	}
	delete(m.InternetGateways, id)

	return &ec2.DeleteInternetGatewayOutput{}, nil
}

func (m *MockEC2) DeleteInternetGatewayWithContext(aws.Context, *ec2.DeleteInternetGatewayInput, ...request.Option) (*ec2.DeleteInternetGatewayOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DeleteInternetGatewayRequest(*ec2.DeleteInternetGatewayInput) (*request.Request, *ec2.DeleteInternetGatewayOutput) {
	panic("Not implemented")
}
