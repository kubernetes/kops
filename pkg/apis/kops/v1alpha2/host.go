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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:metadata:annotations="api-approved.kubernetes.io=https://github.com/kubernetes/enhancements/pull/1111"
// +kubebuilder:subresource:status

// Host represents a bare-metal machine that could be registered as a Node.
type Host struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HostSpec `json:"spec,omitempty"`
}

type HostSpec struct {
	PublicKey     string `json:"publicKey,omitempty"`
	InstanceGroup string `json:"instanceGroup,omitempty"`

	// PodCIDRs configures the IP ranges to be used for pods on this node/host.
	PodCIDRs []string `json:"podCIDRs,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Host `json:"items"`
}
