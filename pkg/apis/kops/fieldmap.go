/*
Copyright 2024 The Kubernetes Authors.

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

package kops

// Because we have multiple versions of the API, fields have different paths in different API versions.
// This file contains the mappings from the old paths to the new paths.

// These functions are somewhat of a work in progress, we are adding function mappings as we hit them.

// HumanPathForClusterField returns the path for a given v1alpha3 cluster field, as we should print it for users.
func HumanPathForClusterField(fieldPath string) string {
	f := NewClusterField(fieldPath)
	return f.HumanPath()
}

// InternalPathForClusterField returns the path for a given cluster field, as it appears in the internal API.
func InternalPathForClusterField(fieldPath string) string {
	f := NewClusterField(fieldPath)
	return f.InternalPath()
}

// ClusterField represents a field in the Cluster resource.
// +k8s:deepcopy-gen=false
type ClusterField struct {
	Path string
}

// NewClusterField creates a new ClusterField for the given path.
func NewClusterField(path string) *ClusterField {
	return &ClusterField{Path: path}
}

// HumanPath returns the path for the field, as we should print it for users.
func (f *ClusterField) HumanPath() string {
	return f.PathInV1Alpha2()
}

// InternalPath returns the path for the field, as it appears in the internal API.
func (f *ClusterField) InternalPath() string {
	return f.PathInV1Alpha3()
}

// PathInV1Alpha2 returns the path for the field in the v1alpha2 API.
func (f *ClusterField) PathInV1Alpha2() string {
	for _, mapping := range clusterFieldMappings {
		if mapping.V1Alpha3 == f.Path {
			return mapping.V1Alpha2
		}
	}
	return f.Path
}

// PathInV1Alpha3 returns the path for the field in the v1alpha3 API.
func (f *ClusterField) PathInV1Alpha3() string {
	for _, mapping := range clusterFieldMappings {
		if mapping.V1Alpha2 == f.Path {
			return mapping.V1Alpha3
		}
	}
	return f.Path
}

// clusterFieldMappings is a list of mappings from v1alpha2 field paths to v1alpha3 field paths.
var clusterFieldMappings = []struct {
	V1Alpha2 string
	V1Alpha3 string
}{
	{V1Alpha2: "spec.masterPublicName", V1Alpha3: "spec.api.publicName"},
}
