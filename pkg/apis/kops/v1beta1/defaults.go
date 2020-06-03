/*
Copyright 2020 The Kubernetes Authors.

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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_ClusterSpec(obj *ClusterSpec) {
	if obj.Topology == nil {
		obj.Topology = &TopologySpec{}
	}

	if obj.Topology.Masters == "" {
		obj.Topology.Masters = TopologyPublic
	}

	if obj.Topology.Nodes == "" {
		obj.Topology.Nodes = TopologyPublic
	}

	if obj.Topology.DNS == nil {
		obj.Topology.DNS = &DNSSpec{}
	}

	if obj.Topology.DNS.Type == "" {
		obj.Topology.DNS.Type = DNSTypePublic
	}

	if obj.API == nil {
		obj.API = &AccessSpec{}
	}

	if obj.API.IsEmpty() {
		switch obj.Topology.Masters {
		case TopologyPublic:
			obj.API.DNS = &DNSAccessSpec{}

		case TopologyPrivate:
			obj.API.LoadBalancer = &LoadBalancerAccessSpec{}

		default:
			klog.Infof("unknown master topology type: %q", obj.Topology.Masters)
		}
	}

	if obj.API.LoadBalancer != nil && obj.API.LoadBalancer.Type == "" {
		obj.API.LoadBalancer.Type = LoadBalancerTypePublic
	}

	if obj.Authorization == nil {
		obj.Authorization = &AuthorizationSpec{}
	}
	if obj.Authorization.IsEmpty() {
		// Before the Authorization field was introduced, the behaviour was alwaysAllow
		obj.Authorization.AlwaysAllow = &AlwaysAllowAuthorizationSpec{}
	}

	if obj.IAM == nil {
		obj.IAM = &IAMSpec{
			Legacy: true,
		}
	}

	if obj.Networking != nil {
		if obj.Networking.Flannel != nil {
			if obj.Networking.Flannel.Backend == "" {
				// Populate with legacy default value; new clusters will be created with vxlan by create cluster
				obj.Networking.Flannel.Backend = "udp"
			}
		}
	}
}
