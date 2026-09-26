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
	"sync"

	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type MockAutoscaling struct {
	// Mock out interface
	awsinterfaces.AutoScalingAPI

	mutex             sync.Mutex
	Groups            map[string]*autoscalingtypes.AutoScalingGroup
	WarmPoolInstances map[string][]autoscalingtypes.Instance
	LifecycleHooks    map[string]*autoscalingtypes.LifecycleHook
	// ScalingActivities is the canned response for DescribeScalingActivities,
	// keyed by Auto Scaling group name.
	ScalingActivities map[string][]autoscalingtypes.Activity
	// ScalingActivityPageSize, when positive, makes DescribeScalingActivities
	// paginate at this size regardless of the caller's MaxRecords. Callers that
	// do not set MaxRecords are otherwise served in a single page, which leaves
	// their pagination behaviour untestable.
	ScalingActivityPageSize int

	describeScalingActivitiesCalls int
}

// ScalingActivityCalls returns how many DescribeScalingActivities requests the
// mock has served, including paginated continuations. Safe to call while other
// goroutines are driving the mock.
func (m *MockAutoscaling) ScalingActivityCalls() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.describeScalingActivitiesCalls
}

var _ awsinterfaces.AutoScalingAPI = &MockAutoscaling{}
