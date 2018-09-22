/*
Copyright 2017 The Kubernetes Authors.

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

package resources

// Instance is the type for an instance in a dump
type Instance struct {
	Name            string   `json:"name,omitempty"`
	PublicAddresses []string `json:"publicAddresses,omitempty"`
	Roles           []string `json:"roles,omitempty"`
	SSHUser         string   `json:"sshUser,omitempty"`
}

// Subnet is the type for an subnetwork in a dump
type Subnet struct {
	ID   string `json:"id,omitempty"`
	Zone string `json:"zone,omitempty"`
}

// VPC is the type for an VPC in a dump
type VPC struct {
	ID string `json:"id,omitempty"`
}

// Dump is the type for a dump result
type Dump struct {
	Resources []interface{} `json:"resources,omitempty"`
	Instances []*Instance   `json:"instances,omitempty"`
	Subnets   []*Subnet     `json:"subnets,omitempty"`
	VPC       *VPC          `json:"vpc,omitempty"`
}
