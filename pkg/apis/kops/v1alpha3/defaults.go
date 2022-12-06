/*
Copyright 2021 The Kubernetes Authors.

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

package v1alpha3

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_ClusterSpec(obj *ClusterSpec) {
	rebindIfEmpty := func(s *string, replacement string) bool {
		if *s != "" {
			return false
		}
		*s = replacement
		return true
	}

	if obj.Networking.Topology == nil {
		obj.Networking.Topology = &TopologySpec{}
	}

	rebindIfEmpty(&obj.Networking.Topology.ControlPlane, TopologyPublic)

	rebindIfEmpty(&obj.Networking.Topology.Nodes, TopologyPublic)

	if obj.Networking.Topology.DNS == "" {
		obj.Networking.Topology.DNS = DNSTypePublic
	}

	if obj.CloudProvider.Openstack == nil {
		if obj.API.DNS == nil && obj.API.LoadBalancer == nil {
			switch obj.Networking.Topology.ControlPlane {
			case TopologyPublic:
				obj.API.DNS = &DNSAccessSpec{}

			case TopologyPrivate:
				obj.API.LoadBalancer = &LoadBalancerAccessSpec{}

			default:
				klog.Infof("unknown controlPlane topology type: %q", obj.Networking.Topology.ControlPlane)
			}
		}

		if obj.API.LoadBalancer != nil && obj.API.LoadBalancer.Type == "" {
			obj.API.LoadBalancer.Type = LoadBalancerTypePublic
		}

	}

	if obj.API.LoadBalancer != nil && obj.API.LoadBalancer.Class == "" && obj.CloudProvider.AWS != nil {
		obj.API.LoadBalancer.Class = LoadBalancerClassClassic
	}

	if obj.Authorization == nil {
		obj.Authorization = &AuthorizationSpec{}
	}
	if obj.Authorization.IsEmpty() {
		// Before the Authorization field was introduced, the behaviour was alwaysAllow
		obj.Authorization.AlwaysAllow = &AlwaysAllowAuthorizationSpec{}
	}

	if obj.Networking.Flannel != nil {
		// Populate with legacy default value; new clusters will be created with "vxlan" by
		// "create cluster."
		rebindIfEmpty(&obj.Networking.Flannel.Backend, "udp")
	}
}
