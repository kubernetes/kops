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

package spotinst

import (
	"context"
	"time"
)

type (
	// InstanceGroup wraps a cloud-specific instance group object.
	InstanceGroup interface {
		// Id returns the ID of the InstanceGroup.
		Id() string

		// Type returns the type of the InstanceGroup.
		Type() InstanceGroupType

		// Name returns the name of the InstanceGroup.
		Name() string

		// MinSize returns the minimum size of the InstanceGroup.
		MinSize() int

		// MaxSize returns the maximum size of the InstanceGroup.
		MaxSize() int

		// CreatedAt returns the timestamp when the InstanceGroup has been created.
		CreatedAt() time.Time

		// UpdatedAt returns the timestamp when the InstanceGroup has been updated.
		UpdatedAt() time.Time

		// Obj returns the raw object which is a cloud-specific implementation.
		Obj() interface{}
	}

	// Instance wraps a cloud-specific instance object.
	Instance interface {
		// Id returns the ID of the instance.
		Id() string

		// CreatedAt returns the timestamp when the Instance has been created.
		CreatedAt() time.Time

		// Obj returns the raw object which is a cloud-specific implementation.
		Obj() interface{}
	}

	// LaunchSpec wraps a cloud-specific launch specification object.
	LaunchSpec interface {
		// Id returns the ID of the LaunchSpec.
		Id() string

		// Name returns the name of the LaunchSpec.
		Name() string

		// OceanId returns the ID of the Ocean instance group.
		OceanId() string

		// CreatedAt returns the timestamp when the LaunchSpec has been created.
		CreatedAt() time.Time

		// UpdatedAt returns the timestamp when the LaunchSpec has been updated.
		UpdatedAt() time.Time

		// Obj returns the raw object which is a cloud-specific implementation.
		Obj() interface{}
	}

	// Cloud wraps all services provided by Spotinst.
	Cloud interface {
		// Elastigroups returns a new Elastigroup service.
		Elastigroup() InstanceGroupService

		// Ocean returns a new Ocean service.
		Ocean() InstanceGroupService

		// LaunchSpec returns a new LaunchSpec service.
		LaunchSpec() LaunchSpecService
	}

	// InstanceGroupService wraps all common functionality for InstanceGroups.
	InstanceGroupService interface {
		// List returns a list of InstanceGroups.
		List(ctx context.Context) ([]InstanceGroup, error)

		// Create creates a new InstanceGroup and returns its ID.
		Create(ctx context.Context, group InstanceGroup) (string, error)

		// Read returns an existing InstanceGroup by ID.
		Read(ctx context.Context, groupID string) (InstanceGroup, error)

		// Update updates an existing InstanceGroup.
		Update(ctx context.Context, group InstanceGroup) error

		// Delete deletes an existing InstanceGroup by ID.
		Delete(ctx context.Context, groupID string) error

		// Detach removes one or more instances from the specified InstanceGroup.
		Detach(ctx context.Context, groupID string, instanceIDs []string) error

		// Instances returns a list of all instances that belong to specified InstanceGroup.
		Instances(ctx context.Context, groupID string) ([]Instance, error)
	}

	// LaunchSpecService wraps all common functionality for LaunchSpecs.
	LaunchSpecService interface {
		// List returns a list of LaunchSpecs.
		List(ctx context.Context, oceanID string) ([]LaunchSpec, error)

		// Create creates a new LaunchSpec and returns its ID.
		Create(ctx context.Context, spec LaunchSpec) (string, error)

		// Read returns an existing LaunchSpec by ID.
		Read(ctx context.Context, specID string) (LaunchSpec, error)

		// Update updates an existing LaunchSpec.
		Update(ctx context.Context, spec LaunchSpec) error

		// Delete deletes an existing LaunchSpec by ID.
		Delete(ctx context.Context, specID string) error
	}
)

type ResourceType string

const (
	ResourceTypeInstanceGroup ResourceType = "instancegroup"
	ResourceTypeLaunchSpec    ResourceType = "launchspec"
)

type InstanceGroupType string

const (
	InstanceGroupElastigroup InstanceGroupType = "elastigroup"
	InstanceGroupOcean       InstanceGroupType = "ocean"
)
