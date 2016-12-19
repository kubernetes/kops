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

package model

import (
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kubernetes/pkg/util/sets"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	sharedVPC := b.Cluster.SharedVPC()

	// VPC that holds everything for the cluster
	{
		t := &awstasks.VPC{
			Name:               s(b.ClusterName()),
			EnableDNSHostnames: fi.Bool(true),
			EnableDNSSupport:   fi.Bool(true),
			Shared:             fi.Bool(sharedVPC),
		}

		if b.Cluster.Spec.NetworkID != "" {
			t.ID = s(b.Cluster.Spec.NetworkID)
		}
		if b.Cluster.Spec.NetworkCIDR != "" {
			t.CIDR = s(b.Cluster.Spec.NetworkCIDR)
		}
		c.AddTask(t)
	}

	if !sharedVPC {
		dhcp := &awstasks.DHCPOptions{
			Name:              s(b.ClusterName()),
			DomainNameServers: s("AmazonProvidedDNS"),
		}
		if b.Region == "us-east-1" {
			dhcp.DomainName = s("ec2.internal")
		} else {
			dhcp.DomainName = s(b.Region + ".compute.internal")
		}
		c.AddTask(dhcp)

		c.AddTask(&awstasks.VPCDHCPOptionsAssociation{
			Name: s(b.ClusterName()),

			VPC:         b.LinkToVPC(),
			DHCPOptions: dhcp,
		})
	} else {
		// TODO: would be good to create these as shared, to verify them
	}

	// We always have a public route table, though for private networks it is only used for NGWs and ELBs
	var publicRouteTable *awstasks.RouteTable
	{
		// The internet gateway is the main entry point to the cluster.
		igw := &awstasks.InternetGateway{
			Name:   s(b.ClusterName()),
			VPC:    b.LinkToVPC(),
			Shared: fi.Bool(sharedVPC),
		}
		c.AddTask(igw)

		publicRouteTable = &awstasks.RouteTable{
			Name: s(b.ClusterName()),
			VPC:  b.LinkToVPC(),
		}
		c.AddTask(publicRouteTable)

		c.AddTask(&awstasks.Route{
			Name:            s("0.0.0.0/0"),
			CIDR:            s("0.0.0.0/0"),
			RouteTable:      publicRouteTable,
			InternetGateway: igw,
		})
	}

	privateZones := sets.NewString()

	for i := range b.Cluster.Spec.Subnets {
		subnetSpec := &b.Cluster.Spec.Subnets[i]
		sharedSubnet := subnetSpec.ProviderID != ""

		subnet := &awstasks.Subnet{
			Name:             s(subnetSpec.Name + "." + b.ClusterName()),
			VPC:              b.LinkToVPC(),
			AvailabilityZone: s(subnetSpec.Zone),
			CIDR:             s(subnetSpec.CIDR),
			Shared:           fi.Bool(sharedSubnet),
		}
		if subnetSpec.ProviderID != "" {
			subnet.ID = s(subnetSpec.ProviderID)
		}
		c.AddTask(subnet)

		switch subnetSpec.Type {
		case kops.SubnetTypePublic, kops.SubnetTypeUtility:
			if !sharedSubnet {
				c.AddTask(&awstasks.RouteTableAssociation{
					Name:       s(subnetSpec.Name + "." + b.ClusterName()),
					RouteTable: publicRouteTable,
					Subnet:     subnet,
				})
			}

		case kops.SubnetTypePrivate:
			// Private subnets get a Network Gateway, and their own route table to associate them with the network gateway

			if !sharedSubnet {
				// Private Subnet Route Table Associations
				//
				// Map the Private subnet to the Private route table
				c.AddTask(&awstasks.RouteTableAssociation{
					Name:       s("private-" + subnetSpec.Name + "." + b.ClusterName()),
					RouteTable: &awstasks.RouteTable{Name: s(b.NamePrivateRouteTableInZone(subnetSpec.Zone))},
					Subnet:     subnet,
				})

				privateZones.Insert(subnetSpec.Zone)
			}

		default:
			return fmt.Errorf("subnet %q has unknown type %q", subnetSpec.Name, subnetSpec.Type)
		}
	}

	for _, zone := range privateZones.List() {
		utilitySubnet, err := b.LinkToUtilitySubnetInZone(zone)
		if err != nil {
			return err
		}
		// Every NGW needs a public (Elastic) IP address, every private
		// subnet needs a NGW, lets create it. We tie it to a subnet
		// so we can track it in AWS
		eip := &awstasks.ElasticIP{
			Name:   s(zone + "." + b.ClusterName()),
			Subnet: utilitySubnet,
		}
		c.AddTask(eip)

		// NAT Gateway
		//
		// All private subnets will need a NGW, one per zone
		//
		// The instances in the private subnet can access the Internet by
		// using a network address translation (NAT) gateway that resides
		// in the public subnet.
		ngw := &awstasks.NatGateway{
			Name:      s(zone + "." + b.ClusterName()),
			Subnet:    utilitySubnet,
			ElasticIp: eip,
		}
		c.AddTask(ngw)

		// Private Route Table
		//
		// The private route table that will route to the NAT Gateway
		rt := &awstasks.RouteTable{
			Name: s(b.NamePrivateRouteTableInZone(zone)),
			VPC:  b.LinkToVPC(),
		}
		c.AddTask(rt)

		// Private Routes
		//
		// Routes for the private route table.
		// Will route to the NAT Gateway
		c.AddTask(&awstasks.Route{
			Name:       s("private-" + zone + "-0.0.0.0/0"),
			CIDR:       s("0.0.0.0/0"),
			RouteTable: rt,
			NatGateway: ngw,
		})

	}

	return nil
}
