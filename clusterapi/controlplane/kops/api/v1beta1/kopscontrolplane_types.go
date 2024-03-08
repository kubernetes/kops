/*
Copyright 2023 The Kubernetes Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "k8s.io/kops/clusterapi/snapshot/cluster-api/api/v1beta1"
)

// RolloutStrategyType defines the rollout strategies for a KopsControlPlane.
type RolloutStrategyType string

// KopsControlPlaneSpec defines the desired state of KopsControlPlane.
type KopsControlPlaneSpec struct {
}

// KopsControlPlaneMachineTemplate defines the template for Machines
// in a KopsControlPlane object.
type KopsControlPlaneMachineTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`
}

// KopsControlPlaneStatus defines the observed state of KopsControlPlane.
type KopsControlPlaneStatus struct {
	// initialization provides observations of the KopsControlPlane initialization process.
	// NOTE: Fields in this struct are part of the Cluster API contract and are used to orchestrate initial Machine provisioning.
	// +optional
	Initialization KopsControlPlaneInitializationStatus `json:"initialization,omitempty,omitzero"`

	// KopsControllerEndpoint represents the endpoints used to communicate with the control plane.
	SystemEndpoints []SystemEndpoint `json:"systemEndpoints,omitempty"`
}

// KopsControlPlaneInitializationStatus provides observations of the KopsControlPlane initialization process.
// +kubebuilder:validation:MinProperties=1
type KopsControlPlaneInitializationStatus struct {
	// controlPlaneInitialized is true when the KopsControlPlane provider reports that the Kubernetes control plane is initialized;
	// A control plane is considered initialized when it can accept requests, no matter if this happens before
	// the control plane is fully provisioned or not.
	// NOTE: this field is part of the Cluster API contract, and it is used to orchestrate initial Machine provisioning.
	// +optional
	ControlPlaneInitialized *bool `json:"controlPlaneInitialized,omitempty"`
}

// SystemEndpointType identifies the service that the SystemEndpoint is describing.
type SystemEndpointType string

const (
	// SystemEndpointTypeKubeAPIServer indicates that the endpoint is for the Kubernetes API server.
	SystemEndpointTypeKubeAPIServer SystemEndpointType = "kube-apiserver"
	// SystemEndpointTypeKopsController indicates that the endpoint is for the kops-controller.
	SystemEndpointTypeKopsController SystemEndpointType = "kops-controller"
)

// SystemEndpointScope describes whether an endpoint is intended for internal or external use.
type SystemEndpointScope string

const (
	// SystemEndpointScopeInternal indicates that the endpoint is intended for internal use.
	SystemEndpointScopeInternal SystemEndpointScope = "internal"
	// SystemEndpointScopeExternal indicates that the endpoint is intended for external use.
	SystemEndpointScopeExternal SystemEndpointScope = "external"
)

// SystemEndpoint represents a reachable Kubernetes API endpoint.
type SystemEndpoint struct {
	// The type of the endpoint
	Type SystemEndpointType `json:"type"`

	// The hostname or IP on which the API server is serving.
	Endpoint string `json:"endpoint"`

	// Whether the endpoint is intended for internal or external use.
	Scope SystemEndpointScope `json:"scope"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kopscontrolplanes,shortName=kcp,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:metadata:labels=cluster.x-k8s.io/v1beta2=v1beta1
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels['cluster\\.x-k8s\\.io/cluster-name']",description="Cluster"
// +kubebuilder:printcolumn:name="Initialized",type=boolean,JSONPath=".status.initialized",description="This denotes whether or not the control plane has the uploaded kops-config configmap"
// +kubebuilder:printcolumn:name="API Server Available",type=boolean,JSONPath=".status.ready",description="KopsControlPlane API Server is ready to receive requests"
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=".spec.replicas",description="Total number of machines desired by this control plane",priority=10
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=".status.replicas",description="Total number of non-terminated machines targeted by this control plane"
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=".status.readyReplicas",description="Total number of fully running and ready control plane machines"
// +kubebuilder:printcolumn:name="Updated",type=integer,JSONPath=".status.updatedReplicas",description="Total number of non-terminated machines targeted by this control plane that have the desired template spec"
// +kubebuilder:printcolumn:name="Unavailable",type=integer,JSONPath=".status.unavailableReplicas",description="Total number of unavailable machines targeted by this control plane"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of KopsControlPlane"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=".spec.version",description="Kubernetes version associated with this control plane"

// KopsControlPlane is the Schema for the KopsControlPlane API.
type KopsControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KopsControlPlaneSpec   `json:"spec,omitempty"`
	Status KopsControlPlaneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KopsControlPlaneList contains a list of KopsControlPlane.
type KopsControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KopsControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KopsControlPlane{}, &KopsControlPlaneList{})
}
