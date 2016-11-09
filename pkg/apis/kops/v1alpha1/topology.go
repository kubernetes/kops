/*
Copyright 2016 The Kubernetes Authors.

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

package v1alpha1

const (
	TopologyPublic         = "public"
	TopologyPrivate        = "private"
)

type TopologySpec struct {
	// The environment to launch the Kubernetes masters in public|private
	Masters       string `json:"masters,omitempty"`

	// The environment to launch the Kubernetes nodes in public|private
	Nodes         string `json:"nodes,omitempty"`

	// Controls if a private topology should deploy a bastion host or not
	// The bastion host is designed to be a simple, and secure bridge between
	// the public subnet and the private subnet
	BypassBastion bool `json:"bypassBastion,omitempty"`
}