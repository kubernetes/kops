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

// KopsConfigTemplateSpec defines the desired state of KopsConfigTemplate.
type KopsConfigTemplateSpec struct {
	Template KopsConfigTemplateResource `json:"template"`
}

// KopsConfigTemplateResource defines the Template structure.
type KopsConfigTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`

	Spec KopsConfigSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kopsconfigtemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of KopsConfigTemplate"

// KopsConfigTemplate is the Schema for the kopsconfigtemplates API.
type KopsConfigTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KopsConfigTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// KopsConfigTemplateList contains a list of KopsConfigTemplate.
type KopsConfigTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KopsConfigTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KopsConfigTemplate{}, &KopsConfigTemplateList{})
}
