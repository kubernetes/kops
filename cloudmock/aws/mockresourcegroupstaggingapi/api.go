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
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"k8s.io/klog"
)

type MockResourceGroupsTaggingAPI struct {
	resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI

	mutex sync.Mutex
}

var _ resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI = &MockResourceGroupsTaggingAPI{}

func (m *MockResourceGroupsTaggingAPI) GetResources(input *resourcegroupstaggingapi.GetResourcesInput) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(2).Infof("mock getResources: %v", input)

	return &resourcegroupstaggingapi.GetResourcesOutput{
		ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{{
			ResourceARN: aws.String("arn:aws:elasticloadbalancing:region:accountId:loadbalancer/loadBalancerName"),
		}},
	}, nil
}
