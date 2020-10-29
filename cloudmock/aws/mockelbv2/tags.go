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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) AddTags(request *elbv2.AddTagsInput) (*elbv2.AddTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AddTags v2 %v", request)

	if m.Tags == nil {
		m.Tags = make(map[string]*elbv2.TagDescription)
	}

	for _, reqARN := range request.ResourceArns {
		arn := aws.StringValue(reqARN)
		if t, ok := m.Tags[arn]; ok {
			for _, reqTag := range request.Tags {
				found := false
				for _, tag := range t.Tags {
					if aws.StringValue(reqTag.Key) == aws.StringValue(tag.Key) {
						tag.Value = reqTag.Value
					}
				}
				if !found {
					m.Tags[arn].Tags = append(m.Tags[arn].Tags, reqTag)
				}
			}
		} else {
			m.Tags[arn] = &elbv2.TagDescription{
				ResourceArn: reqARN,
				Tags:        request.Tags,
			}
		}
	}

	return &elbv2.AddTagsOutput{}, nil
}

func (m *MockELBV2) DescribeTags(request *elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeTags v2 %v", request)

	resp := &elbv2.DescribeTagsOutput{
		TagDescriptions: make([]*elbv2.TagDescription, 0),
	}
	for tagARN, tagDesc := range m.Tags {
		for _, reqARN := range request.ResourceArns {
			if tagARN == aws.StringValue(reqARN) {
				resp.TagDescriptions = append(resp.TagDescriptions, tagDesc)
			}
		}
	}
	return resp, nil
}
