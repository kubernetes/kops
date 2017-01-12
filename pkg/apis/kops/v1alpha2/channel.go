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
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type Channel struct {
	unversioned.TypeMeta `json:",inline"`
	ObjectMeta           api.ObjectMeta `json:"metadata,omitempty"`

	Spec ChannelSpec `json:"spec,omitempty"`
}

type ChannelSpec struct {
	Images []*ChannelImageSpec `json:"images,omitempty"`

	Cluster *ClusterSpec `json:"cluster,omitempty"`
}

type ChannelImageSpec struct {
	Labels map[string]string `json:"labels,omitempty"`

	ProviderID string `json:"providerID,omitempty"`

	Name string `json:"name,omitempty"`
}
