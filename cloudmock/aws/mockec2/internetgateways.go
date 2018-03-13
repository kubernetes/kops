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
	"github.com/golang/glog"
)

type internetGatewayInfo struct {
	main ec2.InternetGateway
}

func (m *MockEC2) FindInternetGateway(id string) *ec2.InternetGateway {
	internetGateway := m.InternetGateways[id]
	if internetGateway == nil {
		return nil
	}

	copy := internetGateway.main
	copy.Tags = m.getTags(ec2.ResourceTypeInternetGateway, id)
	return &copy
}

func (m *MockEC2) InternetGatewayIds() []string {
	var ids []string
	for id := range m.InternetGateways {
		ids = append(ids, id)
	}
	return ids
}

func (m *MockEC2) CreateInternetGatewayRequest(*ec2.CreateInternetGatewayInput) (*request.Request, *ec2.CreateInternetGatewayOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) CreateInternetGatewayWithContext(aws.Context, *ec2.CreateInternetGatewayInput, ...request.Option) (*ec2.CreateInternetGatewayOutput, error) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) CreateInternetGateway(request *ec2.CreateInternetGatewayInput) (*ec2.CreateInternetGatewayOutput, error) {
	glog.Infof("CreateInternetGateway: %v", request)

	m.internetGatewayNumber++
	n := m.internetGatewayNumber

	internetGateway := &ec2.InternetGateway{
		InternetGatewayId: s(fmt.Sprintf("igw-%d", n)),
	}

	if m.InternetGateways == nil {
		m.InternetGateways = make(map[string]*internetGatewayInfo)
	}
	m.InternetGateways[*internetGateway.InternetGatewayId] = &internetGatewayInfo{
		main: *internetGateway,
	}

	response := &ec2.CreateInternetGatewayOutput{
		InternetGateway: internetGateway,
	}
	return response, nil
}

func (m *MockEC2) DescribeInternetGatewaysRequest(*ec2.DescribeInternetGatewaysInput) (*request.Request, *ec2.DescribeInternetGatewaysOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) DescribeInternetGatewaysWithContext(aws.Context, *ec2.DescribeInternetGatewaysInput, ...request.Option) (*ec2.DescribeInternetGatewaysOutput, error) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) DescribeInternetGateways(request *ec2.DescribeInternetGatewaysInput) (*ec2.DescribeInternetGatewaysOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	glog.Infof("DescribeInternetGateways: %v", request)

	var internetGateways []*ec2.InternetGateway

	for id, internetGateway := range m.InternetGateways {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			case "attachment.vpc-id":
				for _, v := range filter.Values {
					if internetGateway.main.Attachments != nil {
						for _, attachment := range internetGateway.main.Attachments {
							if *attachment.VpcId == *v {
								match = true
							}
						}
					}
				}

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeInternetGateway, *internetGateway.main.InternetGatewayId, filter)
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

		copy := internetGateway.main
		copy.Tags = m.getTags(ec2.ResourceTypeInternetGateway, id)
		internetGateways = append(internetGateways, &copy)
	}

	response := &ec2.DescribeInternetGatewaysOutput{
		InternetGateways: internetGateways,
	}

	return response, nil
}

func (m *MockEC2) AttachInternetGateway(request *ec2.AttachInternetGatewayInput) (*ec2.AttachInternetGatewayOutput, error) {
	for id, internetGateway := range m.InternetGateways {
		if id == *request.InternetGatewayId {
			internetGateway.main.Attachments = append(internetGateway.main.Attachments,
				&ec2.InternetGatewayAttachment{
					VpcId: request.VpcId,
				})
		}
	}
	return &ec2.AttachInternetGatewayOutput{}, nil
}

func (m *MockEC2) AttachInternetGatewayWithContext(aws.Context, *ec2.AttachInternetGatewayInput, ...request.Option) (*ec2.AttachInternetGatewayOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) AttachInternetGatewayRequest(*ec2.AttachInternetGatewayInput) (*request.Request, *ec2.AttachInternetGatewayOutput) {
	panic("Not implemented")
	return nil, nil
}
