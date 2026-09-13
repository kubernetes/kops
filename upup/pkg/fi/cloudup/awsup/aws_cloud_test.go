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
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/google/go-cmp/cmp"

	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	"k8s.io/kops/pkg/cloudinstances"
)

func TestGetGroupFailures(t *testing.T) {
	asgName := "test-asg"
	asg := &autoscalingtypes.AutoScalingGroup{AutoScalingGroupName: aws.String(asgName)}

	t0 := time.Date(2026, 5, 12, 20, 35, 0, 0, time.UTC) // before watermark
	t1 := time.Date(2026, 5, 12, 20, 45, 0, 0, time.UTC) // watermark (matches a LaunchTime)
	t2 := time.Date(2026, 5, 12, 20, 50, 0, 0, time.UTC) // after watermark
	t3 := time.Date(2026, 5, 12, 20, 55, 0, 0, time.UTC) // after watermark

	insufficientCapacity := "Launching a new EC2 instance: i-0abc1234. Status Reason: Insufficient capacity."

	tests := []struct {
		name       string
		group      *cloudinstances.CloudInstanceGroup
		activities []autoscalingtypes.Activity
		wantErrors []cloudinstances.GroupFailure
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
			wantErrors: []cloudinstances.GroupFailure{
				{
					Code:      string(autoscalingtypes.ScalingActivityStatusCodeFailed),
					Message:   insufficientCapacity,
					Instance:  "i-0abc1234",
					Count:     2,
					FirstSeen: t0,
					LastSeen:  t3,
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
			wantErrors: []cloudinstances.GroupFailure{
				{
					Code:      string(autoscalingtypes.ScalingActivityStatusCodeFailed),
					Message:   insufficientCapacity,
					Instance:  "i-0abc1234",
					Count:     1,
					FirstSeen: t2,
					LastSeen:  t2,
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
			wantErrors: []cloudinstances.GroupFailure{
				{
					Code:      string(autoscalingtypes.ScalingActivityStatusCodeFailed),
					Message:   insufficientCapacity,
					Instance:  "i-0abc1234",
					Count:     1,
					FirstSeen: t2,
					LastSeen:  t2,
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
			wantErrors: []cloudinstances.GroupFailure{
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
			wantErrors: []cloudinstances.GroupFailure{
				{Code: "Cancelled", Message: "cancelled", Count: 1, FirstSeen: t3, LastSeen: t3},
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
			got, err := cloud.GetGroupFailures(t.Context(), tt.group)
			if err != nil {
				t.Fatalf("GetGroupFailures returned error: %v", err)
			}
			if diff := cmp.Diff(tt.wantErrors, got); diff != "" {
				t.Errorf("GetGroupFailures(%s) mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestGetGroupFailuresStopsPaginatingPastWatermark(t *testing.T) {
	asgName := "test-asg"
	asg := &autoscalingtypes.AutoScalingGroup{AutoScalingGroupName: aws.String(asgName)}

	watermark := time.Date(2026, 5, 12, 20, 45, 0, 0, time.UTC)
	tBefore := watermark.Add(-1 * time.Hour)
	tAfter := watermark.Add(1 * time.Minute)

	// 3 recent failures then 5 older ones, most-recent-first as AWS returns them.
	// Everything strictly before the existing instance's CreationTimestamp must
	// be filtered out, and reaching the first of them must stop pagination.
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
	asgMock := &mockautoscaling.MockAutoscaling{
		ScalingActivities: map[string][]autoscalingtypes.Activity{asgName: activities},
		// Two activities per page, so 8 activities span 4 pages. Production sets
		// no MaxRecords, so without this the mock serves everything at once and
		// the short-circuit is indistinguishable from a plain inner-loop break.
		ScalingActivityPageSize: 2,
	}
	cloud.MockAutoscaling = asgMock

	got, err := getGroupFailures(t.Context(), cloud, &cloudinstances.CloudInstanceGroup{
		Raw: asg,
		Ready: []*cloudinstances.CloudInstance{
			{ID: "i-existing", CreationTimestamp: watermark},
		},
	})
	if err != nil {
		t.Fatalf("getGroupFailures(%q) error = %v, want nil", asgName, err)
	}

	gotMessages := make([]string, 0, len(got))
	for _, f := range got {
		gotMessages = append(gotMessages, f.Message)
	}
	slices.Sort(gotMessages)
	wantMessages := []string{"err A", "err B", "err C"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Errorf("getGroupFailures(%q) messages = %v, want %v", asgName, gotMessages, wantMessages)
	}

	// Page 1 is the first two recent failures; page 2 holds the third plus the
	// first pre-watermark one, which ends the scan. Pages 3 and 4 must never be
	// requested.
	if calls := asgMock.ScalingActivityCalls(); calls != 2 {
		t.Errorf("getGroupFailures(%q) made %d DescribeScalingActivities calls, want 2", asgName, calls)
	}
}

func TestGetGroupFailuresCapsPagination(t *testing.T) {
	asgName := "test-asg"
	asg := &autoscalingtypes.AutoScalingGroup{AutoScalingGroupName: aws.String(asgName)}

	// An empty group has no watermark, so the short-circuit above never fires.
	// The page cap is the only thing bounding the scan.
	ts := time.Date(2026, 5, 12, 20, 45, 0, 0, time.UTC)
	activities := make([]autoscalingtypes.Activity, 40)
	for i := range activities {
		activities[i] = autoscalingtypes.Activity{
			StartTime:     &ts,
			StatusCode:    autoscalingtypes.ScalingActivityStatusCodeFailed,
			StatusMessage: aws.String(fmt.Sprintf("failure %d", i)),
		}
	}

	cloud := BuildMockAWSCloud("us-east-1", "a")
	asgMock := &mockautoscaling.MockAutoscaling{
		ScalingActivities:       map[string][]autoscalingtypes.Activity{asgName: activities},
		ScalingActivityPageSize: 1,
	}
	cloud.MockAutoscaling = asgMock

	if _, err := getGroupFailures(t.Context(), cloud, &cloudinstances.CloudInstanceGroup{Raw: asg}); err != nil {
		t.Fatalf("getGroupFailures(%q) error = %v, want nil", asgName, err)
	}
	if calls := asgMock.ScalingActivityCalls(); calls != maxScalingActivityPages {
		t.Errorf("getGroupFailures(%q) made %d DescribeScalingActivities calls, want the %d-page cap",
			asgName, calls, maxScalingActivityPages)
	}
}
