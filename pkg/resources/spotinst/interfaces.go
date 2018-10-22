/*
Copyright 2018 The Kubernetes Authors.

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
)

type (
	// Elastigroup contains configuration info and functions to control a set
	// of instances.
	Elastigroup interface {
		// Id returns the ID of the Elastigroup.
		Id() string

		// Name returns the name of the Elastigroup.
		Name() string

		// MinSize returns the minimum size of the Elastigroup.
		MinSize() int

		// MaxSize returns the maximum size of the Elastigroup.
		MaxSize() int

		// Obj returns the raw object which is a cloud-specific implementation.
		Obj() interface{}
	}

	// Instance wraps a cloud-specific instance object.
	Instance interface {
		// Id returns the ID of the instance.
		Id() string

		// Obj returns the raw object which is a cloud-specific implementation.
		Obj() interface{}
	}

	// Service is an interface that a cloud provider that is supported
	// by Spotinst MUST implement to manage its Elastigroups.
	Service interface {
		// List returns a list of Elastigroups.
		List(ctx context.Context) ([]Elastigroup, error)

		// Create creates a new Elastigroup and returns its ID.
		Create(ctx context.Context, group Elastigroup) (string, error)

		// Read returns an existing Elastigroup by ID.
		Read(ctx context.Context, groupID string) (Elastigroup, error)

		// Update updates an existing Elastigroup.
		Update(ctx context.Context, group Elastigroup) error

		// Delete deletes an existing Elastigroup by ID.
		Delete(ctx context.Context, groupID string) error

		// Detach removes one or more instances from the specified Elastigroup.
		Detach(ctx context.Context, groupID string, instanceIDs []string) error

		// Instances returns a list of all instances that belong to specified Elastigroup.
		Instances(ctx context.Context, groupID string) ([]Instance, error)
	}
)
