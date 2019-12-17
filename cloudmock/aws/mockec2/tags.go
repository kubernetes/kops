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
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

const (
	// Not (yet?) in aws-sdk-go
	ResourceTypeNatGateway = "nat-gateway"
	ResourceTypeAddress    = "elastic-ip"
)

func (m *MockEC2) CreateTagsRequest(*ec2.CreateTagsInput) (*request.Request, *ec2.CreateTagsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) CreateTagsWithContext(aws.Context, *ec2.CreateTagsInput, ...request.Option) (*ec2.CreateTagsOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) CreateTags(request *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateTags %v", request)

	for _, v := range request.Resources {
		resourceId := *v
		for _, tag := range request.Tags {
			m.addTag(resourceId, tag)
		}
	}
	response := &ec2.CreateTagsOutput{}
	return response, nil
}

func (m *MockEC2) addTag(resourceId string, tag *ec2.Tag) {
	resourceType := ""
	if strings.HasPrefix(resourceId, "subnet-") {
		resourceType = ec2.ResourceTypeSubnet
	} else if strings.HasPrefix(resourceId, "vpc-") {
		resourceType = ec2.ResourceTypeVpc
	} else if strings.HasPrefix(resourceId, "sg-") {
		resourceType = ec2.ResourceTypeSecurityGroup
	} else if strings.HasPrefix(resourceId, "vol-") {
		resourceType = ec2.ResourceTypeVolume
	} else if strings.HasPrefix(resourceId, "igw-") {
		resourceType = ec2.ResourceTypeInternetGateway
	} else if strings.HasPrefix(resourceId, "nat-") {
		resourceType = ResourceTypeNatGateway
	} else if strings.HasPrefix(resourceId, "dopt-") {
		resourceType = ec2.ResourceTypeDhcpOptions
	} else if strings.HasPrefix(resourceId, "rtb-") {
		resourceType = ec2.ResourceTypeRouteTable
	} else if strings.HasPrefix(resourceId, "eipalloc-") {
		resourceType = ResourceTypeAddress
	} else {
		klog.Fatalf("Unknown resource-type in create tags: %v", resourceId)
	}

	t := &ec2.TagDescription{
		Key:          tag.Key,
		Value:        tag.Value,
		ResourceId:   s(resourceId),
		ResourceType: s(resourceType),
	}
	m.Tags = append(m.Tags, t)
}

func (m *MockEC2) DescribeTagsRequest(*ec2.DescribeTagsInput) (*request.Request, *ec2.DescribeTagsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeTagsWithContext(aws.Context, *ec2.DescribeTagsInput, ...request.Option) (*ec2.DescribeTagsOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) hasTag(resourceType string, resourceId string, filter *ec2.Filter) bool {
	name := *filter.Name
	if strings.HasPrefix(name, "tag:") {
		tagKey := name[4:]

		for _, tag := range m.Tags {
			if *tag.ResourceId != resourceId {
				continue
			}
			if *tag.ResourceType != resourceType {
				continue
			}
			if *tag.Key != tagKey {
				continue
			}

			for _, v := range filter.Values {
				if *tag.Value == *v {
					return true
				}
			}
		}
	} else if name == "tag-key" {
		for _, tag := range m.Tags {
			if *tag.ResourceId != resourceId {
				continue
			}
			if *tag.ResourceType != resourceType {
				continue
			}
			for _, v := range filter.Values {
				if *tag.Key == *v {
					return true
				}
			}
		}
	} else {
		klog.Fatalf("Unsupported filter: %v", filter)
	}
	return false
}

func (m *MockEC2) getTags(resourceType string, resourceId string) []*ec2.Tag {
	var tags []*ec2.Tag
	for _, tag := range m.Tags {
		if *tag.ResourceId != resourceId {
			continue
		}
		if *tag.ResourceType != resourceType {
			continue
		}

		t := &ec2.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		}
		tags = append(tags, t)
	}
	return tags
}

func (m *MockEC2) DescribeTags(request *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTags %v", request)

	var tags []*ec2.TagDescription

	for _, tag := range m.Tags {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "key":
				for _, v := range filter.Values {
					if *v == *tag.Key {
						match = true
					}
				}

			case "resource-id":
				for _, v := range filter.Values {
					if *v == *tag.ResourceId {
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

		copy := *tag
		tags = append(tags, &copy)
	}

	response := &ec2.DescribeTagsOutput{
		Tags: tags,
	}

	return response, nil
}
func (m *MockEC2) DescribeTagsPages(*ec2.DescribeTagsInput, func(*ec2.DescribeTagsOutput, bool) bool) error {
	panic("Not implemented")
}
func (m *MockEC2) DescribeTagsPagesWithContext(aws.Context, *ec2.DescribeTagsInput, func(*ec2.DescribeTagsOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

// SortTags sorts the slice of tags by Key
func SortTags(tags []*ec2.Tag) {
	keys := make([]string, len(tags))
	for i := range tags {
		if tags[i] != nil {
			keys[i] = aws.StringValue(tags[i].Key)
		}
	}
	sort.SliceStable(tags, func(i, j int) bool { return keys[i] < keys[j] })
}
