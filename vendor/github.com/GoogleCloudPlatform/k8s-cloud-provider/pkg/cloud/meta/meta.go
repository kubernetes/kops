/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package meta

const (
	// NoGet prevents the Get() method from being generated.
	NoGet = 1 << iota
	// NoList prevents the List() method from being generated.
	NoList = 1 << iota
	// NoDelete prevents the Delete() method from being generated.
	NoDelete = 1 << iota
	// NoInsert prevents the Insert() method from being generated.
	NoInsert = 1 << iota
	// CustomOps specifies that an empty interface xxxOps will be generated to
	// enable custom method calls to be attached to the generated service
	// interface.
	CustomOps = 1 << iota
	// AggregatedList will generated a method for AggregatedList().
	AggregatedList = 1 << iota
	// ListUsable will generate a method for ListUsable().
	ListUsable = 1 << iota

	// ReadOnly specifies that the given resource is read-only and should not
	// have insert() or delete() methods generated for the wrapper.
	ReadOnly = NoDelete | NoInsert
)

// Version of the API (ga, alpha, beta).
type Version string

const (
	// VersionGA is the GA API version.
	VersionGA Version = "ga"
	// VersionAlpha is the alpha API version.
	VersionAlpha Version = "alpha"
	// VersionBeta is the beta API version.
	VersionBeta Version = "beta"
)

// APIGroup is the API Group of the resource. When unspecified, "compute" is assumed.
type APIGroup string

const (
	// APIGroupCompute is the compute API group.
	APIGroupCompute APIGroup = "compute"

	// APIGroupNetworkServices is the networkservices API group.
	APIGroupNetworkServices APIGroup = "networkservices"
)

// AllVersions is a list of all versions of the GCP APIs.
var AllVersions = []Version{
	VersionGA,
	VersionAlpha,
	VersionBeta,
}

// AllServices are a list of all the services to generate code for. Keep
// this list in lexicographical order by object type.
var AllServices = []*ServiceInfo{}
