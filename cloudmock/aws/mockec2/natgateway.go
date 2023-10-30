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
	"k8s.io/klog/v2"
)

func (m *MockEC2) CreateNatGatewayWithId(request *ec2.CreateNatGatewayInput, id string) (*ec2.CreateNatGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tags := tagSpecificationsToTags(request.TagSpecifications, ec2.ResourceTypeNatgateway)

	ngw := &ec2.NatGateway{
		NatGatewayId: s(id),
		SubnetId:     request.SubnetId,
		Tags:         tags,
	}

	// Immediately mark it ready
	ngw.State = aws.String("available")

	if request.AllocationId != nil {
		var eip *ec2.Address
		for _, address := range m.Addresses {
			if aws.StringValue(address.AllocationId) == *request.AllocationId {
				eip = address
			}
		}
		if eip == nil {
			return nil, fmt.Errorf("AllocationId %q not found", *request.AllocationId)
		}
		ngw.NatGatewayAddresses = append(ngw.NatGatewayAddresses, &ec2.NatGatewayAddress{
			AllocationId: eip.AllocationId,
			PrivateIp:    eip.PrivateIpAddress,
			PublicIp:     eip.PublicIp,
		})
	}

	if m.NatGateways == nil {
		m.NatGateways = make(map[string]*ec2.NatGateway)
	}
	m.NatGateways[*ngw.NatGatewayId] = ngw

	m.addTags(id, tags...)

	copy := *ngw

	return &ec2.CreateNatGatewayOutput{
		NatGateway:  &copy,
		ClientToken: request.ClientToken,
	}, nil
}

func (m *MockEC2) CreateNatGateway(request *ec2.CreateNatGatewayInput) (*ec2.CreateNatGatewayOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) WaitUntilNatGatewayAvailable(request *ec2.DescribeNatGatewaysInput) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("WaitUntilNatGatewayAvailable: %v", request)

	if len(request.NatGatewayIds) != 1 {
		return fmt.Errorf("we only support WaitUntilNatGatewayAvailable with one NatGatewayId")
	}

	ngw := m.NatGateways[*request.NatGatewayIds[0]]
	if ngw == nil {
		return fmt.Errorf("NatGateway not found")
	}

	// We just immediately mark it ready
	ngw.State = aws.String("available")

	return nil
}

func (m *MockEC2) WaitUntilNatGatewayAvailableWithContext(aws.Context, *ec2.DescribeNatGatewaysInput, ...request.WaiterOption) error {
	panic("Not implemented")
}

func (m *MockEC2) CreateNatGatewayWithContext(ctx aws.Context, request *ec2.CreateNatGatewayInput, options ...request.Option) (*ec2.CreateNatGatewayOutput, error) {
	klog.Infof("CreateNatGateway: %v", request)

	id := m.allocateId("nat")
	return m.CreateNatGatewayWithId(request, id)
}

func (m *MockEC2) CreateNatGatewayRequest(*ec2.CreateNatGatewayInput) (*request.Request, *ec2.CreateNatGatewayOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeNatGateways(*ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeNatGatewaysWithContext(ctx aws.Context, request *ec2.DescribeNatGatewaysInput, options ...request.Option) (*ec2.DescribeNatGatewaysOutput, error) {
	var pages []*ec2.DescribeNatGatewaysOutput
	callback := func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
		pages = append(pages, page)
		return true
	}

	if err := m.DescribeNatGatewaysPagesWithContext(ctx, request, callback, options...); err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("DescribeNatGatewaysPagesWithContext did not return any pages")
	}
	return pages[0], nil
}

func (m *MockEC2) DescribeNatGatewaysRequest(*ec2.DescribeNatGatewaysInput) (*request.Request, *ec2.DescribeNatGatewaysOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeNatGatewaysPages(*ec2.DescribeNatGatewaysInput, func(*ec2.DescribeNatGatewaysOutput, bool) bool) error {
	panic("Not implemented")
}

func (m *MockEC2) DescribeNatGatewaysPagesWithContext(ctx aws.Context, request *ec2.DescribeNatGatewaysInput, callback func(*ec2.DescribeNatGatewaysOutput, bool) bool, options ...request.Option) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeNatGatewaysPagesWithContext: %v", request)

	var ngws []*ec2.NatGateway

	if len(request.NatGatewayIds) != 0 {
		request.Filter = append(request.Filter, &ec2.Filter{Name: s("nat-gateway-id"), Values: request.NatGatewayIds})
	}

	for id, ngw := range m.NatGateways {
		allFiltersMatch := true
		for _, filter := range request.Filter {
			match := false
			switch *filter.Name {
			case "nat-gateway-id":
				for _, v := range filter.Values {
					if *ngw.NatGatewayId == *v {
						match = true
					}
				}
			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeNatgateway, *ngw.NatGatewayId, filter)
				} else {
					return fmt.Errorf("unknown filter name: %q", *filter.Name)
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

		copy := *ngw
		copy.Tags = m.getTags(ec2.ResourceTypeNatgateway, id)
		ngws = append(ngws, &copy)
	}

	response := &ec2.DescribeNatGatewaysOutput{
		NatGateways: ngws,
	}

	callback(response, false)
	return nil
}

func (m *MockEC2) DeleteNatGateway(request *ec2.DeleteNatGatewayInput) (*ec2.DeleteNatGatewayOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteNatGatewayWithContext(ctx aws.Context, request *ec2.DeleteNatGatewayInput, options ...request.Option) (*ec2.DeleteNatGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteNatGateway: %v", request)

	id := aws.StringValue(request.NatGatewayId)
	o := m.NatGateways[id]
	if o == nil {
		return nil, fmt.Errorf("NatGateway %q not found", id)
	}
	delete(m.NatGateways, id)

	return &ec2.DeleteNatGatewayOutput{}, nil
}

func (m *MockEC2) DeleteNatGatewayRequest(*ec2.DeleteNatGatewayInput) (*request.Request, *ec2.DeleteNatGatewayOutput) {
	panic("Not implemented")
}
