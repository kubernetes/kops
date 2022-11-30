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

package v1alpha2

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

	if obj.Topology == nil {
		obj.Topology = &TopologySpec{}
	}

	rebindIfEmpty(&obj.Topology.ControlPlane, TopologyPublic)

	rebindIfEmpty(&obj.Topology.Nodes, TopologyPublic)

	if obj.Topology.LegacyDNS == nil {
		obj.Topology.LegacyDNS = &DNSSpec{}
	}

	if obj.Topology.LegacyDNS.Type == "" {
		obj.Topology.LegacyDNS.Type = DNSTypePublic
	}

	if obj.LegacyCloudProvider != "openstack" {
		if obj.LegacyAPI == nil {
			obj.LegacyAPI = &APISpec{}
		}

		if obj.LegacyAPI.IsEmpty() {
			switch obj.Topology.ControlPlane {
			case TopologyPublic:
				obj.LegacyAPI.DNS = &DNSAccessSpec{}

			case TopologyPrivate:
				obj.LegacyAPI.LoadBalancer = &LoadBalancerAccessSpec{}

			default:
				klog.Infof("unknown master topology type: %q", obj.Topology.ControlPlane)
			}
		}

		if obj.LegacyAPI.LoadBalancer != nil && obj.LegacyAPI.LoadBalancer.Type == "" {
			obj.LegacyAPI.LoadBalancer.Type = LoadBalancerTypePublic
		}

		if obj.LegacyAPI.LoadBalancer != nil && obj.LegacyAPI.LoadBalancer.Class == "" && obj.LegacyCloudProvider == "aws" {
			obj.LegacyAPI.LoadBalancer.Class = LoadBalancerClassClassic
		}
	}

	if obj.Authorization == nil {
		obj.Authorization = &AuthorizationSpec{}
	}
	if obj.Authorization.IsEmpty() {
		// Before the Authorization field was introduced, the behaviour was alwaysAllow
		obj.Authorization.AlwaysAllow = &AlwaysAllowAuthorizationSpec{}
	}

	if obj.LegacyNetworking != nil {
		if obj.LegacyNetworking.Flannel != nil {
			// Populate with legacy default value; new clusters will be created with "vxlan" by
			// "create cluster."
			rebindIfEmpty(&obj.LegacyNetworking.Flannel.Backend, "udp")
		}
	}
}
