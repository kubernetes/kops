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
	"encoding/binary"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func (m *MockEC2) AllocateAddressRequest(*ec2.AllocateAddressInput) (*request.Request, *ec2.AllocateAddressOutput) {
	panic("Not implemented")
}

func (m *MockEC2) AllocateAddressWithContext(aws.Context, *ec2.AllocateAddressInput, ...request.Option) (*ec2.AllocateAddressOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) AllocateAddressWithId(request *ec2.AllocateAddressInput, id string) (*ec2.AllocateAddressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.addressNumber++
	n := m.addressNumber

	publicIP := net.ParseIP("192.0.2.0").To4()
	{
		v := binary.BigEndian.Uint32(publicIP)
		v += uint32(n)
		publicIP = make(net.IP, len(publicIP))
		binary.BigEndian.PutUint32(publicIP, v)
	}

	address := &ec2.Address{
		AllocationId: s(id),
		Domain:       s("vpc"),
		PublicIp:     s(publicIP.String()),
	}
	if m.Addresses == nil {
		m.Addresses = make(map[string]*ec2.Address)
	}
	m.Addresses[id] = address
	response := &ec2.AllocateAddressOutput{
		AllocationId: address.AllocationId,
		Domain:       address.Domain,
		PublicIp:     address.PublicIp,
	}
	return response, nil
}

func (m *MockEC2) AllocateAddress(request *ec2.AllocateAddressInput) (*ec2.AllocateAddressOutput, error) {
	klog.Infof("AllocateAddress: %v", request)
	id := m.allocateId("eipalloc")
	return m.AllocateAddressWithId(request, id)
}

func (m *MockEC2) AssignPrivateIpAddressesRequest(*ec2.AssignPrivateIpAddressesInput) (*request.Request, *ec2.AssignPrivateIpAddressesOutput) {
	panic("Not implemented")
}

func (m *MockEC2) AssignPrivateIpAddressesWithContext(aws.Context, *ec2.AssignPrivateIpAddressesInput, ...request.Option) (*ec2.AssignPrivateIpAddressesOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) AssignPrivateIpAddresses(*ec2.AssignPrivateIpAddressesInput) (*ec2.AssignPrivateIpAddressesOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) AssociateAddressRequest(*ec2.AssociateAddressInput) (*request.Request, *ec2.AssociateAddressOutput) {
	panic("Not implemented")
}

func (m *MockEC2) AssociateAddressWithContext(aws.Context, *ec2.AssociateAddressInput, ...request.Option) (*ec2.AssociateAddressOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) AssociateAddress(*ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeAddressesRequest(*ec2.DescribeAddressesInput) (*request.Request, *ec2.DescribeAddressesOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeAddressesWithContext(aws.Context, *ec2.DescribeAddressesInput, ...request.Option) (*ec2.DescribeAddressesOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeAddresses(request *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeAddresses: %v", request)

	var addresses []*ec2.Address

	if len(request.AllocationIds) != 0 {
		request.Filters = append(request.Filters, &ec2.Filter{Name: s("allocation-id"), Values: request.AllocationIds})
	}
	for _, address := range m.Addresses {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			case "allocation-id":
				for _, v := range filter.Values {
					if *address.AllocationId == *v {
						match = true
					}
				}

			case "public-ip":
				for _, v := range filter.Values {
					if *address.PublicIp == *v {
						match = true
					}
				}

			default:
				return nil, fmt.Errorf("unknown filter name: %q", *filter.Name)
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := *address
		addresses = append(addresses, &copy)
	}

	response := &ec2.DescribeAddressesOutput{
		Addresses: addresses,
	}

	return response, nil
}

func (m *MockEC2) ReleaseAddressRequest(*ec2.ReleaseAddressInput) (*request.Request, *ec2.ReleaseAddressOutput) {
	panic("Not implemented")
}
func (m *MockEC2) ReleaseAddressWithContext(aws.Context, *ec2.ReleaseAddressInput, ...request.Option) (*ec2.ReleaseAddressOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) ReleaseAddress(request *ec2.ReleaseAddressInput) (*ec2.ReleaseAddressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ReleaseAddress: %v", request)

	id := aws.StringValue(request.AllocationId)
	o := m.Addresses[id]
	if o == nil {
		return nil, fmt.Errorf("Address %q not found", id)
	}
	delete(m.Addresses, id)

	return &ec2.ReleaseAddressOutput{}, nil
}
