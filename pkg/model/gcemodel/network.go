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

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*GCEModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	network := &gcetasks.Network{
		Name:      s(b.NameForNetwork()),
		Lifecycle: b.Lifecycle,
		Mode:      "auto", // Automatically create subnets, but stop using legacy mode
	}
	c.AddTask(network)

	if gce.UsesIPAliases(b.Cluster) {
		if len(b.Cluster.Spec.Subnets) != 1 {
			return fmt.Errorf("expected exactly one subnet for IPAlias mode")
		}
		subnet := b.Cluster.Spec.Subnets[0]

		// The primary CIDR is used by the nodes,
		// services and pods draw from the secondary IP ranges.
		// All the CIDRs must be valid RFC1918 IP addresses, which makes conversion from the "pure kubenet" 100.64.0.0 GCE range difficult

		t := &gcetasks.Subnet{
			Name:      s(b.NameForIPAliasSubnet()),
			Network:   b.LinkToNetwork(),
			Lifecycle: b.Lifecycle,
			Region:    s(b.Region),
			CIDR:      s(subnet.CIDR),
			SecondaryIpRanges: map[string]string{
				b.NameForIPAliasRange("pods"):     b.Cluster.Spec.PodCIDR,
				b.NameForIPAliasRange("services"): b.Cluster.Spec.ServiceClusterIPRange,
			},
		}

		t.GCEName = t.Name
		c.AddTask(t)

	}

	// Create CloudNAT for private topology
	if b.Cluster.Spec.Topology.Masters == kops.TopologyPrivate {
		// All private subnets get a CloudNAT in their region.
		privateRegions := map[string]bool{}
		for i := range b.Cluster.Spec.Subnets {
			subnetSpec := &b.Cluster.Spec.Subnets[i]
			if subnetSpec.Type == kops.SubnetTypePrivate {
				klog.Infof("adding %s to list of private regions\n", subnetSpec.Region)
				privateRegions[subnetSpec.Region] = true
			}
		}

		for region, _ := range privateRegions {
			nat := &gcetasks.NatGateway{
				// Name:      s(region + "-" + gce.SafeClusterName(b.ClusterName())),
				Name:      s(b.SafeObjectName(region + "-" + b.ClusterName())),
				Lifecycle: b.Lifecycle,
				Network:   s(b.LinkToNetwork().URL(b.Cluster.Spec.Project)),
				Region:    s(region),
			}
			c.AddTask(nat)
		}
	}

	return nil
}
