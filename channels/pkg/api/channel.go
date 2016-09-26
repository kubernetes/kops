package api

import (
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type Addons struct {
	unversioned.TypeMeta `json:",inline"`
	k8sapi.ObjectMeta    `json:"metadata,omitempty"`

	Spec AddonsSpec `json:"spec,omitempty"`
}

type AddonsSpec struct {
	Addons []*AddonSpec `json:"addons,omitempty"`
}

type AddonSpec struct {
	Name *string `json:"name,omitempty"`

	Namespace *string `json:"namespace,omitempty"`

	// Selector is a label query over pods that should match the Replicas count.
	Selector map[string]string `json:"selector"`

	// Version is a semver version
	Version *string `json:"version,omitempty"`

	// Manifest is a strings containing the URL to the manifest that should be applied
	Manifest *string `json:"manifest,omitempty"`
}
