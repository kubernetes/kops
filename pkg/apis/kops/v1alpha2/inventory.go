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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Inventory struct {
	v1.TypeMeta `json:",inline"`
	ObjectMeta  metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InventorySpec `json:"spec,omitempty"`
}

type InventorySpec struct {
	Cluster              *ClusterSpec           `json:"cluster,omitempty"`
	ChannelAsset         *ChannelAsset          `json:"channel,omitempty"`
	KopsVersion          *string                `json:"kopsVersion,omitempty"`
	KubernetesVersion    *string                `json:"kubernetesVersion,omitempty"`
	ExecutableFileAsset  []*ExecutableFileAsset `json:"executableFileAssets,omitempty"`
	CompressedFileAssets []*CompressedFileAsset `json:"compressedFileAssets,omitempty"`
	ContainerAssets      []*ContainerAsset      `json:"containerAssets,omitempty"`
	HostAssets           []*HostAsset           `json:"hostAsset,omitempty"`
}

type ChannelAsset struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
	Version  string `json:"version,omitempty"`
}

type HostAsset struct {
	Name          string `json:"name"`
	Cloud         string `json:"cloud"`
	Role          string `json:"role"`
	InstanceGroup string `json:"instanceGroup"`
}

type ExecutableFileAsset struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
	Version  string `json:"version,omitempty"`
	SHA      string `json:"sha,omitempty"`
}

type CompressedFileAsset struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
	Version  string `json:"version,omitempty"`
	SHA      string `json:"sha,omitempty"`
}

type ContainerAsset struct {
	String string `json:"string,omitempty"`
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Digest string `json:"digest,omitempty"`

	Location string `json:"location,omitempty"`
	SHA      string `json:"sha,omitempty"`
	Hash     string `json:"hash,omitempty"`
}
