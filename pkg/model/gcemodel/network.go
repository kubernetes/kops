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
	"fmt"

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

	network, err := b.LinkToNetwork()
	if err != nil {
		return nil
	}
	network.Lifecycle = b.Lifecycle
	network.Shared = fi.Bool(sharedNetwork)
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

		network, err := b.LinkToNetwork()
		if err != nil {
			return nil
		}
		t := &gcetasks.Subnet{
			Name:              b.LinkToSubnet(subnet).Name,
			Network:           network,
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
	{
		// We only consider private subnets.
		// Then if we are creating subnet, we will create a NAT gateway tied to those subnets.
		// This can be over-ridden by specifying "external", in which case we will not create a NAT gateway.
		// If we are reusing an existing subnet, we assume that the NAT gateway is already configured.

		var subnetworks []*gcetasks.Subnet

		for i := range b.Cluster.Spec.Subnets {
			subnet := &b.Cluster.Spec.Subnets[i]
			// Only need to deal with private subnets
			if subnet.Type != kops.SubnetTypeDualStack && subnet.Type != kops.SubnetTypePrivate {
				continue
			}

			// If we're in an existing subnet, we assume egress is already configured.
			if subnet.ProviderID != "" {
				continue
			}

			switch subnet.Egress {
			case kops.EgressExternal:
				// User has request we ignore this
				continue

			case kops.EgressNatGateway, "":
				// OK, should create
				subnetworks = append(subnetworks, b.LinkToSubnet(subnet))

			default:
				return fmt.Errorf("egress mode %q is not supported", subnet.Egress)
			}
		}

		if len(subnetworks) != 0 {

			network, err := b.LinkToNetwork()
			if err != nil {
				return nil
			}
			r := &gcetasks.Router{
				Name:                          s(b.NameForRouter("nat")),
				Lifecycle:                     b.Lifecycle,
				Network:                       network,
				Region:                        s(b.Region),
				NATIPAllocationOption:         s(gcetasks.NATIPAllocationOptionAutoOnly),
				SourceSubnetworkIPRangesToNAT: s(gcetasks.SourceSubnetworkIPRangesSpecificSubnets),
				Subnetworks:                   subnetworks,
			}
			c.AddTask(r)
		}
	}

	return nil
}
