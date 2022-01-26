/*
Copyright 2022 The Kubernetes Authors.

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

package mockresourcegroupstaggingapi

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"k8s.io/klog"
	"k8s.io/kops/cloudmock/aws/mockelb"
)

type MockResourceGroupsTaggingAPI struct {
	resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI

	ElbLoadBalancers map[string]*mockelb.LoadBalancer
	ElbV2Tags        map[string]*elbv2.TagDescription
}

var _ resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI = &MockResourceGroupsTaggingAPI{}

func (m *MockResourceGroupsTaggingAPI) GetResources(input *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {

	klog.V(2).Infof("mock getResources: %v", input)
	klog.Infof("debug mock getResouces loadBalancers: %v, elbv2Tags: %v", m.ElbLoadBalancers, m.ElbV2Tags)

	result := &resourcegroupstaggingapi.GetResourcesOutput{}

	for _, lb := range m.ElbLoadBalancers {
		for _, tagFilter := range input.TagFilters {
			if val, ok := lb.Tags[*tagFilter.Key]; ok {
				for _, tagFilterValue := range tagFilter.Values {
					if val == *tagFilterValue {
						result.ResourceTagMappingList = append(result.ResourceTagMappingList, &resourcegroupstaggingapi.ResourceTagMapping{
							ResourceARN: aws.String(fmt.Sprintf("arn:aws:elasticloadbalancing:region:accountId:loadbalancer/%s", *lb.Description.LoadBalancerName)),
						})
						return result, nil
					}
				}
			}
		}
	}

	for _, tagFilter := range input.TagFilters {
		for _, elbTags := range m.ElbV2Tags {
			for _, tag := range elbTags.Tags {
				if tagFilter.Key == tag.Key {
					for _, value := range tagFilter.Values {
						if tag.Value == value {
							result.ResourceTagMappingList = append(result.ResourceTagMappingList, &resourcegroupstaggingapi.ResourceTagMapping{
								ResourceARN: elbTags.ResourceArn,
							})
							return result, nil
						}
					}
				}
			}
		}
	}

	return result, nil
}
