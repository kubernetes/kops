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

	"github.com/blang/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Addons struct {
	metav1.TypeMeta `json:",inline"`

	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

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

	// Manifest is the URL to the manifest that should be applied
	Manifest *string `json:"manifest,omitempty"`

	// Manifesthash is the sha1 hash of our manifest
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

	// ReplaceBeforeVersion is the version before which upgrades should replace instead of apply the
	// manifest. For example, we made a change on an immutable field in version "1.1.0-kops.2" of an
	// addon spec which cannot be successfully updated with kubectl apply. Setting ReplaceBeforeVersion
	// to "1.1.0-kops.2" makes sure when updating from any version below "1.1.0-kops.2" the channel should
	// update the addon using kubectl replace instead of kubectl apply.
	ReplaceBeforeVersion *string `json:"replaceBeforeVersion,omitempty"`
}

func (a *Addons) Verify() error {
	for _, addon := range a.Spec.Addons {
		if addon != nil && addon.Version != nil && *addon.Version != "" {
			name := a.ObjectMeta.Name
			if addon.Name != nil {
				name = *addon.Name
			}

			_, err := semver.ParseTolerant(*addon.Version)
			if err != nil {
				return fmt.Errorf("addon %q has unparseable version %q: %v", name, *addon.Version, err)
			}
		}
	}

	return nil
}
