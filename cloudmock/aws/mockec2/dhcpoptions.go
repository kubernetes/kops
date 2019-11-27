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
	"k8s.io/klog"
)

func (m *MockEC2) DescribeDhcpOptions(request *ec2.DescribeDhcpOptionsInput) (*ec2.DescribeDhcpOptionsOutput, error) {
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
			match := false
			switch *filter.Name {
			// case "vpc-id":
			// 	if *subnet.main.VpcId == *filter.Values[0] {
			// 		match = true
			// 	}
			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeDhcpOptions, *dhcpOptions.DhcpOptionsId, filter)
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

		copy := *dhcpOptions
		copy.Tags = m.getTags(ec2.ResourceTypeDhcpOptions, id)
		response.DhcpOptions = append(response.DhcpOptions, &copy)
	}

	return response, nil
}

func (m *MockEC2) DescribeDhcpOptionsWithContext(aws.Context, *ec2.DescribeDhcpOptionsInput, ...request.Option) (*ec2.DescribeDhcpOptionsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeDhcpOptionsRequest(*ec2.DescribeDhcpOptionsInput) (*request.Request, *ec2.DescribeDhcpOptionsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) AssociateDhcpOptions(request *ec2.AssociateDhcpOptionsInput) (*ec2.AssociateDhcpOptionsOutput, error) {
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
func (m *MockEC2) AssociateDhcpOptionsWithContext(aws.Context, *ec2.AssociateDhcpOptionsInput, ...request.Option) (*ec2.AssociateDhcpOptionsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) AssociateDhcpOptionsRequest(*ec2.AssociateDhcpOptionsInput) (*request.Request, *ec2.AssociateDhcpOptionsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) CreateDhcpOptions(request *ec2.CreateDhcpOptionsInput) (*ec2.CreateDhcpOptionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateDhcpOptions: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	n := len(m.DhcpOptions) + 1

	dhcpOptions := &ec2.DhcpOptions{
		DhcpOptionsId: s(fmt.Sprintf("dopt-%d", n)),
	}

	for _, o := range request.DhcpConfigurations {
		c := &ec2.DhcpConfiguration{
			Key: o.Key,
		}
		for _, v := range o.Values {
			c.Values = append(c.Values, &ec2.AttributeValue{
				Value: v,
			})
		}
		dhcpOptions.DhcpConfigurations = append(dhcpOptions.DhcpConfigurations, c)
	}
	if m.DhcpOptions == nil {
		m.DhcpOptions = make(map[string]*ec2.DhcpOptions)
	}
	m.DhcpOptions[*dhcpOptions.DhcpOptionsId] = dhcpOptions

	copy := *dhcpOptions
	copy.Tags = m.getTags(ec2.ResourceTypeDhcpOptions, *dhcpOptions.DhcpOptionsId)
	return &ec2.CreateDhcpOptionsOutput{DhcpOptions: &copy}, nil
}
func (m *MockEC2) CreateDhcpOptionsWithContext(aws.Context, *ec2.CreateDhcpOptionsInput, ...request.Option) (*ec2.CreateDhcpOptionsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) CreateDhcpOptionsRequest(*ec2.CreateDhcpOptionsInput) (*request.Request, *ec2.CreateDhcpOptionsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteDhcpOptions(request *ec2.DeleteDhcpOptionsInput) (*ec2.DeleteDhcpOptionsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteDhcpOptions: %v", request)

	id := aws.StringValue(request.DhcpOptionsId)
	o := m.DhcpOptions[id]
	if o == nil {
		return nil, fmt.Errorf("DhcpOptions %q not found", id)
	}
	delete(m.DhcpOptions, id)

	return &ec2.DeleteDhcpOptionsOutput{}, nil
}

func (m *MockEC2) DeleteDhcpOptionsWithContext(aws.Context, *ec2.DeleteDhcpOptionsInput, ...request.Option) (*ec2.DeleteDhcpOptionsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DeleteDhcpOptionsRequest(*ec2.DeleteDhcpOptionsInput) (*request.Request, *ec2.DeleteDhcpOptionsOutput) {
	panic("Not implemented")
}
