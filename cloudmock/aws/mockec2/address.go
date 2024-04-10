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
	"encoding/binary"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

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

	tags := tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeElasticIp)
	address := &ec2types.Address{
		AllocationId: s(id),
		Domain:       ec2types.DomainTypeVpc,
		PublicIp:     s(publicIP.String()),
		Tags:         tags,
	}
	if request.Address != nil {
		address.PublicIp = request.Address
	}

	if m.Addresses == nil {
		m.Addresses = make(map[string]*ec2types.Address)
	}
	m.Addresses[id] = address
	m.addTags(id, tags...)

	response := &ec2.AllocateAddressOutput{
		AllocationId: address.AllocationId,
		Domain:       address.Domain,
		PublicIp:     address.PublicIp,
	}
	return response, nil
}

func (m *MockEC2) AllocateAddress(ctx context.Context, request *ec2.AllocateAddressInput, optFns ...func(*ec2.Options)) (*ec2.AllocateAddressOutput, error) {
	klog.Infof("AllocateAddress: %v", request)
	id := m.allocateId("eipalloc")
	return m.AllocateAddressWithId(request, id)
}

func (m *MockEC2) DescribeAddresses(ctx context.Context, request *ec2.DescribeAddressesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeAddresses: %v", request)

	var addresses []ec2types.Address

	if len(request.AllocationIds) != 0 {
		request.Filters = append(request.Filters, ec2types.Filter{Name: s("allocation-id"), Values: request.AllocationIds})
	}
	for _, address := range m.Addresses {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			case "allocation-id":
				for _, v := range filter.Values {
					if *address.AllocationId == v {
						match = true
					}
				}

			case "public-ip":
				for _, v := range filter.Values {
					if *address.PublicIp == v {
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
		copy.Tags = m.getTags(ec2types.ResourceTypeElasticIp, *address.AllocationId)
		addresses = append(addresses, copy)
	}

	response := &ec2.DescribeAddressesOutput{
		Addresses: addresses,
	}

	return response, nil
}

func (m *MockEC2) ReleaseAddress(ctx context.Context, request *ec2.ReleaseAddressInput, optFns ...func(*ec2.Options)) (*ec2.ReleaseAddressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ReleaseAddress: %v", request)

	id := aws.ToString(request.AllocationId)
	o := m.Addresses[id]
	if o == nil {
		return nil, fmt.Errorf("Address %q not found", id)
	}
	delete(m.Addresses, id)

	return &ec2.ReleaseAddressOutput{}, nil
}
