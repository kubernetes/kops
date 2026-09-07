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

package gce_test

import (
	"testing"
	"time"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/pkg/cloudinstances"
	gceup "k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

const (
	testProject = "testproject"
	testRegion  = "us-test1"
	testZone    = "us-test1-a"
	testMIG     = "a-nodes-testcluster-example-com"
)

func migSelfLink(name string) string {
	return "https://www.googleapis.com/compute/v1/projects/" + testProject +
		"/zones/" + testZone + "/instanceGroupManagers/" + name
}

func igmError(code, message, instance, timestamp string) *compute.InstanceManagedByIgmError {
	e := &compute.InstanceManagedByIgmError{
		Error:     &compute.InstanceManagedByIgmErrorManagedInstanceError{Code: code, Message: message},
		Timestamp: timestamp,
	}
	if instance != "" {
		e.InstanceActionDetails = &compute.InstanceManagedByIgmErrorInstanceActionDetails{
			Instance: "https://www.googleapis.com/compute/v1/projects/" + testProject +
				"/zones/" + testZone + "/instances/" + instance,
		}
	}
	return e
}

func newTestCloud(t *testing.T) *gce.MockGCECloud {
	t.Helper()
	return gce.InstallMockGCECloud(testRegion, testProject)
}

func TestGetGroupFailures(t *testing.T) {
	exhausted := "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS"
	quota := "QUOTA_EXCEEDED"

	watermark := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	before := watermark.Add(-time.Hour).Format(time.RFC3339)
	after := watermark.Add(time.Minute).Format(time.RFC3339)
	later := watermark.Add(2 * time.Minute).Format(time.RFC3339)

	tests := []struct {
		name       string
		group      *cloudinstances.CloudInstanceGroup
		igmErrors  []*compute.InstanceManagedByIgmError
		wantCodes  []string
		wantCounts []int
		wantFirst  string
	}{
		{
			name:  "empty group surfaces every error",
			group: &cloudinstances.CloudInstanceGroup{Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)}},
			igmErrors: []*compute.InstanceManagedByIgmError{
				igmError(exhausted, "the zone does not have enough resources", "node-1", after),
				igmError(exhausted, "the zone does not have enough resources", "node-2", before),
			},
			wantCodes:  []string{exhausted},
			wantCounts: []int{2},
			wantFirst:  exhausted,
		},
		{
			name: "watermark drops errors older than the newest instance",
			group: &cloudinstances.CloudInstanceGroup{
				Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)},
				Ready: []*cloudinstances.CloudInstance{
					{ID: "existing", CreationTimestamp: watermark},
				},
			},
			igmErrors: []*compute.InstanceManagedByIgmError{
				igmError(exhausted, "recent", "node-1", after),
				igmError(quota, "from a previous incarnation of the group", "node-0", before),
			},
			wantCodes:  []string{exhausted},
			wantCounts: []int{1},
			wantFirst:  exhausted,
		},
		{
			name:  "distinct codes are not merged and sort most recent first",
			group: &cloudinstances.CloudInstanceGroup{Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)}},
			igmErrors: []*compute.InstanceManagedByIgmError{
				igmError(exhausted, "capacity", "node-1", after),
				igmError(quota, "quota", "node-2", later),
			},
			wantCodes:  []string{quota, exhausted},
			wantCounts: []int{1, 1},
			wantFirst:  quota,
		},
		{
			name:  "unparseable timestamps do not drop the error",
			group: &cloudinstances.CloudInstanceGroup{Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)}},
			igmErrors: []*compute.InstanceManagedByIgmError{
				igmError(exhausted, "capacity", "node-1", "not-a-timestamp"),
			},
			wantCodes:  []string{exhausted},
			wantCounts: []int{1},
			wantFirst:  exhausted,
		},
		{
			name:      "entries without an Error payload are skipped",
			group:     &cloudinstances.CloudInstanceGroup{Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)}},
			igmErrors: []*compute.InstanceManagedByIgmError{{Timestamp: after}},
			wantCodes: nil,
		},
		{
			name:      "a group whose Raw is not a MIG reports nothing",
			group:     &cloudinstances.CloudInstanceGroup{Raw: "not-a-mig"},
			igmErrors: []*compute.InstanceManagedByIgmError{igmError(exhausted, "capacity", "node-1", after)},
			wantCodes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud := newTestCloud(t)
			cloud.ComputeClient().SetInstanceGroupManagerErrors(testProject, testZone, testMIG, tt.igmErrors)

			got, err := gceup.GetGroupFailures(t.Context(), cloud, tt.group)
			if err != nil {
				t.Fatalf("GetGroupFailures() error = %v, want nil", err)
			}
			if len(got) != len(tt.wantCodes) {
				t.Fatalf("GetGroupFailures() returned %d failures, want %d: %+v", len(got), len(tt.wantCodes), got)
			}
			for i, wantCode := range tt.wantCodes {
				if got[i].Code != wantCode {
					t.Errorf("failure %d: Code = %q, want %q", i, got[i].Code, wantCode)
				}
				if got[i].Count != tt.wantCounts[i] {
					t.Errorf("failure %d: Count = %d, want %d", i, got[i].Count, tt.wantCounts[i])
				}
			}
			if tt.wantFirst != "" && got[0].Code != tt.wantFirst {
				t.Errorf("first failure Code = %q, want %q", got[0].Code, tt.wantFirst)
			}
		})
	}
}

