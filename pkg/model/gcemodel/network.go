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

package gcemodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*GCEModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	sharedNetwork := b.Cluster.Spec.NetworkID != ""

	network := &gcetasks.Network{
		Name:      b.LinkToNetwork().Name,
		Lifecycle: b.Lifecycle,
		Shared:    fi.Bool(sharedNetwork),
	}
	if !sharedNetwork {
		// As we're creating the network, we're also creating the subnets.
		// We therefore use custom mode, for a few reasons:
		// 1) We aren't going to use the auto-allocated subnets anyway
		// 2) The GCE docs recommend that production usage plan CIDR allocation by using custom mode
		network.Mode = "custom"
	}
	c.AddTask(network)

	for i := range b.Cluster.Spec.Subnets {
		subnet := &b.Cluster.Spec.Subnets[i]

		sharedSubnet := subnet.ProviderID != ""

		t := &gcetasks.Subnet{
			Name:              b.LinkToSubnet(subnet).Name,
			Network:           b.LinkToNetwork(),
			Lifecycle:         b.Lifecycle,
			Region:            s(b.Region),
			Shared:            fi.Bool(sharedSubnet),
			SecondaryIpRanges: make(map[string]string),
		}

		if subnet.CIDR != "" {
			t.CIDR = s(subnet.CIDR)
		}

		t.SecondaryIpRanges = make(map[string]string)
		if gce.UsesIPAliases(b.Cluster) {
			// The primary CIDR is used by the nodes,
			// services and pods draw from the secondary IP ranges.
			// All the CIDRs must be valid RFC1918 IP addresses, which makes conversion from the "pure kubenet" 100.64.0.0 GCE range difficult

			t.CIDR = s(subnet.CIDR)
			t.SecondaryIpRanges[b.NameForIPAliasRange("pods")] = b.Cluster.Spec.PodCIDR
			t.SecondaryIpRanges[b.NameForIPAliasRange("services")] = b.Cluster.Spec.ServiceClusterIPRange
		}

		c.AddTask(t)
	}

	// Create a CloudNAT for private topology.
	if b.Cluster.Spec.Topology.Masters == kops.TopologyPrivate {
		var hasPrivateSubnet bool
		for _, subnet := range b.Cluster.Spec.Subnets {
			if subnet.Type == kops.SubnetTypePrivate {
				hasPrivateSubnet = true
				break
			}
		}

		if hasPrivateSubnet {
			r := &gcetasks.Router{
				Name:                          s(b.SafeObjectName("nat")),
				Lifecycle:                     b.Lifecycle,
				Network:                       b.LinkToNetwork(),
				Region:                        s(b.Region),
				NATIPAllocationOption:         s(gcetasks.NATIPAllocationOptionAutoOnly),
				SourceSubnetworkIPRangesToNAT: s(gcetasks.SourceSubnetworkIPRangesAll),
			}
			c.AddTask(r)
		}
	}

	return nil
}
