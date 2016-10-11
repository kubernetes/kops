package kops

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
)


// Federation represents a federated set of kubernetes clusters
type Federation struct {
	unversioned.TypeMeta `json:",inline"`
	ObjectMeta    `json:"metadata,omitempty"`

	Spec FederationSpec `json:"spec,omitempty"`
}

type FederationSpec struct {
	Controllers []string `json:"controllers,omitempty"`
	Members     []string `json:"members,omitempty"`

	DNSName     string `json:"dnsName,omitempty"`
}

type FederationList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []Federation `json:"items"`
}

func (f *Federation) Validate() error {
	return nil
}
