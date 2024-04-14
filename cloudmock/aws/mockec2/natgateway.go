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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

func (m *MockEC2) CreateNatGatewayWithId(request *ec2.CreateNatGatewayInput, id string) (*ec2.CreateNatGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tags := tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeNatgateway)

	ngw := &ec2types.NatGateway{
		NatGatewayId: s(id),
		SubnetId:     request.SubnetId,
		Tags:         tags,
	}

	if request.AllocationId != nil {
		var eip *ec2types.Address
		for _, address := range m.Addresses {
			if aws.ToString(address.AllocationId) == *request.AllocationId {
				eip = address
			}
		}
		if eip == nil {
			return nil, fmt.Errorf("AllocationId %q not found", *request.AllocationId)
		}
		ngw.NatGatewayAddresses = append(ngw.NatGatewayAddresses, ec2types.NatGatewayAddress{
			AllocationId: eip.AllocationId,
			PrivateIp:    eip.PrivateIpAddress,
			PublicIp:     eip.PublicIp,
		})
	}

	if m.NatGateways == nil {
		m.NatGateways = make(map[string]*ec2types.NatGateway)
	}
	m.NatGateways[*ngw.NatGatewayId] = ngw

	m.addTags(id, tags...)

	copy := *ngw

	return &ec2.CreateNatGatewayOutput{
		NatGateway:  &copy,
		ClientToken: request.ClientToken,
	}, nil
}

func (m *MockEC2) CreateNatGateway(ctx context.Context, request *ec2.CreateNatGatewayInput, optFns ...func(*ec2.Options)) (*ec2.CreateNatGatewayOutput, error) {
	klog.Infof("CreateNatGateway: %v", request)

	id := m.allocateId("nat")
	return m.CreateNatGatewayWithId(request, id)
}

/*func (m *MockEC2) WaitUntilNatGatewayAvailable(request *ec2.DescribeNatGatewaysInput) error {
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
	ngw.State = aws.String("Available")

	return nil
}*/

func (m *MockEC2) DescribeNatGateways(ctx context.Context, request *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeNatGateways: %v", request)

	var ngws []ec2types.NatGateway

	if len(request.NatGatewayIds) != 0 {
		request.Filter = append(request.Filter, ec2types.Filter{Name: s("nat-gateway-id"), Values: request.NatGatewayIds})
	}

	for id, ngw := range m.NatGateways {
		allFiltersMatch := true
		for _, filter := range request.Filter {
			match := false
			switch *filter.Name {
			case "nat-gateway-id":
				for _, v := range filter.Values {
					if *ngw.NatGatewayId == v {
						match = true
					}
				}
			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2types.ResourceTypeNatgateway, *ngw.NatGatewayId, filter)
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

		copy := *ngw
		copy.Tags = m.getTags(ec2types.ResourceTypeNatgateway, id)
		ngws = append(ngws, copy)
	}

	response := &ec2.DescribeNatGatewaysOutput{
		NatGateways: ngws,
	}

	return response, nil
}

func (m *MockEC2) DeleteNatGateway(ctx context.Context, request *ec2.DeleteNatGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNatGatewayOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteNatGateway: %v", request)

	id := aws.ToString(request.NatGatewayId)
	o := m.NatGateways[id]
	if o == nil {
		return nil, fmt.Errorf("NatGateway %q not found", id)
	}
	delete(m.NatGateways, id)

	return &ec2.DeleteNatGatewayOutput{}, nil
}
