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
)

// KopsConfigSpec defines the desired state of KopsConfig.
// Either ClusterConfiguration and InitConfiguration should be defined or the JoinConfiguration should be defined.
type KopsConfigSpec struct {
}

// KopsConfigStatus defines the observed state of KopsConfig.
type KopsConfigStatus struct {
	// Ready indicates the BootstrapData field is ready to be consumed
	// +optional
	Ready bool `json:"ready"`

	// DataSecretName is the name of the secret that stores the bootstrap data script.
	// +optional
	DataSecretName *string `json:"dataSecretName,omitempty"`

	// // FailureReason will be set on non-retryable errors
	// // +optional
	// FailureReason string `json:"failureReason,omitempty"`

	// // FailureMessage will be set on non-retryable errors
	// // +optional
	// FailureMessage string `json:"failureMessage,omitempty"`

	// // ObservedGeneration is the latest generation observed by the controller.
	// // +optional
	// ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// // Conditions defines current service state of the KopsConfig.
	// // +optional
	// Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kopsconfigs,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels['cluster\\.x-k8s\\.io/cluster-name']",description="Cluster"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of KopsConfig"

// KopsConfig is the Schema for the kopsconfigs API.
type KopsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KopsConfigSpec   `json:"spec,omitempty"`
	Status KopsConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KopsConfigList contains a list of KopsConfig.
type KopsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KopsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KopsConfig{}, &KopsConfigList{})
}
