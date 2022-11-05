/*
Copyright 2021 The Kubernetes Authors.

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

package gcetpm

// TPMVerifierOptions describes how we authenticate instances with GCP TPM authentication.
type TPMVerifierOptions struct {
	// ProjectID is the GCP project we require
	ProjectID string `json:"projectID,omitempty"`

	// Region is the region we require instances to be in.
	Region string `json:"region,omitempty"`

	// ClusterName is the cluster-name tag we require
	ClusterName string `json:"clusterName,omitempty"`

	// MaxTimeSkew is the maximum time skew to allow (in seconds)
	MaxTimeSkew int64 `json:"MaxTimeSkew,omitempty"`
}
