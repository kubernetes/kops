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

package awsup

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"

	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	"k8s.io/kops/pkg/cloudinstances"
)

func TestGetCloudGroupErrors(t *testing.T) {
	asgName := "test-asg"
	asg := &autoscalingtypes.AutoScalingGroup{AutoScalingGroupName: aws.String(asgName)}

	t0 := time.Date(2026, 5, 12, 20, 35, 0, 0, time.UTC) // before watermark
	t1 := time.Date(2026, 5, 12, 20, 45, 0, 0, time.UTC) // watermark (matches a LaunchTime)
	t2 := time.Date(2026, 5, 12, 20, 50, 0, 0, time.UTC) // after watermark
	t3 := time.Date(2026, 5, 12, 20, 55, 0, 0, time.UTC) // after watermark

	insufficientCapacity := "Launching a new EC2 instance: i-0abc1234. Status Reason: Insufficient capacity."

	tests := []struct {
		name              string
		group             *cloudinstances.CloudInstanceGroup
		activities        []autoscalingtypes.Activity
		wantErrors        []cloudinstances.CloudGroupError
		wantInstanceMatch string
	}{
		{
			name: "empty group surfaces all failed activities",
			group: &cloudinstances.CloudInstanceGroup{
				Raw:        asg,
				TargetSize: 2,
			},
			activities: []autoscalingtypes.Activity{
				{StartTime: &t3, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String(insufficientCapacity)},
				{StartTime: &t0, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String(insufficientCapacity)},
			},
			wantErrors: []cloudinstances.CloudGroupError{
				{
					Code:     string(autoscalingtypes.ScalingActivityStatusCodeFailed),
					Message:  insufficientCapacity,
					Instance: "i-0abc1234",
					Count:    2,
				},
			},
		},
		{
			name: "watermark filters out activities older than newest instance",
			group: &cloudinstances.CloudInstanceGroup{
				Raw:        asg,
				TargetSize: 2,
				Ready: []*cloudinstances.CloudInstance{
					{ID: "i-existing", CreationTimestamp: t1},
				},
			},
			activities: []autoscalingtypes.Activity{
				{StartTime: &t2, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String(insufficientCapacity)},
				{StartTime: &t0, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("ancient error")},
			},
			wantErrors: []cloudinstances.CloudGroupError{
				{
					Code:     string(autoscalingtypes.ScalingActivityStatusCodeFailed),
					Message:  insufficientCapacity,
					Instance: "i-0abc1234",
					Count:    1,
				},
			},
		},
		{
			name: "non-failed activities are ignored",
			group: &cloudinstances.CloudInstanceGroup{
				Raw:        asg,
				TargetSize: 2,
			},
			activities: []autoscalingtypes.Activity{
				{StartTime: &t3, StatusCode: autoscalingtypes.ScalingActivityStatusCodeSuccessful, StatusMessage: aws.String("ok")},
				{StartTime: &t2, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String(insufficientCapacity)},
				{StartTime: &t1, StatusCode: autoscalingtypes.ScalingActivityStatusCodeInProgress, StatusMessage: aws.String("still going")},
			},
			wantErrors: []cloudinstances.CloudGroupError{
				{
					Code:     string(autoscalingtypes.ScalingActivityStatusCodeFailed),
					Message:  insufficientCapacity,
					Instance: "i-0abc1234",
					Count:    1,
				},
			},
		},
		{
			name: "identical messages are aggregated",
			group: &cloudinstances.CloudInstanceGroup{
				Raw:        asg,
				TargetSize: 3,
			},
			activities: []autoscalingtypes.Activity{
				{StartTime: &t3, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("same error")},
				{StartTime: &t2, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("same error")},
				{StartTime: &t1, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("different error")},
			},
			wantErrors: []cloudinstances.CloudGroupError{
				{Code: "Failed", Message: "same error", Count: 2, FirstSeen: t2, LastSeen: t3},
				{Code: "Failed", Message: "different error", Count: 1, FirstSeen: t1, LastSeen: t1},
			},
		},
		{
			name: "cancelled activities are surfaced",
			group: &cloudinstances.CloudInstanceGroup{
				Raw:        asg,
				TargetSize: 1,
			},
			activities: []autoscalingtypes.Activity{
				{StartTime: &t3, StatusCode: autoscalingtypes.ScalingActivityStatusCodeCancelled, StatusMessage: aws.String("cancelled")},
			},
			wantErrors: []cloudinstances.CloudGroupError{
				{Code: "Cancelled", Message: "cancelled", Count: 1},
			},
		},
		{
			name:       "non-ASG Raw returns nil",
			group:      &cloudinstances.CloudInstanceGroup{Raw: "not-an-asg"},
			activities: nil,
			wantErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud := BuildMockAWSCloud("us-east-1", "a")
			cloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{
				ScalingActivities: map[string][]autoscalingtypes.Activity{
					asgName: tt.activities,
				},
			}
			got, err := cloud.GetCloudGroupErrors(context.Background(), tt.group)
			if err != nil {
				t.Fatalf("GetCloudGroupErrors returned error: %v", err)
			}
			if len(got) != len(tt.wantErrors) {
				t.Fatalf("got %d errors, want %d: %+v", len(got), len(tt.wantErrors), got)
			}
			for i, want := range tt.wantErrors {
				if got[i].Code != want.Code || got[i].Message != want.Message || got[i].Count != want.Count {
					t.Errorf("error %d: got %+v, want %+v", i, got[i], want)
				}
				if want.Instance != "" && got[i].Instance != want.Instance {
					t.Errorf("error %d instance: got %q, want %q", i, got[i].Instance, want.Instance)
				}
				if !want.FirstSeen.IsZero() && !got[i].FirstSeen.Equal(want.FirstSeen) {
					t.Errorf("error %d FirstSeen: got %v, want %v", i, got[i].FirstSeen, want.FirstSeen)
				}
				if !want.LastSeen.IsZero() && !got[i].LastSeen.Equal(want.LastSeen) {
					t.Errorf("error %d LastSeen: got %v, want %v", i, got[i].LastSeen, want.LastSeen)
				}
			}
		})
	}
}

