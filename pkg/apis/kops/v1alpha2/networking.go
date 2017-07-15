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

// NetworkingSpec allows selection and configuration of a networking plugin
type NetworkingSpec struct {
	Classic    *ClassicNetworkingSpec    `json:"classic,omitempty"`
	Kubenet    *KubenetNetworkingSpec    `json:"kubenet,omitempty"`
	External   *ExternalNetworkingSpec   `json:"external,omitempty"`
	CNI        *CNINetworkingSpec        `json:"cni,omitempty"`
	Kopeio     *KopeioNetworkingSpec     `json:"kopeio,omitempty"`
	Weave      *WeaveNetworkingSpec      `json:"weave,omitempty"`
	Flannel    *FlannelNetworkingSpec    `json:"flannel,omitempty"`
	Calico     *CalicoNetworkingSpec     `json:"calico,omitempty"`
	Canal      *CanalNetworkingSpec      `json:"canal,omitempty"`
	Kuberouter *KuberouterNetworkingSpec `json:"kuberouter,omitempty"`
}

// ClassicNetworkingSpec is the specification of classic networking mode, integrated into kubernetes
type ClassicNetworkingSpec struct {
}

// KubenetNetworkingSpec is the specification for kubenet networking, largely integrated but intended to replace classic
type KubenetNetworkingSpec struct {
}

// ExternalNetworkingSpec is the specification for networking that is implemented by a Daemonset
// It also uses kubenet
type ExternalNetworkingSpec struct {
}

// CNI is the specification for networking that is implemented by a Daemonset
// Networking is not managed by kops - we can create options here that directly configure e.g. weave
// but this is useful for arbitrary network modes or for modes that don't need additional configuration.
type CNINetworkingSpec struct {
}

// Kopeio declares that we want Kopeio networking
type KopeioNetworkingSpec struct {
}

// Weave declares that we want Weave networking
type WeaveNetworkingSpec struct {
	MTU *int32 `json:"mtu,omitempty"`
}

// Flannel declares that we want Flannel networking
type FlannelNetworkingSpec struct {
}

// Calico declares that we want Calico networking
type CalicoNetworkingSpec struct {
	CrossSubnet bool `json:"crossSubnet,omitempty"` // Enables Calico's cross-subnet mode when set to true
}

// Canal declares that we want Canal networking
type CanalNetworkingSpec struct {
}

// Kuberouter declares that we want Canal networking
type KuberouterNetworkingSpec struct {
}
