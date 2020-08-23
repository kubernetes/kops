/*
Copyright 2020 The Kubernetes Authors.

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

import v1 "k8s.io/api/core/v1"

// CloudInstanceStatusDetached means the instance needs update and has been detached.
const CloudInstanceStatusDetached = "Detached"

// CloudInstanceStatusNeedsUpdate means the instance has joined the cluster, is not detached, and needs to be updated.
const CloudInstanceStatusNeedsUpdate = "NeedsUpdate"

// CloudInstanceStatusReady means the instance has joined the cluster, is not detached, and is up to date.
const CloudInstanceStatusUpToDate = "UpToDate"

// CloudInstance describes an instance in a CloudInstanceGroup group.
type CloudInstance struct {
	// ID is a unique identifier for the instance, meaningful to the cloud
	ID string
	// Node is the associated k8s instance, if it is known
	Node *v1.Node
	// CloudInstanceGroup is the managing CloudInstanceGroup
	CloudInstanceGroup *CloudInstanceGroup
	// Status indicates if the instance has joined the cluster and if it needs any updates.
	Status string
	// Roles are the roles the instance have.
	Roles []string
	// MachineType is the hardware resource class of the instance.
	MachineType string
	// Private IP is the private ip address of the instance.
	PrivateIP string
}
