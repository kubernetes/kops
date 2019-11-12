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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CommonObject is an interface that must be implemented by
// all Addon objects in order to use the Addon pattern
type CommonObject interface {
	runtime.Object
	metav1.Object
	ComponentName() string
	CommonSpec() CommonSpec
	GetCommonStatus() CommonStatus
	SetCommonStatus(CommonStatus)
}

// CommonSpec defines the set of configuration attributes that must be exposed on all addons.
type CommonSpec struct {
	// Version specifies the exact addon version to be deployed, eg 1.2.3
	// It should not be specified if Channel is specified
	Version string `json:"version,omitempty"`
	// Channel specifies a channel that can be used to resolve a specific addon, eg: stable
	// It will be ignored if Version is specified
	Channel string `json:"channel,omitempty"`
}

//go:generate go run ../../../../../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./... -h ../../../../../../hack/boilerplate.go.txt
// +k8s:deepcopy-gen=true

// CommonSpec is a set of status attributes that must be exposed on all addons.
type CommonStatus struct {
	Healthy bool     `json:"healthy"`
	Errors  []string `json:"errors,omitempty"`
}

// Patchable is a trait for addon CRDs that expose a raw set of Patches to be
// applied to the declarative manifest.
type Patchable interface {
	PatchSpec() PatchSpec
}

// +k8s:deepcopy-gen=true
type PatchSpec struct {
	Patches []*runtime.RawExtension `json:"patches,omitempty"`
}
