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
	"context"
	"time"
)

// CloudGroupError describes a failure the cloud provider encountered while
// trying to provision an instance in a CloudInstanceGroup. Identical errors
// (same Code + Message) are aggregated into a single entry.
type CloudGroupError struct {
	// Code is the cloud provider's structured error code (e.g.
	// "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS").
	Code string
	// Message is the human-readable error message.
	Message string
	// Instance is the name of the most recent affected instance, if known.
	Instance string
	// Count is the number of times an identical error was observed.
	Count int
	// FirstSeen and LastSeen bracket the observed occurrences.
	FirstSeen time.Time
	LastSeen  time.Time
}

// CloudGroupErrorReporter is an optional capability exposed by cloud
// implementations to surface provisioning errors during cluster validation.
//
// Validation calls this only for groups that have fewer instances than their
// target size, so the implementation may assume something is already wrong.
type CloudGroupErrorReporter interface {
	GetCloudGroupErrors(ctx context.Context, group *CloudInstanceGroup) ([]CloudGroupError, error)
}
