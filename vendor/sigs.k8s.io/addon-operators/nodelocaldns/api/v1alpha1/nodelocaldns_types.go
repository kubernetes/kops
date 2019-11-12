package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addonv1alpha1 "sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/apis/v1alpha1"
)

// NodeLocalDNSSpec defines the desired state of NodeLocalDNS
type NodeLocalDNSSpec struct {
	addonv1alpha1.CommonSpec `json:",inline"`
	addonv1alpha1.PatchSpec  `json:",inline"`
}

// NodeLocalDNSStatus defines the observed state of NodeLocalDNS
type NodeLocalDNSStatus struct {
	addonv1alpha1.CommonStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// NodeLocalDNS is the Schema for the  API
type NodeLocalDNS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeLocalDNSSpec   `json:"spec,omitempty"`
	Status NodeLocalDNSStatus `json:"status,omitempty"`
}

var _ addonv1alpha1.CommonObject = &NodeLocalDNS{}

func (o *NodeLocalDNS) ComponentName() string {
	return "nodelocaldns"
}

func (o *NodeLocalDNS) CommonSpec() addonv1alpha1.CommonSpec {
	return o.Spec.CommonSpec
}

func (o *NodeLocalDNS) PatchSpec() addonv1alpha1.PatchSpec {
	return o.Spec.PatchSpec
}

func (o *NodeLocalDNS) GetCommonStatus() addonv1alpha1.CommonStatus {
	return o.Status.CommonStatus
}

func (o *NodeLocalDNS) SetCommonStatus(s addonv1alpha1.CommonStatus) {
	o.Status.CommonStatus = s
}

// +kubebuilder:object:root=true

// NodeLocalDNSList contains a list of NodeLocalDNS
type NodeLocalDNSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeLocalDNS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeLocalDNS{}, &NodeLocalDNSList{})
}
