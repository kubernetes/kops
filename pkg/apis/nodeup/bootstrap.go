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

package nodeup

const BootstrapAPIVersion = "bootstrap.kops.k8s.io/v1alpha1"

// BootstrapRequest is a request from nodeup to kops-controller for bootstrapping a node.
type BootstrapRequest struct {
	// APIVersion defines the versioned schema of this representation of a request.
	APIVersion string `json:"apiVersion"`
}

// BootstrapRespose is a response to a BootstrapRequest.
type BootstrapResponse struct {
}
