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

package nodeup

import (
	"k8s.io/kops/util/pkg/architectures"
)

// Config is the configuration for the nodeup binary
type Config struct {
	// Tags enable/disable chunks of the model
	Tags []string `json:",omitempty"`
	// Assets are locations where we can find files to be installed
	// TODO: Remove once everything is in containers?
	Assets map[architectures.Architecture][]string `json:",omitempty"`
	// Images are a list of images we should preload
	Images []*Image `json:"images,omitempty"`
	// ConfigBase is the base VFS path for config objects
	ConfigBase *string `json:",omitempty"`
	// ClusterLocation is the VFS path to the cluster spec (deprecated: prefer ConfigBase)
	ClusterLocation *string `json:",omitempty"`
	// InstanceGroupName is the name of the instance group
	InstanceGroupName string `json:",omitempty"`
	// ClusterName is the name of the cluster
	ClusterName string `json:",omitempty"`
	// ProtokubeImage is the docker image to load for protokube (bootstrapping)
	ProtokubeImage *Image `json:"protokubeImage,omitempty"`
	// Channels is a list of channels that we should apply
	Channels []string `json:"channels,omitempty"`

	// Manifests for running etcd
	EtcdManifests []string `json:"etcdManifests,omitempty"`
}

// Image is a docker image we should pre-load
type Image struct {
	// This is the name we would pass to "docker run", whereas source could be a URL from which we would download an image.
	Name string `json:"name,omitempty"`
	// Sources is a list of URLs from which we should download the image
	Sources []string `json:"sources,omitempty"`
	// Hash is the hash of the file, to verify image integrity (even over http)
	Hash string `json:"hash,omitempty"`
}
