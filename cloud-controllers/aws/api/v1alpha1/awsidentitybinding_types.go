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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSIdentityBindingSpec defines the desired state of AWSIdentityBinding
type AWSIdentityBindingSpec struct {
	// Subject describes the kubernetes service account.
	Subject SubjectName `json:"subject,omitempty"`

	// InlinePolicy allows for direct specification of an IAM policy.
	InlinePolicy string `json:"inlinePolicy,omitempty"`

	// IAMPolicyARNs specifies the policies that should be attached.
	IAMPolicyARNs []string `json:"iamPolicyARNs,omitempty"`
}

// SubjectName identifies a kubernetes serviceaccount.
type SubjectName struct {
	// Name is the name of the kubernetes ServiceAccount.
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the kubernetes ServiceAccount.
	Namespace string `json:"namespace,omitempty"`
}

// AWSIdentityBindingStatus defines the observed state of AWSIdentityBinding
type AWSIdentityBindingStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSIdentityBinding is the Schema for the awsidentitybindings API
type AWSIdentityBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSIdentityBindingSpec   `json:"spec,omitempty"`
	Status AWSIdentityBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSIdentityBindingList contains a list of AWSIdentityBinding
type AWSIdentityBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSIdentityBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSIdentityBinding{}, &AWSIdentityBindingList{})
}
