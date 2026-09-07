/*
Copyright 2026 The Kubernetes Authors.

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
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

// DescribeScalingActivities returns the canned activities for the named group,
// paginating at ScalingActivityPageSize or the caller's MaxRecords, whichever
// is smaller and positive.
func (m *MockAutoscaling) DescribeScalingActivities(ctx context.Context, input *autoscaling.DescribeScalingActivitiesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.describeScalingActivitiesCalls++

	name := aws.ToString(input.AutoScalingGroupName)
	activities := m.ScalingActivities[name]

	start := 0
	if input.NextToken != nil {
		s, err := strconv.Atoi(aws.ToString(input.NextToken))
		if err != nil {
			return nil, fmt.Errorf("invalid NextToken %q: %w", aws.ToString(input.NextToken), err)
		}
		if s < 0 || s > len(activities) {
			return nil, fmt.Errorf("invalid NextToken %q: out of range for %d activities", aws.ToString(input.NextToken), len(activities))
		}
		start = s
	}

	end := len(activities)
	if m.ScalingActivityPageSize > 0 {
		end = min(start+m.ScalingActivityPageSize, end)
	}
	if input.MaxRecords != nil && *input.MaxRecords > 0 {
		end = min(start+int(*input.MaxRecords), end)
	}

	// Copy: the caller iterates the page without holding the mock's lock.
	page := make([]autoscalingtypes.Activity, end-start)
	copy(page, activities[start:end])

	out := &autoscaling.DescribeScalingActivitiesOutput{
		Activities: page,
	}
	if end < len(activities) {
		out.NextToken = aws.String(strconv.Itoa(end))
	}
	return out, nil
}
