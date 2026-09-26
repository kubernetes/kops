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

package cloudinstances

import (
	"cmp"
	"context"
	"slices"
	"time"
)

// GroupFailure describes a failure the cloud provider encountered while
// trying to provision an instance in a CloudInstanceGroup. Identical failures
// (same Code + Message) are aggregated into a single entry.
//
// It is a report rather than a Go error: it is never returned in an error
// position and does not implement the error interface.
type GroupFailure struct {
	// Code is the cloud provider's structured error code (e.g.
	// "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS").
	Code string `json:"code"`
	// Message is the human-readable error message.
	Message string `json:"message"`
	// Instance is the name of the most recent affected instance, if known.
	Instance string `json:"instance,omitempty"`
	// Count is the number of times an identical error was observed.
	Count int `json:"count"`
	// FirstSeen and LastSeen bracket the observed occurrences.
	FirstSeen time.Time `json:"firstSeen,omitzero"`
	LastSeen  time.Time `json:"lastSeen,omitzero"`
}

// Watermark returns the most recent instance-creation time in the group, used
// to discard provisioning failures left over from an earlier, healthier
// incarnation of the group. A group with no instances has a zero watermark, so
// callers surface the provider's full retention window — precisely the
// empty-group failure mode this reporting exists to diagnose.
func Watermark(group *CloudInstanceGroup) time.Time {
	var w time.Time
	for _, m := range slices.Concat(group.Ready, group.NeedUpdate) {
		if m.CreationTimestamp.After(w) {
			w = m.CreationTimestamp
		}
	}
	return w
}

// SortFailures orders failures most-recent-first with a total, deterministic
// tie-break. LastSeen alone is not a total order — distinct failures commonly
// share a timestamp — so without the Code/Message tie-break the validation
// output and the dump artifact would reorder run to run.
func SortFailures(failures []GroupFailure) {
	slices.SortFunc(failures, func(a, b GroupFailure) int {
		return cmp.Or(
			b.LastSeen.Compare(a.LastSeen),
			cmp.Compare(a.Code, b.Code),
			cmp.Compare(a.Message, b.Message),
		)
	})
}

// GroupFailureReporter is an optional capability exposed by cloud
// implementations to surface provisioning errors for a group.
//
// Cluster validation calls this only for groups short of their target size,
// but "kops toolbox dump" calls it for every group, so implementations must
// return cleanly for a healthy group rather than assuming a failure.
type GroupFailureReporter interface {
	GetGroupFailures(ctx context.Context, group *CloudInstanceGroup) ([]GroupFailure, error)
}

// GroupInstanceStatus is the cloud provider's view of an instance the
// group is managing. Unlike CloudInstance this covers instances that do not
// exist: a GCE MIG retrying a failed create reports each attempt here, and
// those attempts never become CloudInstances because no instance resource was
// ever created for them.
type GroupInstanceStatus struct {
	Name string `json:"name"`
	// Status is the instance's own state, e.g. "RUNNING". Empty while the
	// instance does not exist.
	Status string `json:"status,omitempty"`
	// CurrentAction is what the group is doing with the instance, e.g.
	// "CREATING".
	CurrentAction string `json:"currentAction,omitempty"`
	// Failures are from the group's most recent attempt to act on the instance.
	// These are raw per-attempt errors rather than aggregated observations, so
	// their Count is unset.
	Failures []GroupFailure `json:"failures,omitempty"`
}

// GroupInstanceStatusReporter is an optional capability exposed by cloud
// implementations that can report on instances a group is managing but has not
// successfully created.
type GroupInstanceStatusReporter interface {
	GetGroupInstanceStatuses(ctx context.Context, group *CloudInstanceGroup) ([]GroupInstanceStatus, error)
}
