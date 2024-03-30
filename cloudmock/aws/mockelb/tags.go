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

package mockelb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"k8s.io/klog/v2"
)

func (m *MockELB) DescribeTags(ctx context.Context, request *elb.DescribeTagsInput, optFns ...func(*elb.Options)) (*elb.DescribeTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTags %v", request)

	var tags []elbtypes.TagDescription

	for k, lb := range m.LoadBalancers {
		match := false
		if len(request.LoadBalancerNames) == 0 {
			match = true
		} else {
			for _, name := range request.LoadBalancerNames {
				if name == k {
					match = true
				}
			}
		}

		if !match {
			continue
		}

		tagDescription := elbtypes.TagDescription{
			LoadBalancerName: aws.String(k),
		}
		for k, v := range lb.tags {
			tagDescription.Tags = append(tagDescription.Tags, elbtypes.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		tags = append(tags, tagDescription)
	}

	response := &elb.DescribeTagsOutput{
		TagDescriptions: tags,
	}

	return response, nil
}

func (m *MockELB) AddTags(ctx context.Context, request *elb.AddTagsInput, optFns ...func(*elb.Options)) (*elb.AddTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AddTags %v", request)

	for _, name := range request.LoadBalancerNames {
		elb := m.LoadBalancers[name]
		if elb == nil {
			return nil, fmt.Errorf("ELB %q not found", name)
		}
		for _, tag := range request.Tags {
			elb.tags[*tag.Key] = *tag.Value
		}
	}

	return &elb.AddTagsOutput{}, nil
}
