/*
Copyright 2019 The Kubernetes Authors.

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

package api

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/values"
)

type Addons struct {
	metav1.TypeMeta `json:",inline"`

	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AddonsSpec `json:"spec,omitempty"`
}

type AddonsSpec struct {
	Addons []*AddonSpec `json:"addons,omitempty"`
}

type NeedsRollingUpdate string

const (
	NeedsRollingUpdateControlPlane NeedsRollingUpdate = "control-plane"
	NeedsRollingUpdateWorkers      NeedsRollingUpdate = "workers"
	NeedsRollingUpdateAll          NeedsRollingUpdate = "all"
)

type AddonSpec struct {
	Name *string `json:"name,omitempty"`

	Namespace *string `json:"namespace,omitempty"`

	// Selector is a label query over pods that should match the Replicas count.
	Selector map[string]string `json:"selector"`

	// Manifest is the URL to the manifest that should be applied
	Manifest *string `json:"manifest,omitempty"`

	// ManifestHash is the sha256 hash of our manifest
	ManifestHash string `json:"manifestHash,omitempty"`

	// KubernetesVersion is a semver version range on which this version of the addon can be applied
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// Id is an optional value which can be used to force a refresh even if the Version matches
	// This is useful for when we have two manifests expressing the same addon version for two
	// different kubernetes api versions.  For example, we might label the 1.5 version "k8s-1.5"
	// and the 1.6 version "k8s-1.6".  Both would have the same Version, determined by the
	// version of the software we are packaging.  But we always want to reinstall when we
	// switch kubernetes versions.
	Id string `json:"id,omitempty"`

	// NeedsRollingUpdate determines if we should mark nodes as needing an update.
	// Legal values are control-plane, workers, and all
	// Empty value means no update needed
	NeedsRollingUpdate NeedsRollingUpdate `json:"needsRollingUpdate,omitempty"`

	// NeedsPKI determines if channels should provision a CA and a cert-manager issuer for the addon.
	NeedsPKI bool `json:"needsPKI,omitempty"`

	Version string `json:"version,omitempty"`

	// PruneSpec specifies how old objects should be removed (pruned).
	Prune *PruneSpec `json:"prune,omitempty"`
}

// PruneSpec specifies how old objects should be removed (pruned).
type PruneSpec struct {
	// Kinds specifies the objects to be pruned, by Kind.
	Kinds []PruneKindSpec `json:"kinds,omitempty"`
}

// PruneKindSpec specifies pruning for a particular Kind of object.
type PruneKindSpec struct {
	// Group specifies the object Group to be pruned (required).
	Group string `json:"group,omitempty"`
	// Kind specifies the object Kind to be pruned (required).
	Kind string `json:"kind,omitempty"`

	// Namespaces limits pruning only to objects in certain namespaces.
	Namespaces []string `json:"namespaces,omitempty"`

	// LabelSelector limits pruning only to objects matching the specified labels.
	LabelSelector string `json:"labelSelector,omitempty"`

	// FieldSelector allows pruning only of objects matching the field selector.
	// (This isn't currently used, but adding it now lets us start without worrying about version skew)
	FieldSelector string `json:"fieldSelector,omitempty"`
}

func (a *Addons) Verify() error {
	for _, addon := range a.Spec.Addons {
		if addon == nil {
			continue
		}
		if addon.KubernetesVersion != "" {
			return fmt.Errorf("bootstrap addon %q has a KubernetesVersion", values.StringValue(addon.Name))
		}
	}

	return nil
}