func TestGetGroupFailuresExtractsInstanceName(t *testing.T) {
	cloud := newTestCloud(t)
	cloud.ComputeClient().SetInstanceGroupManagerErrors(testProject, testZone, testMIG,
		[]*compute.InstanceManagedByIgmError{
			igmError("ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS", "capacity", "a-nodes-xyz",
				time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)),
		})

	got, err := gceup.GetGroupFailures(t.Context(), cloud,
		&cloudinstances.CloudInstanceGroup{Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)}})
	if err != nil {
		t.Fatalf("GetGroupFailures() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetGroupFailures() returned %d failures, want 1", len(got))
	}
	// The API reports a full self link; only the last component is useful.
	if got[0].Instance != "a-nodes-xyz" {
		t.Errorf("Instance = %q, want %q", got[0].Instance, "a-nodes-xyz")
	}
}

func TestGetGroupInstanceStatuses(t *testing.T) {
	cloud := newTestCloud(t)

	// A managed instance the MIG keeps failing to create: it has no backing
	// compute instance, so GetCloudGroups drops it and it is invisible without
	// this capability.
	cloud.ComputeClient().SetManagedInstance(testProject, testZone, "a-nodes-pending", &compute.ManagedInstance{
		Name:          "a-nodes-pending",
		Instance:      "https://www.googleapis.com/compute/v1/projects/" + testProject + "/zones/" + testZone + "/instances/a-nodes-pending",
		CurrentAction: "CREATING",
		LastAttempt: &compute.ManagedInstanceLastAttempt{
			Errors: &compute.ManagedInstanceLastAttemptErrors{
				Errors: []*compute.ManagedInstanceLastAttemptErrorsErrors{
					{Code: "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS", Message: "the zone does not have enough resources"},
				},
			},
		},
	})

	got, err := gceup.GetGroupInstanceStatuses(t.Context(), cloud,
		&cloudinstances.CloudInstanceGroup{Raw: &compute.InstanceGroupManager{Name: testMIG, SelfLink: migSelfLink(testMIG)}})
	if err != nil {
		t.Fatalf("GetGroupInstanceStatuses() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetGroupInstanceStatuses() returned %d statuses, want 1: %+v", len(got), got)
	}
	if got[0].Name != "a-nodes-pending" {
		t.Errorf("Name = %q, want %q", got[0].Name, "a-nodes-pending")
	}
	if got[0].CurrentAction != "CREATING" {
		t.Errorf("CurrentAction = %q, want %q", got[0].CurrentAction, "CREATING")
	}
	if len(got[0].Failures) != 1 {
		t.Fatalf("got %d failures, want 1", len(got[0].Failures))
	}
	if got[0].Failures[0].Code != "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS" {
		t.Errorf("failure Code = %q, want %q", got[0].Failures[0].Code, "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS")
	}
	// These are raw per-attempt errors, not aggregated observations.
	if got[0].Failures[0].Count != 0 {
		t.Errorf("failure Count = %d, want 0 (per-attempt errors are not aggregated)", got[0].Failures[0].Count)
	}
}

func TestGetGroupInstanceStatusesNonMIG(t *testing.T) {
	cloud := newTestCloud(t)
	got, err := gceup.GetGroupInstanceStatuses(t.Context(), cloud,
		&cloudinstances.CloudInstanceGroup{Raw: "not-a-mig"})
	if err != nil {
		t.Fatalf("GetGroupInstanceStatuses() error = %v, want nil", err)
	}
	if got != nil {
		t.Errorf("GetGroupInstanceStatuses() = %+v, want nil", got)
	}
}
