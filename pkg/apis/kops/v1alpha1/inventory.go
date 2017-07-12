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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Inventory provides a data model for assets that compose a kops installation.
// This API is a top level API that is only used for Inventory CRUD. Create and
// read are implemented at this point.
type Inventory struct {
	v1.TypeMeta `json:",inline"`
	ObjectMeta  metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InventorySpec `json:"spec,omitempty"`
}

type InventorySpec struct {

	// full cluster
	Cluster *ClusterSpec `json:"cluster,omitempty"`

	// The file that contains the kops channel
	// see: https://raw.githubusercontent.com/kubernetes/kops/master/channels/stable
	ChannelAsset *ChannelAsset `json:"channel,omitempty"`

	// kops version of kops that generated the inventory.  There is an open issue
	// to track the kops version of a cluster.
	KopsVersion *string `json:"kopsVersion,omitempty"`

	KubernetesVersion *string `json:"kubernetesVersion,omitempty"`

	// List of executables including such things as nodeup, and all k8s binaries.
	ExecutableFileAsset []*ExecutableFileAsset `json:"executableFileAssets,omitempty"`

	// Compressed tar balls, for instance, cni package.
	CompressedFileAssets []*CompressedFileAsset `json:"compressedFileAssets,omitempty"`

	// Containers
	ContainerAssets []*ContainerAsset `json:"containerAssets,omitempty"`

	// All the contains all of the images from the various instance groups.
	HostAssets []*HostAsset `json:"hostAsset,omitempty"`
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