func TestGetCloudGroupErrors_PaginationShortCircuit(t *testing.T) {
	asgName := "test-asg"
	asg := &autoscalingtypes.AutoScalingGroup{AutoScalingGroupName: aws.String(asgName)}

	watermark := time.Date(2026, 5, 12, 20, 45, 0, 0, time.UTC)
	tBefore := watermark.Add(-1 * time.Hour)
	tAfter := watermark.Add(1 * time.Minute)

	// 3 recent failures + 5 older ones. The watermark must filter out everything
	// at-or-before the existing instance's CreationTimestamp.
	activities := []autoscalingtypes.Activity{
		{StartTime: &tAfter, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("err A")},
		{StartTime: &tAfter, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("err B")},
		{StartTime: &tAfter, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("err C")},
		{StartTime: &tBefore, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("old 1")},
		{StartTime: &tBefore, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("old 2")},
		{StartTime: &tBefore, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("old 3")},
		{StartTime: &tBefore, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("old 4")},
		{StartTime: &tBefore, StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed, StatusMessage: aws.String("old 5")},
	}

	cloud := BuildMockAWSCloud("us-east-1", "a")
	cloud.MockAutoscaling = &mockautoscaling.MockAutoscaling{
		ScalingActivities: map[string][]autoscalingtypes.Activity{asgName: activities},
	}

	got, err := getCloudGroupErrors(context.Background(), cloud, &cloudinstances.CloudInstanceGroup{
		Raw: asg,
		Ready: []*cloudinstances.CloudInstance{
			{ID: "i-existing", CreationTimestamp: watermark},
		},
	})
	if err != nil {
		t.Fatalf("getCloudGroupErrors: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 errors after watermark filter, got %d: %+v", len(got), got)
	}
}
