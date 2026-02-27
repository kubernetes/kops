/*
Copyright The Kubernetes Authors.

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
	"github.com/awslabs/operatorpkg/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LambdaAINodeClassSpec defines the desired state of LambdaAINodeClass
type LambdaAINodeClassSpec struct {
	// Image is the image to use for the instances
	// +optional
	Image string `json:"image,omitempty"`

	// Type is the instance type to use
	// +optional
	Type string `json:"type,omitempty"`
}

// LambdaAINodeClassStatus defines the observed state of LambdaAINodeClass
type LambdaAINodeClassStatus struct {
	// Conditions contains signals for health and readiness
	// +optional
	Conditions []status.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=lambdaainodeclasses,scope=Cluster,categories=karpenter

// LambdaAINodeClass is the Schema for the lambdaainodeclasses API
type LambdaAINodeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LambdaAINodeClassSpec   `json:"spec,omitempty"`
	Status LambdaAINodeClassStatus `json:"status,omitempty"`
}

func (c *LambdaAINodeClass) StatusConditions() status.ConditionSet {
	return status.NewReadyConditions().For(c)
}

func (c *LambdaAINodeClass) GetConditions() []status.Condition {
	return c.Status.Conditions
}

func (c *LambdaAINodeClass) SetConditions(conditions []status.Condition) {
	c.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// LambdaAINodeClassList contains a list of LambdaAINodeClass
type LambdaAINodeClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LambdaAINodeClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LambdaAINodeClass{}, &LambdaAINodeClassList{})
}
