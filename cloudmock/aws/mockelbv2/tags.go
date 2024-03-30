/*
Copyright 2020 The Kubernetes Authors.

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

package mockelbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) AddTags(ctx context.Context, request *elbv2.AddTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.AddTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AddTags v2 %v", request)

	if m.Tags == nil {
		m.Tags = make(map[string]elbv2types.TagDescription)
	}

	for _, arn := range request.ResourceArns {
		if t, ok := m.Tags[arn]; ok {
			for _, reqTag := range request.Tags {
				found := false
				for _, tag := range t.Tags {
					if aws.ToString(reqTag.Key) == aws.ToString(tag.Key) {
						tag.Value = reqTag.Value
					}
				}
				if !found {
					tags := m.Tags[arn]
					tags.Tags = append(m.Tags[arn].Tags, reqTag)
				}
			}
		} else {
			m.Tags[arn] = elbv2types.TagDescription{
				ResourceArn: aws.String(arn),
				Tags:        request.Tags,
			}
		}
	}

	return &elbv2.AddTagsOutput{}, nil
}

func (m *MockELBV2) DescribeTags(ctx context.Context, request *elbv2.DescribeTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTags v2 %v", request)

	resp := &elbv2.DescribeTagsOutput{
		TagDescriptions: make([]elbv2types.TagDescription, 0),
	}
	for tagARN, tagDesc := range m.Tags {
		for _, reqARN := range request.ResourceArns {
			if tagARN == reqARN {
				resp.TagDescriptions = append(resp.TagDescriptions, tagDesc)
			}
		}
	}
	return resp, nil
}
