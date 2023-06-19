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
	// bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kops/api/v1beta1"
)

// KopsControlPlaneTemplateSpec defines the desired state of KopsControlPlaneTemplate.
type KopsControlPlaneTemplateSpec struct {
	Template KopsControlPlaneTemplateResource `json:"template"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kopscontrolplanetemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of KopsControlPlaneTemplate"

// KopsControlPlaneTemplate is the Schema for the kopscontrolplanetemplates API.
type KopsControlPlaneTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KopsControlPlaneTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// KopsControlPlaneTemplateList contains a list of KopsControlPlaneTemplate.
type KopsControlPlaneTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KopsControlPlaneTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KopsControlPlaneTemplate{}, &KopsControlPlaneTemplateList{})
}

// KopsControlPlaneTemplateResource describes the data needed to create a KopsControlPlane from a template.
type KopsControlPlaneTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`

	Spec KopsControlPlaneTemplateResourceSpec `json:"spec"`
}

// KopsControlPlaneTemplateResourceSpec defines the desired state of KopsControlPlane.
// NOTE: KopsControlPlaneTemplateResourceSpec is similar to KopsControlPlaneSpec but
// omits Replicas and Version fields. These fields do not make sense on the KopsControlPlaneTemplate,
// because they are calculated by the Cluster topology reconciler during reconciliation and thus cannot
// be configured on the KopsControlPlaneTemplate.
type KopsControlPlaneTemplateResourceSpec struct {
}

// KopsControlPlaneTemplateMachineTemplate defines the template for Machines
// in a KopsControlPlaneTemplate object.
// NOTE: KopsControlPlaneTemplateMachineTemplate is similar to KopsControlPlaneMachineTemplate but
// omits ObjectMeta and InfrastructureRef fields. These fields do not make sense on the KopsControlPlaneTemplate,
// because they are calculated by the Cluster topology reconciler during reconciliation and thus cannot
// be configured on the KopsControlPlaneTemplate.
type KopsControlPlaneTemplateMachineTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`
}
