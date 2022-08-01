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

package awsmodel

import (
	"fmt"
	"strings"

	aws "k8s.io/cloud-provider-aws/pkg/providers/v1"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*AWSModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

type zoneInfo struct {
	NATSubnets           []*kops.ClusterSubnetSpec
	HaveIPv6PublicSubnet bool
	HavePrivateSubnet    bool
}

func isUnmanaged(subnet *kops.ClusterSubnetSpec) bool {
	return subnet.Egress == kops.EgressExternal
}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	sharedVPC := b.Cluster.SharedVPC()
	vpcName := b.ClusterName()
	tags := b.CloudTags(vpcName, sharedVPC)

	// VPC that holds everything for the cluster
	{
		vpcTags := tags
		if sharedVPC {
			// We don't tag a shared VPC - we can identify it by its ID anyway.  Issue #4265
			vpcTags = nil
		}
		t := &awstasks.VPC{
			Name:             fi.String(vpcName),
			Lifecycle:        b.Lifecycle,
			Shared:           fi.Bool(sharedVPC),
			EnableDNSSupport: fi.Bool(true),
			Tags:             vpcTags,
		}

		if sharedVPC {
			// If we have e.g.  --kubelet-preferred-address-types=InternalIP,Hostname,ExternalIP,LegacyHostIP
			// then we don't need EnableDNSHostnames
			klog.V(4).Info("Skipping EnableDNSHostnames requirement on VPC")
		} else {
			// In theory we don't need to enable it for >= 1.5,
			// but seems safer to stick with existing behaviour
			t.EnableDNSHostnames = fi.Bool(true)

			// Used only for Terraform rendering.
			// Direct and CloudFormation rendering is handled via the VPCAmazonIPv6CIDRBlock task
			t.AmazonIPv6 = fi.Bool(true)
			t.AssociateExtraCIDRBlocks = b.Cluster.Spec.AdditionalNetworkCIDRs
		}

		if b.Cluster.Spec.NetworkID != "" {
			t.ID = fi.String(b.Cluster.Spec.NetworkID)
		}

		if b.Cluster.Spec.NetworkCIDR != "" {
			t.CIDR = fi.String(b.Cluster.Spec.NetworkCIDR)
		}

		c.AddTask(t)
	}

	if !sharedVPC {
		// Associate an Amazon-provided IPv6 CIDR block with the VPC
		c.AddTask(&awstasks.VPCAmazonIPv6CIDRBlock{
			Name:      fi.String("AmazonIPv6"),
			Lifecycle: b.Lifecycle,
			VPC:       b.LinkToVPC(),
			Shared:    fi.Bool(false),
		})

		// Associate additional CIDR blocks with the VPC
		for _, cidr := range b.Cluster.Spec.AdditionalNetworkCIDRs {
			c.AddTask(&awstasks.VPCCIDRBlock{
				Name:      fi.String(cidr),
				Lifecycle: b.Lifecycle,
				VPC:       b.LinkToVPC(),
				Shared:    fi.Bool(false),
				CIDRBlock: fi.String(cidr),
			})
		}
	}

	// TODO: would be good to create these as shared, to verify them
	if !sharedVPC {
		dhcp := &awstasks.DHCPOptions{
			Name:              fi.String(b.ClusterName()),
			Lifecycle:         b.Lifecycle,
			DomainNameServers: fi.String("AmazonProvidedDNS"),

			Tags:   tags,
			Shared: fi.Bool(sharedVPC),
		}
		if b.Region == "us-east-1" {
			dhcp.DomainName = fi.String("ec2.internal")
		} else {
			dhcp.DomainName = fi.String(b.Region + ".compute.internal")
		}
		c.AddTask(dhcp)

		c.AddTask(&awstasks.VPCDHCPOptionsAssociation{
			Name:        fi.String(b.ClusterName()),
			Lifecycle:   b.Lifecycle,
			VPC:         b.LinkToVPC(),
			DHCPOptions: dhcp,
		})
	}

	allSubnetsUnmanaged := true
	allPrivateSubnetsUnmanaged := true
	allSubnetsShared := true
	allSubnetsSharedInZone := make(map[string]bool)
	for i := range b.Cluster.Spec.Subnets {
		subnetSpec := &b.Cluster.Spec.Subnets[i]
		allSubnetsSharedInZone[subnetSpec.Zone] = true
	}

	for i := range b.Cluster.Spec.Subnets {
		subnetSpec := &b.Cluster.Spec.Subnets[i]
		sharedSubnet := subnetSpec.ProviderID != ""
		if !sharedSubnet {
			allSubnetsShared = false
			allSubnetsSharedInZone[subnetSpec.Zone] = false
		}

		if !isUnmanaged(subnetSpec) {
			allSubnetsUnmanaged = false
			if subnetSpec.Type == kops.SubnetTypeDualStack || subnetSpec.Type == kops.SubnetTypePrivate {
				allPrivateSubnetsUnmanaged = false
			}
		}
	}

	// We always have a public route table, though for private networks it is only used for NGWs and ELBs
	var publicRouteTable *awstasks.RouteTable
	var igw *awstasks.InternetGateway
	if !allSubnetsUnmanaged {
		// The internet gateway is the main entry point to the cluster.
		igw = &awstasks.InternetGateway{
			Name:      fi.String(b.ClusterName()),
			Lifecycle: b.Lifecycle,
			VPC:       b.LinkToVPC(),
			Shared:    fi.Bool(sharedVPC),
		}
		igw.Tags = b.CloudTags(*igw.Name, *igw.Shared)
		c.AddTask(igw)

		if !allSubnetsShared {
			// The route table is not shared if we're creating a subnet for our cluster
			// That subnet will be owned, and will be associated with our RouteTable.
			// On deletion we delete the subnet & the route table.
			sharedRouteTable := false
			routeTableTags := b.CloudTags(vpcName, sharedRouteTable)
			routeTableTags[awsup.TagNameKopsRole] = "public"
			publicRouteTable = &awstasks.RouteTable{
				Name:      fi.String(b.ClusterName()),
				Lifecycle: b.Lifecycle,

				VPC: b.LinkToVPC(),

				Tags:   routeTableTags,
				Shared: fi.Bool(sharedRouteTable),
			}
			c.AddTask(publicRouteTable)

			// TODO: Validate when allSubnetsShared
			c.AddTask(&awstasks.Route{
				Name:            fi.String("0.0.0.0/0"),
				Lifecycle:       b.Lifecycle,
				CIDR:            fi.String("0.0.0.0/0"),
				RouteTable:      publicRouteTable,
				InternetGateway: igw,
			})
			c.AddTask(&awstasks.Route{
				Name:            fi.String("::/0"),
				Lifecycle:       b.Lifecycle,
				IPv6CIDR:        fi.String("::/0"),
				RouteTable:      publicRouteTable,
				InternetGateway: igw,
			})
		}
	}

	infoByZone := make(map[string]*zoneInfo)

	haveDualStack := map[string]bool{}
	for _, subnetSpec := range b.Cluster.Spec.Subnets {
		if subnetSpec.Type == kops.SubnetTypeDualStack {
			haveDualStack[subnetSpec.Zone] = true
		}
	}

	for i := range b.Cluster.Spec.Subnets {
		subnetSpec := &b.Cluster.Spec.Subnets[i]
		sharedSubnet := subnetSpec.ProviderID != ""
		subnetName := subnetSpec.Name + "." + b.ClusterName()
		tags := map[string]string{}

		// Apply tags so that Kubernetes knows which subnets should be used for internal/external ELBs
		if b.Cluster.Spec.TagSubnets == nil || *b.Cluster.Spec.TagSubnets {
			klog.V(2).Infof("applying subnet tags")
			tags = b.CloudTags(subnetName, sharedSubnet)
			tags["SubnetType"] = string(subnetSpec.Type)

			switch subnetSpec.Type {
			case kops.SubnetTypePublic, kops.SubnetTypeUtility:
				tags[aws.TagNameSubnetPublicELB] = "1"

				// AWS ALB contoller won't provision any internal ELBs unless this tag is set.
				// So we add this to public subnets as well if we do not expect any private subnets.
				if b.Cluster.Spec.Topology.Nodes == kops.TopologyPublic {
					tags[aws.TagNameSubnetInternalELB] = "1"
				}

			case kops.SubnetTypeDualStack:
				tags[aws.TagNameSubnetInternalELB] = "1"

			case kops.SubnetTypePrivate:
				if !haveDualStack[subnetSpec.Zone] {
					tags[aws.TagNameSubnetInternalELB] = "1"
				}

			default:
				klog.V(2).Infof("unable to properly tag subnet %q because it has unknown type %q. Load balancers may be created in incorrect subnets", subnetSpec.Name, subnetSpec.Type)
			}

			for _, ig := range b.InstanceGroups {
				for _, igSubnetName := range ig.Spec.Subnets {
					if subnetSpec.Name == igSubnetName {
						tags["kops.k8s.io/instance-group/"+ig.GetName()] = "true"
					}
				}
			}

		} else {
			klog.V(2).Infof("skipping subnet tags. Ensure these are maintained externally.")
		}

		subnet := &awstasks.Subnet{
			Name:             fi.String(subnetName),
			ShortName:        fi.String(subnetSpec.Name),
			Lifecycle:        b.Lifecycle,
			VPC:              b.LinkToVPC(),
			AvailabilityZone: fi.String(subnetSpec.Zone),
			Shared:           fi.Bool(sharedSubnet),
			Tags:             tags,
		}

		if b.Cluster.Spec.ExternalCloudControllerManager != nil && b.Cluster.IsKubernetesGTE("1.22") {
			subnet.ResourceBasedNaming = fi.Bool(true)
		}

		if subnetSpec.CIDR != "" {
			subnet.CIDR = fi.String(subnetSpec.CIDR)
		}

		if subnetSpec.IPv6CIDR != "" {
			if !sharedVPC {
				subnet.AmazonIPv6CIDR = b.LinkToAmazonVPCIPv6CIDR()
			}
			subnet.IPv6CIDR = fi.String(subnetSpec.IPv6CIDR)
		}
		if subnetSpec.ProviderID != "" {
			subnet.ID = fi.String(subnetSpec.ProviderID)
		}
		c.AddTask(subnet)

		switch subnetSpec.Type {
		case kops.SubnetTypePublic, kops.SubnetTypeUtility:
			if !sharedSubnet && !isUnmanaged(subnetSpec) {
				if b.IsIPv6Only() && subnetSpec.Type == kops.SubnetTypePublic && subnetSpec.IPv6CIDR != "" {
					// Public IPv6-capable subnets route NAT64 to a NAT gateway
					c.AddTask(&awstasks.RouteTableAssociation{
						Name:       fi.String("public-" + subnetSpec.Name + "." + b.ClusterName()),
						Lifecycle:  b.Lifecycle,
						RouteTable: b.LinkToPublicRouteTableInZone(subnetSpec.Zone),
						Subnet:     subnet,
					})

					if infoByZone[subnetSpec.Zone] == nil {
						infoByZone[subnetSpec.Zone] = &zoneInfo{}
					}
					infoByZone[subnetSpec.Zone].NATSubnets = append(infoByZone[subnetSpec.Zone].NATSubnets, subnetSpec)
					infoByZone[subnetSpec.Zone].HaveIPv6PublicSubnet = true
				} else {
					c.AddTask(&awstasks.RouteTableAssociation{
						Name:       fi.String(subnetSpec.Name + "." + b.ClusterName()),
						Lifecycle:  b.Lifecycle,
						RouteTable: publicRouteTable,
						Subnet:     subnet,
					})
				}
			}

		case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
			// Private subnets get a Network Gateway, and their own route table to associate them with the network gateway

			if !sharedSubnet && !isUnmanaged(subnetSpec) {
				// Private Subnet Route Table Associations
				//
				// Map the Private subnet to the Private route table
				c.AddTask(&awstasks.RouteTableAssociation{
					Name:       fi.String("private-" + subnetSpec.Name + "." + b.ClusterName()),
					Lifecycle:  b.Lifecycle,
					RouteTable: b.LinkToPrivateRouteTableInZone(subnetSpec.Zone),
					Subnet:     subnet,
				})

				// TODO: validate even if shared?
				if infoByZone[subnetSpec.Zone] == nil {
					infoByZone[subnetSpec.Zone] = &zoneInfo{}
				}
				infoByZone[subnetSpec.Zone].NATSubnets = append(infoByZone[subnetSpec.Zone].NATSubnets, subnetSpec)
				infoByZone[subnetSpec.Zone].HavePrivateSubnet = true
			}
		default:
			return fmt.Errorf("subnet %q has unknown type %q", subnetSpec.Name, subnetSpec.Type)
		}
	}

	// Set up private route tables & egress

	// The instances in the private subnet can access the IPv6 Internet by
	// using an egress-only internet gateway.
	var eigw *awstasks.EgressOnlyInternetGateway
	if !allPrivateSubnetsUnmanaged && b.IsIPv6Only() {
		eigw = &awstasks.EgressOnlyInternetGateway{
			Name:      fi.String(b.ClusterName()),
			Lifecycle: b.Lifecycle,
			VPC:       b.LinkToVPC(),
			Shared:    fi.Bool(sharedVPC),
		}
		eigw.Tags = b.CloudTags(*eigw.Name, *eigw.Shared)
		c.AddTask(eigw)
	}

	for zone, info := range infoByZone {
		if len(info.NATSubnets) == 0 {
			continue
		}

		var egressSubnet *awstasks.Subnet
		var egressRouteTable *awstasks.RouteTable
		var err error
		if info.HavePrivateSubnet {
			egressSubnet, err = b.LinkToUtilitySubnetInZone(zone)
			egressRouteTable = b.LinkToPrivateRouteTableInZone(zone)
		} else {
			egressSubnet, err = b.LinkToPublicSubnetInZone(zone)
			egressRouteTable = b.LinkToPublicRouteTableInZone(zone)
		}
		if err != nil {
			return err
		}

		egress := info.NATSubnets[0].Egress
		publicIP := info.NATSubnets[0].PublicIP

		allUnmanaged := true
		for _, subnetSpec := range info.NATSubnets {
			if !isUnmanaged(subnetSpec) {
				allUnmanaged = false
			}
		}
		if allUnmanaged {
			klog.V(4).Infof("skipping network configuration in zone %s - all subnets unmanaged", zone)
			continue
		}

		// Verify we don't have mixed values for egress/publicIP - the code doesn't handle it
		for _, subnet := range info.NATSubnets {
			if subnet.Egress != egress {
				return fmt.Errorf("cannot mix egress values in private or IPv6-capable subnets")
			}
			if subnet.PublicIP != publicIP {
				return fmt.Errorf("cannot mix publicIP values in private or IPv6-capable subnets")
			}
		}

		var ngw *awstasks.NatGateway
		var tgwID *string
		var in *awstasks.Instance
		if egress != "" {
			if strings.HasPrefix(egress, "nat-") {

				ngw = &awstasks.NatGateway{
					Name:                 fi.String(zone + "." + b.ClusterName()),
					Lifecycle:            b.Lifecycle,
					Subnet:               egressSubnet,
					ID:                   fi.String(egress),
					AssociatedRouteTable: egressRouteTable,
					// If we're here, it means this NatGateway was specified, so we are Shared
					Shared: fi.Bool(true),
					Tags:   b.CloudTags(zone+"."+b.ClusterName(), true),
				}

				c.AddTask(ngw)

			} else if strings.HasPrefix(egress, "eipalloc-") {

				eip := &awstasks.ElasticIP{
					Name:                           fi.String(zone + "." + b.ClusterName()),
					ID:                             fi.String(egress),
					Lifecycle:                      b.Lifecycle,
					AssociatedNatGatewayRouteTable: egressRouteTable,
					Shared:                         fi.Bool(true),
					Tags:                           b.CloudTags(zone+"."+b.ClusterName(), true),
				}
				c.AddTask(eip)

				ngw = &awstasks.NatGateway{
					Name:                 fi.String(zone + "." + b.ClusterName()),
					Lifecycle:            b.Lifecycle,
					Subnet:               egressSubnet,
					ElasticIP:            eip,
					AssociatedRouteTable: egressRouteTable,
					Tags:                 b.CloudTags(zone+"."+b.ClusterName(), false),
				}
				c.AddTask(ngw)

			} else if strings.HasPrefix(egress, "i-") {

				in = &awstasks.Instance{
					Name:      fi.String(egress),
					Lifecycle: b.Lifecycle,
					ID:        fi.String(egress),
					Shared:    fi.Bool(true),
					Tags:      nil, // We don't need to add tags here
				}

				c.AddTask(in)
			} else if strings.HasPrefix(egress, "tgw-") {
				tgwID = &egress
			} else if egress == "External" {
				// Nothing to do here
			} else {
				return fmt.Errorf("kops currently only supports re-use of either NAT EC2 Instances or NAT Gateways. We will support more eventually! Please see https://github.com/kubernetes/kops/issues/1530")
			}
		} else {

			// Every NGW needs a public (Elastic) IP address, every private
			// subnet needs a NGW, lets create it. We tie it to a subnet
			// so we can track it in AWS
			eip := &awstasks.ElasticIP{
				Name:                           fi.String(zone + "." + b.ClusterName()),
				Lifecycle:                      b.Lifecycle,
				AssociatedNatGatewayRouteTable: egressRouteTable,
			}

			if publicIP != "" {
				eip.PublicIP = fi.String(publicIP)
				eip.Tags = b.CloudTags(*eip.Name, true)
			} else {
				eip.Tags = b.CloudTags(*eip.Name, false)
			}

			c.AddTask(eip)
			// NAT Gateway
			//
			// All private subnets will need a NGW, one per zone
			//
			// The instances in the private subnet can access the IPv4 Internet by
			// using a network address translation (NAT) gateway that resides
			// in the public subnet.

			// var ngw = &awstasks.NatGateway{}
			ngw = &awstasks.NatGateway{
				Name:                 fi.String(zone + "." + b.ClusterName()),
				Lifecycle:            b.Lifecycle,
				Subnet:               egressSubnet,
				ElasticIP:            eip,
				AssociatedRouteTable: egressRouteTable,
				Tags:                 b.CloudTags(zone+"."+b.ClusterName(), false),
			}
			c.AddTask(ngw)
		}

		if info.HavePrivateSubnet {
			// Private Route Table
			//
			// We create an owned route table if we created any private subnet in that zone.
			// Otherwise we consider it shared.
			routeTableShared := allSubnetsSharedInZone[zone]
			routeTableTags := b.CloudTags(b.NamePrivateRouteTableInZone(zone), routeTableShared)
			routeTableTags[awsup.TagNameKopsRole] = "private-" + zone
			rt := &awstasks.RouteTable{
				Name:      fi.String(b.NamePrivateRouteTableInZone(zone)),
				VPC:       b.LinkToVPC(),
				Lifecycle: b.Lifecycle,

				Shared: fi.Bool(routeTableShared),
				Tags:   routeTableTags,
			}
			c.AddTask(rt)

			// Private Routes
			//
			// Routes for the private route table.
			// Will route IPv4 to the NAT Gateway
			var r *awstasks.Route
			if in != nil {
				r = &awstasks.Route{
					Name:       fi.String("private-" + zone + "-0.0.0.0/0"),
					Lifecycle:  b.Lifecycle,
					CIDR:       fi.String("0.0.0.0/0"),
					RouteTable: rt,
					Instance:   in,
				}
			} else {
				r = &awstasks.Route{
					Name:       fi.String("private-" + zone + "-0.0.0.0/0"),
					Lifecycle:  b.Lifecycle,
					CIDR:       fi.String("0.0.0.0/0"),
					RouteTable: rt,
					// Only one of these will be not nil
					NatGateway:       ngw,
					TransitGatewayID: tgwID,
				}
			}
			c.AddTask(r)

			if b.IsIPv6Only() {
				// Route NAT64 well-known prefix to the NAT gateway
				c.AddTask(&awstasks.Route{
					Name:       fi.String("private-" + zone + "-64:ff9b::/96"),
					Lifecycle:  b.Lifecycle,
					IPv6CIDR:   fi.String("64:ff9b::/96"),
					RouteTable: rt,
					// Only one of these will be not nil
					NatGateway:       ngw,
					TransitGatewayID: tgwID,
				})

				// Route IPv6 to the Egress-only Internet Gateway.
				c.AddTask(&awstasks.Route{
					Name:                      fi.String("private-" + zone + "-::/0"),
					Lifecycle:                 b.Lifecycle,
					IPv6CIDR:                  fi.String("::/0"),
					RouteTable:                rt,
					EgressOnlyInternetGateway: eigw,
				})
			}

			subnets, err := b.LinkToPrivateSubnetsInZone(zone)
			if err != nil {
				return err
			}

			for _, subnetSpec := range b.Cluster.Spec.Subnets {
				for _, subnet := range subnets {
					if strings.HasPrefix(*subnet.Name, subnetSpec.Name) {
						err := addAdditionalRoutes(subnetSpec.AdditionalRoutes, subnetSpec.Name, rt, b.Lifecycle, c)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		if info.HaveIPv6PublicSubnet {
			// Public Route Table
			//
			// We create an owned route table if we created any IPv6-capable public subnet in that zone.
			// Otherwise we consider it shared.
			routeTableShared := allSubnetsSharedInZone[zone]
			routeTableTags := b.CloudTags(b.NamePublicRouteTableInZone(zone), routeTableShared)
			routeTableTags[awsup.TagNameKopsRole] = "public-" + zone
			rt := &awstasks.RouteTable{
				Name:      fi.String(b.NamePublicRouteTableInZone(zone)),
				VPC:       b.LinkToVPC(),
				Lifecycle: b.Lifecycle,

				Shared: fi.Bool(routeTableShared),
				Tags:   routeTableTags,
			}
			c.AddTask(rt)

			// Routes for the public route table.
			c.AddTask(&awstasks.Route{
				Name:            fi.String("public-" + zone + "-0.0.0.0/0"),
				Lifecycle:       b.Lifecycle,
				CIDR:            fi.String("0.0.0.0/0"),
				RouteTable:      rt,
				InternetGateway: igw,
			})
			c.AddTask(&awstasks.Route{
				Name:            fi.String("public-" + zone + "-::/0"),
				Lifecycle:       b.Lifecycle,
				IPv6CIDR:        fi.String("::/0"),
				RouteTable:      rt,
				InternetGateway: igw,
			})

			// Route NAT64 well-known prefix to the NAT gateway
			c.AddTask(&awstasks.Route{
				Name:       fi.String("public-" + zone + "-64:ff9b::/96"),
				Lifecycle:  b.Lifecycle,
				IPv6CIDR:   fi.String("64:ff9b::/96"),
				RouteTable: rt,
				// Only one of these will be not nil
				NatGateway:       ngw,
				TransitGatewayID: tgwID,
			})
		}
	}

	return nil
}

func addAdditionalRoutes(routes []kops.RouteSpec, sbName string, rt *awstasks.RouteTable, lf fi.Lifecycle, c *fi.ModelBuilderContext) error {
	for _, r := range routes {
		t := &awstasks.Route{
			Name:       fi.String(sbName + "." + r.CIDR),
			Lifecycle:  lf,
			CIDR:       fi.String(r.CIDR),
			RouteTable: rt,
		}
		if strings.HasPrefix(r.Target, "pcx-") {
			t.VPCPeeringConnectionID = fi.String(r.Target)
			c.AddTask(t)
		} else if strings.HasPrefix(r.Target, "i-") {
			inst := &awstasks.Instance{
				Name:      fi.String(r.Target),
				Lifecycle: lf,
				ID:        fi.String(r.Target),
				Shared:    fi.Bool(true),
			}
			err := c.EnsureTask(inst)
			if err != nil {
				return err
			}
			t.Instance = inst
			c.AddTask(t)
		} else if strings.HasPrefix(r.Target, "nat-") {
			nat := &awstasks.NatGateway{
				Name:      fi.String(r.Target),
				Lifecycle: lf,
				ID:        fi.String(r.Target),
				Shared:    fi.Bool(true),
			}
			err := c.EnsureTask(nat)
			if err != nil {
				return err
			}
			t.NatGateway = nat
			c.AddTask(t)
		} else if strings.HasPrefix(r.Target, "tgw-") {
			t.TransitGatewayID = fi.String(r.Target)
			c.AddTask(t)
		} else if strings.HasPrefix(r.Target, "igw-") {
			internetGW := &awstasks.InternetGateway{
				Name:      fi.String(r.Target),
				Lifecycle: lf,
				ID:        fi.String(r.Target),
				Shared:    fi.Bool(true),
			}
			err := c.EnsureTask(internetGW)
			if err != nil {
				return err
			}
			t.InternetGateway = internetGW
			c.AddTask(t)
		} else if strings.HasPrefix(r.Target, "eigw-") {
			eigw := &awstasks.EgressOnlyInternetGateway{
				Name:      fi.String(r.Target),
				Lifecycle: lf,
				ID:        fi.String(r.Target),
				Shared:    fi.Bool(true),
			}
			err := c.EnsureTask(eigw)
			if err != nil {
				return err
			}
			t.EgressOnlyInternetGateway = eigw
			c.AddTask(t)
		}
	}
	return nil
}
