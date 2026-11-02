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

package mockautoscaling

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"k8s.io/klog/v2"
)

func (m *MockAutoscaling) DescribeTags(ctx context.Context, request *autoscaling.DescribeTagsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &autoscaling.DescribeTagsOutput{}
	for _, g := range m.Groups {
		for _, tag := range g.Tags {
			allFiltersMatch := true
			for _, filter := range request.Filters {
				match := false
				switch aws.ToString(filter.Name) {
				case "value":
					for _, v := range filter.Values {
						if aws.ToString(tag.Value) == v {
							match = true
						}
					}

				default:
					klog.Fatalf("Unsupported filter: %v", filter)
				}

				if !match {
					allFiltersMatch = false
					break
				}
			}

			if !allFiltersMatch {
				continue
			}

			response.Tags = append(response.Tags, tag)
		}
	}

	return response, nil
}

func (m *MockAutoscaling) CreateOrUpdateTags(ctx context.Context, request *autoscaling.CreateOrUpdateTagsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	klog.V(2).Infof("Mock CreateOrUpdateTags %v", request)

	for _, tag := range request.Tags {
		g := m.Groups[aws.ToString(tag.ResourceId)]
		if g == nil {
			return nil, fmt.Errorf("AutoScaling Group not found: %q", aws.ToString(tag.ResourceId))
		}
		desc := autoscalingtypes.TagDescription{
			Key:               tag.Key,
			PropagateAtLaunch: tag.PropagateAtLaunch,
			ResourceId:        tag.ResourceId,
			ResourceType:      tag.ResourceType,
			Value:             tag.Value,
		}
		i := slices.IndexFunc(g.Tags, func(t autoscalingtypes.TagDescription) bool {
			return aws.ToString(t.Key) == aws.ToString(tag.Key)
		})
		if i >= 0 {
			g.Tags[i] = desc
		} else {
			g.Tags = append(g.Tags, desc)
		}
	}

	return &autoscaling.CreateOrUpdateTagsOutput{}, nil
}

func (m *MockAutoscaling) DeleteTags(ctx context.Context, request *autoscaling.DeleteTagsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DeleteTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	klog.V(2).Infof("Mock DeleteTags %v", request)

	for _, tag := range request.Tags {
		g := m.Groups[aws.ToString(tag.ResourceId)]
		if g == nil {
			return nil, fmt.Errorf("AutoScaling Group not found: %q", aws.ToString(tag.ResourceId))
		}
		g.Tags = slices.DeleteFunc(g.Tags, func(t autoscalingtypes.TagDescription) bool {
			return aws.ToString(t.Key) == aws.ToString(tag.Key)
		})
	}

	return &autoscaling.DeleteTagsOutput{}, nil
}
