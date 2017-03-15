/*
Copyright 2016 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/kubernetes/pkg/api/v1"
)

type Addons struct {
	metav1.TypeMeta `json:",inline"`

	// We use v1.ObjectMeta so we don't serialize everything, even though we are using the full api machinery here (yet!)
	ObjectMeta v1.ObjectMeta `json:"metadata,omitempty"`

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
