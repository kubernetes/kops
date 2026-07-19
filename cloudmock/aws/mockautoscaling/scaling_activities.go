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
)

// DescribeScalingActivities returns the canned activities for the named group.
// To exercise pagination, the mock returns one activity per page when the input
// has MaxRecords == 1.
func (m *MockAutoscaling) DescribeScalingActivities(ctx context.Context, input *autoscaling.DescribeScalingActivitiesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.DescribeScalingActivitiesCalls++

	name := aws.ToString(input.AutoScalingGroupName)
	activities := m.ScalingActivities[name]

	start := 0
	if input.NextToken != nil {
		s, err := strconv.Atoi(aws.ToString(input.NextToken))
		if err != nil {
			return nil, fmt.Errorf("invalid NextToken %q: %w", aws.ToString(input.NextToken), err)
		}
		start = s
	}

	pageSize := len(activities) - start
	if input.MaxRecords != nil && int(*input.MaxRecords) > 0 && int(*input.MaxRecords) < pageSize {
		pageSize = int(*input.MaxRecords)
	}
	end := start + pageSize
	if end > len(activities) {
		end = len(activities)
	}

	out := &autoscaling.DescribeScalingActivitiesOutput{
		Activities: activities[start:end],
	}
	if end < len(activities) {
		out.NextToken = aws.String(strconv.Itoa(end))
	}
	return out, nil
}
