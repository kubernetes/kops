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
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

func (m *MockEC2) CreateTags(ctx context.Context, request *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateTags %v", request)

	for _, v := range request.Resources {
		m.addTags(v, request.Tags...)
	}
	response := &ec2.CreateTagsOutput{}
	return response, nil
}

func (m *MockEC2) addTags(resourceId string, tags ...ec2types.Tag) {
	var resourceType ec2types.ResourceType
	if strings.HasPrefix(resourceId, "subnet-") {
		resourceType = ec2types.ResourceTypeSubnet
	} else if strings.HasPrefix(resourceId, "vpc-") {
		resourceType = ec2types.ResourceTypeVpc
	} else if strings.HasPrefix(resourceId, "sg-") {
		resourceType = ec2types.ResourceTypeSecurityGroup
	} else if strings.HasPrefix(resourceId, "vol-") {
		resourceType = ec2types.ResourceTypeVolume
	} else if strings.HasPrefix(resourceId, "igw-") {
		resourceType = ec2types.ResourceTypeInternetGateway
	} else if strings.HasPrefix(resourceId, "eigw-") {
		resourceType = ec2types.ResourceTypeEgressOnlyInternetGateway
	} else if strings.HasPrefix(resourceId, "nat-") {
		resourceType = ec2types.ResourceTypeNatgateway
	} else if strings.HasPrefix(resourceId, "dopt-") {
		resourceType = ec2types.ResourceTypeDhcpOptions
	} else if strings.HasPrefix(resourceId, "rtb-") {
		resourceType = ec2types.ResourceTypeRouteTable
	} else if strings.HasPrefix(resourceId, "eipalloc-") {
		resourceType = ec2types.ResourceTypeElasticIp
	} else if strings.HasPrefix(resourceId, "lt-") {
		resourceType = ec2types.ResourceTypeLaunchTemplate
	} else if strings.HasPrefix(resourceId, "key-") {
		resourceType = ec2types.ResourceTypeKeyPair
	} else {
		klog.Fatalf("Unknown resource-type in create tags: %v", resourceId)
	}
	for _, tag := range tags {
		t := &ec2types.TagDescription{
			Key:          tag.Key,
			Value:        tag.Value,
			ResourceId:   s(resourceId),
			ResourceType: resourceType,
		}
		m.Tags = append(m.Tags, t)
	}
}

func (m *MockEC2) hasTag(resourceType ec2types.ResourceType, resourceId string, filter ec2types.Filter) bool {
	name := *filter.Name
	if strings.HasPrefix(name, "tag:") {
		tagKey := name[4:]

		for _, tag := range m.Tags {
			if *tag.ResourceId != resourceId {
				continue
			}
			if tag.ResourceType != resourceType {
				continue
			}
			if *tag.Key != tagKey {
				continue
			}

			for _, v := range filter.Values {
				if *tag.Value == v {
					return true
				}
			}
		}
	} else if name == "tag-key" {
		for _, tag := range m.Tags {
			if *tag.ResourceId != resourceId {
				continue
			}
			if tag.ResourceType != resourceType {
				continue
			}
			for _, v := range filter.Values {
				if *tag.Key == v {
					return true
				}
			}
		}
	} else {
		klog.Fatalf("Unsupported filter: %v", filter)
	}
	return false
}

func (m *MockEC2) getTags(resourceType ec2types.ResourceType, resourceId string) []ec2types.Tag {
	var tags []ec2types.Tag
	for _, tag := range m.Tags {
		if *tag.ResourceId != resourceId {
			continue
		}
		if tag.ResourceType != resourceType {
			continue
		}

		t := ec2types.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		}
		tags = append(tags, t)
	}
	return tags
}

func (m *MockEC2) DescribeTags(ctx context.Context, request *ec2.DescribeTagsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTags %v", request)

	var tags []ec2types.TagDescription

	for _, tag := range m.Tags {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "key":
				for _, v := range filter.Values {
					if v == *tag.Key {
						match = true
					}
				}

			case "resource-id":
				for _, v := range filter.Values {
					if v == *tag.ResourceId {
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
		tags = append(tags, copy)
	}

	response := &ec2.DescribeTagsOutput{
		Tags: tags,
	}

	return response, nil
}

// SortTags sorts the slice of tags by Key
func SortTags(tags []ec2types.Tag) {
	keys := make([]string, len(tags))
	for i := range tags {
		keys[i] = aws.ToString(tags[i].Key)
	}
	sort.SliceStable(tags, func(i, j int) bool { return keys[i] < keys[j] })
}

func tagSpecificationsToTags(specifications []ec2types.TagSpecification, resourceType ec2types.ResourceType) []ec2types.Tag {
	tags := make([]ec2types.Tag, 0)
	for _, specification := range specifications {
		if specification.ResourceType != resourceType {
			continue
		}
		tags = append(tags, specification.Tags...)
	}
	return tags
}
