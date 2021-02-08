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
	// Certs are the requested certificates and their respective public keys.
	Certs map[string]string `json:"certs"`

	// IncludeNodeConfig controls whether the cluster & instance group configuration should be returned.
	// This allows for nodes without access to the kops state store.
	IncludeNodeConfig bool `json:"includeNodeConfig"`
}

// BootstrapResponse is a response to a BootstrapRequest.
type BootstrapResponse struct {
	// Certs are the issued certificates.
	Certs map[string]string

	// NodeConfig contains the node configuration, if IncludeNodeConfig is set.
	NodeConfig *NodeConfig `json:"nodeConfig,omitempty"`
}

// NodeConfig holds configuration needed to boot a node (without the kops state store)
type NodeConfig struct {
	// InstanceGroupConfig holds the configuration for the node's instance group
	InstanceGroupConfig string `json:"instanceGroupConfig,omitempty"`

	// ClusterFullConfig holds the configuration for the cluster
	ClusterFullConfig string `json:"clusterFullConfig,omitempty"`

	// Certificates holds certificates that are already issued
	Certificates []*NodeConfigCertificate `json:"certificates,omitempty"`
}

// NodeConfigCertificate holds a certificate that the node needs to boot.
type NodeConfigCertificate struct {
	// Name identifies the certificate.
	Name string `json:"name,omitempty"`

	// Cert is the certificate data.
	Cert string `json:"cert,omitempty"`
}
