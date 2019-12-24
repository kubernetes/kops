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

package wellknownports

const (
	// EtcdMainQuarantinedClientPort is the port used by etcd when quarantined, for the main etcd
	EtcdMainQuarantinedClientPort = 3994

	// EtcdEventsQuarantinedClientPort is the port used by etcd when quarantined, for the events etcd
	EtcdEventsQuarantinedClientPort = 3995

	// EtcdMainGRPC is the GRPC port used by etcd-manager, for the main etcd
	EtcdMainGRPC = 3996

	// EtcdEventsGRPC is the GRPC port used by etcd-manager, for the events etcd
	EtcdEventsGRPC = 3997

	// DNSControllerGossipWeaveMesh is the port where dns-controller listens for the weave-mesh backend gossip
	DNSControllerGossipWeaveMesh = 3998

	// ProtokubeGossipWeaveMesh is the port where protokube listens for the weave-mesh-backed gossip
	ProtokubeGossipWeaveMesh = 3999

	// ProtokubeGossipMemberlist is the port where protokube listens for the memberlist-backed gossip
	ProtokubeGossipMemberlist = 4000

	// DNSControllerGossipMemberlist is the port where dns-controller listens for the memberlist-backed gossip
	DNSControllerGossipMemberlist = 3993

	// KopsControllerGRPCPort is the port where kops-controller listens for grpc requests from nodes, particularly for bootstrap configuration
	KopsControllerGRPCPort = 3992
)

type PortRange struct {
	Min int
	Max int
}

func DNSGossipPortRanges() []PortRange {
	return []PortRange{
		// 3993 is used by dns-controller, which is less important, so we might be able to drop it
		{Min: 3993, Max: 3993},
		{Min: 3998, Max: 4000},
	}
}

func ETCDPortRanges() []PortRange {
	return []PortRange{
		{Min: 3994, Max: 3997},
	}
}
