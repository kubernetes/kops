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

func (m *MockEC2) DescribeDhcpOptions(ctx context.Context, request *ec2.DescribeDhcpOptionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeDhcpOptionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeDhcpOptions: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun not implemented")
	}
	if request.DhcpOptionsIds != nil {
		klog.Fatalf("DhcpOptionsIds not implemented")
	}

	response := &ec2.DescribeDhcpOptionsOutput{}

	for id, dhcpOptions := range m.DhcpOptions {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			var match bool
			if strings.HasPrefix(*filter.Name, "tag:") {
				match = m.hasTag(ec2types.ResourceTypeDhcpOptions, *dhcpOptions.DhcpOptionsId, filter)
			} else {
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

		copy := *dhcpOptions
		copy.Tags = m.getTags(ec2types.ResourceTypeDhcpOptions, id)
		response.DhcpOptions = append(response.DhcpOptions, copy)
	}

	return response, nil
}

func (m *MockEC2) AssociateDhcpOptions(ctx context.Context, request *ec2.AssociateDhcpOptionsInput, optFns ...func(*ec2.Options)) (*ec2.AssociateDhcpOptionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AssociateDhcpOptions: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	dopt := m.DhcpOptions[*request.DhcpOptionsId]
	if dopt == nil {
		return nil, fmt.Errorf("DhcpOptions not found")
	}
	vpc := m.Vpcs[*request.VpcId]
	if vpc == nil {
		return nil, fmt.Errorf("vpc not found")
	}

	vpc.main.DhcpOptionsId = dopt.DhcpOptionsId

	response := &ec2.AssociateDhcpOptionsOutput{}

	return response, nil
}

func (m *MockEC2) CreateDhcpOptions(ctx context.Context, request *ec2.CreateDhcpOptionsInput, optFns ...func(*ec2.Options)) (*ec2.CreateDhcpOptionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateDhcpOptions: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	n := len(m.DhcpOptions) + 1
	id := fmt.Sprintf("dopt-%d", n)

	dhcpOptions := &ec2types.DhcpOptions{
		DhcpOptionsId: s(id),
	}

	for _, o := range request.DhcpConfigurations {
		c := ec2types.DhcpConfiguration{
			Key: o.Key,
		}
		for _, v := range o.Values {
			c.Values = append(c.Values, ec2types.AttributeValue{
				Value: aws.String(v),
			})
		}
		dhcpOptions.DhcpConfigurations = append(dhcpOptions.DhcpConfigurations, c)
	}
	if m.DhcpOptions == nil {
		m.DhcpOptions = make(map[string]*ec2types.DhcpOptions)
	}
	m.DhcpOptions[*dhcpOptions.DhcpOptionsId] = dhcpOptions

	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeDhcpOptions)...)

	copy := *dhcpOptions
	copy.Tags = m.getTags(ec2types.ResourceTypeDhcpOptions, *dhcpOptions.DhcpOptionsId)
	return &ec2.CreateDhcpOptionsOutput{DhcpOptions: &copy}, nil
}

func (m *MockEC2) DeleteDhcpOptions(ctx context.Context, request *ec2.DeleteDhcpOptionsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteDhcpOptionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteDhcpOptions: %v", request)

	id := aws.ToString(request.DhcpOptionsId)
	o := m.DhcpOptions[id]
	if o == nil {
		return nil, fmt.Errorf("DhcpOptions %q not found", id)
	}
	delete(m.DhcpOptions, id)

	return &ec2.DeleteDhcpOptionsOutput{}, nil
}
