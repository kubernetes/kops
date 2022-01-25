package mockresourcegroupstaggingapi

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
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
			ResourceARN: aws.String(arn.ARN{
				Partition: "aws",
				Service:   aws.StringValueSlice(input.ResourceTypeFilters)[0],
				Region:    "us-test-1",
				AccountID: "0000000000",
				Resource:  "test",
			}.String()),
		}},
	}, nil
}
