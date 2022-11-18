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
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// BastionModelBuilder adds model objects to support bastions
//
// Bastion instances live in the utility subnets created in the private topology.
// All traffic goes through an ELB, and the ELB has port 22 open to SSHAccess.
// Bastion instances have access to all internal master and node instances.

type BastionModelBuilder struct {
	*AWSModelContext
	Lifecycle         fi.Lifecycle
	SecurityLifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &BastionModelBuilder{}

func (b *BastionModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var bastionInstanceGroups []*kops.InstanceGroup
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			bastionInstanceGroups = append(bastionInstanceGroups, ig)
		}
	}

	if len(bastionInstanceGroups) == 0 {
		return nil
	}

	bastionGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleBastion)
	if err != nil {
		return err
	}
	nodeGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleNode)
	if err != nil {
		return err
	}
	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleMaster)
	if err != nil {
		return err
	}

	// Create security group for bastion instances
	for _, bastionGroup := range bastionGroups {
		bastionGroup.Task.Lifecycle = b.SecurityLifecycle
		c.AddTask(bastionGroup.Task)
	}

	for _, src := range bastionGroups {
		// Allow traffic from bastion instances to egress freely
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv4-bastion-egress" + src.Suffix),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.PtrTo(true),
				CIDR:          fi.PtrTo("0.0.0.0/0"),
			}
			AddDirectionalGroupRule(c, t)
		}
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv6-bastion-egress" + src.Suffix),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: src.Task,
				Egress:        fi.PtrTo(true),
				IPv6CIDR:      fi.PtrTo("::/0"),
			}
			AddDirectionalGroupRule(c, t)
		}
	}

	var bastionLoadBalancerType kops.LoadBalancerType
	{
		// Check if we requested a public or internal NLB
		if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Bastion != nil && b.Cluster.Spec.Topology.Bastion.LoadBalancer != nil {
			if b.Cluster.Spec.Topology.Bastion.LoadBalancer.Type != "" {
				switch b.Cluster.Spec.Topology.Bastion.LoadBalancer.Type {
				case kops.LoadBalancerTypeInternal:
					bastionLoadBalancerType = "Internal"
				case kops.LoadBalancerTypePublic:
					bastionLoadBalancerType = "Public"
				default:
					return fmt.Errorf("unhandled bastion LoadBalancer type %q", b.Cluster.Spec.Topology.Bastion.LoadBalancer.Type)
				}
			} else {
				// Default to Public
				b.Cluster.Spec.Topology.Bastion.LoadBalancer.Type = kops.LoadBalancerTypePublic
				bastionLoadBalancerType = "Public"
			}
		} else {
			// Default to Public
			bastionLoadBalancerType = "Public"
		}
	}

	// Allow bastion nodes to SSH to masters
	for _, src := range bastionGroups {
		for _, dest := range masterGroups {
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("bastion-to-master-ssh" + JoinSuffixes(src, dest)),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
				Protocol:      fi.PtrTo("tcp"),
				FromPort:      fi.PtrTo(int64(22)),
				ToPort:        fi.PtrTo(int64(22)),
			}
			AddDirectionalGroupRule(c, t)
		}
	}

	// Allow bastion nodes to SSH to nodes
	for _, src := range bastionGroups {
		for _, dest := range nodeGroups {
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("bastion-to-node-ssh" + JoinSuffixes(src, dest)),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: dest.Task,
				SourceGroup:   src.Task,
				Protocol:      fi.PtrTo("tcp"),
				FromPort:      fi.PtrTo(int64(22)),
				ToPort:        fi.PtrTo(int64(22)),
			}
			AddDirectionalGroupRule(c, t)
		}
	}

	var sshAllowedCIDRs []string
	var nlbSubnetMappings []*awstasks.SubnetMapping
	{
		// Compute the subnets - only one per zone, and then break ties based on chooseBestSubnetForNLB
		subnetsByZone := make(map[string][]*kops.ClusterSubnetSpec)
		for i := range b.Cluster.Spec.Subnets {
			subnet := &b.Cluster.Spec.Subnets[i]

			switch subnet.Type {
			case kops.SubnetTypePublic, kops.SubnetTypeUtility:
				if bastionLoadBalancerType != kops.LoadBalancerTypePublic {
					continue
				}

			case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
				if bastionLoadBalancerType != kops.LoadBalancerTypeInternal {
					continue
				}

			default:
				return fmt.Errorf("subnet %q had unknown type %q", subnet.Name, subnet.Type)
			}

			subnetsByZone[subnet.Zone] = append(subnetsByZone[subnet.Zone], subnet)
		}

		for zone, subnets := range subnetsByZone {
			for _, subnet := range subnets {
				sshAllowedCIDRs = append(sshAllowedCIDRs, subnet.CIDR)
			}
			subnet := b.chooseBestSubnetForNLB(zone, subnets)
			nlbSubnetMappings = append(nlbSubnetMappings, &awstasks.SubnetMapping{Subnet: b.LinkToSubnet(subnet)})
		}
	}

	sshAllowedCIDRs = append(sshAllowedCIDRs, b.Cluster.Spec.SSHAccess...)
	for _, cidr := range sshAllowedCIDRs {
		// Allow incoming SSH traffic to bastions, through the NLB
		// TODO: Could we get away without an NLB here?  Tricky to fix if dns-controller breaks though...
		for _, bastionGroup := range bastionGroups {
			{
				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo(fmt.Sprintf("ssh-nlb-%s", cidr)),
					Lifecycle:     b.SecurityLifecycle,
					SecurityGroup: bastionGroup.Task,
					Protocol:      fi.PtrTo("tcp"),
					FromPort:      fi.PtrTo(int64(22)),
					ToPort:        fi.PtrTo(int64(22)),
				}
				t.SetCidrOrPrefix(cidr)
				AddDirectionalGroupRule(c, t)
			}

			if strings.HasPrefix(cidr, "pl-") {
				// In case of a prefix list we do not add a rule for ICMP traffic for PMTU discovery.
				// This would require calling out to AWS to check whether the prefix list is IPv4 or IPv6.
			} else if utils.IsIPv6CIDR(cidr) {
				// Allow ICMP traffic required for PMTU discovery
				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo("icmpv6-pmtu-ssh-nlb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      fi.PtrTo(int64(-1)),
					Protocol:      fi.PtrTo("icmpv6"),
					SecurityGroup: bastionGroup.Task,
					ToPort:        fi.PtrTo(int64(-1)),
				}
				t.SetCidrOrPrefix(cidr)
				c.AddTask(t)
			} else {
				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo("icmp-pmtu-ssh-nlb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      fi.PtrTo(int64(3)),
					Protocol:      fi.PtrTo("icmp"),
					SecurityGroup: bastionGroup.Task,
					ToPort:        fi.PtrTo(int64(4)),
				}
				t.SetCidrOrPrefix(cidr)
				c.AddTask(t)
			}
		}
	}

	// Create NLB itself
	var nlb *awstasks.NetworkLoadBalancer
	{
		loadBalancerName := b.LBName32("bastion")

		tags := b.CloudTags(loadBalancerName, false)
		for k, v := range b.Cluster.Spec.CloudLabels {
			tags[k] = v
		}
		// Override the returned name to be the expected ELB name
		tags["Name"] = "bastion." + b.ClusterName()

		nlbListeners := []*awstasks.NetworkLoadBalancerListener{
			{
				Port:            22,
				TargetGroupName: b.NLBTargetGroupName("bastion"),
			},
		}
		nlb = &awstasks.NetworkLoadBalancer{
			Name:      fi.PtrTo(b.NLBName("bastion")),
			Lifecycle: b.Lifecycle,

			LoadBalancerName: fi.PtrTo(loadBalancerName),
			CLBName:          fi.PtrTo("bastion." + b.ClusterName()),
			SubnetMappings:   nlbSubnetMappings,
			Listeners:        nlbListeners,
			TargetGroups:     make([]*awstasks.TargetGroup, 0),

			Tags:          tags,
			VPC:           b.LinkToVPC(),
			Type:          fi.PtrTo("network"),
			IpAddressType: fi.PtrTo("ipv4"),
		}
		if useIPv6ForBastion(b) {
			nlb.IpAddressType = fi.PtrTo("dualstack")
		}
		// Set the NLB Scheme according to load balancer Type
		switch bastionLoadBalancerType {
		case kops.LoadBalancerTypeInternal:
			nlb.Scheme = fi.PtrTo("internal")
		case kops.LoadBalancerTypePublic:
			nlb.Scheme = nil
		default:
			return fmt.Errorf("unhandled bastion LoadBalancer type %q", bastionLoadBalancerType)
		}

		sshGroupName := b.NLBTargetGroupName("bastion")
		sshGroupTags := b.CloudTags(sshGroupName, false)

		// Override the returned name to be the expected NLB TG name
		sshGroupTags["Name"] = sshGroupName

		tg := &awstasks.TargetGroup{
			Name:               fi.PtrTo(sshGroupName),
			Lifecycle:          b.Lifecycle,
			VPC:                b.LinkToVPC(),
			Tags:               sshGroupTags,
			Protocol:           fi.PtrTo("TCP"),
			Port:               fi.PtrTo(int64(22)),
			Interval:           fi.PtrTo(int64(10)),
			HealthyThreshold:   fi.PtrTo(int64(2)),
			UnhealthyThreshold: fi.PtrTo(int64(2)),
			Shared:             fi.PtrTo(false),
		}

		c.AddTask(tg)

		nlb.TargetGroups = append(nlb.TargetGroups, tg)

		sort.Stable(awstasks.OrderTargetGroupsByName(nlb.TargetGroups))
		c.AddTask(nlb)
	}

	publicName := ""
	if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Bastion != nil {
		publicName = b.Cluster.Spec.Topology.Bastion.PublicName
	}
	if publicName != "" {
		// Here we implement the bastion CNAME logic
		// By default bastions will create a CNAME that follows the `bastion-$clustername` formula
		t := &awstasks.DNSName{
			Name:      fi.PtrTo(publicName),
			Lifecycle: b.Lifecycle,

			Zone:               b.LinkToDNSZone(),
			ResourceName:       fi.PtrTo(publicName),
			ResourceType:       fi.PtrTo("A"),
			TargetLoadBalancer: b.LinkToNLB("bastion"),
		}
		c.AddTask(t)
		if *nlb.IpAddressType == "dualstack" {
			t := &awstasks.DNSName{
				Name:      fi.PtrTo(publicName + "-AAAA"),
				Lifecycle: b.Lifecycle,

				Zone:               b.LinkToDNSZone(),
				ResourceName:       fi.PtrTo(publicName),
				ResourceType:       fi.PtrTo("AAAA"),
				TargetLoadBalancer: b.LinkToNLB("bastion"),
			}
			c.AddTask(t)
		}

	}
	return nil
}

func useIPv6ForBastion(b *BastionModelBuilder) bool {
	for _, ig := range b.InstanceGroups {
		for _, igSubnetName := range ig.Spec.Subnets {
			for _, clusterSubnet := range b.Cluster.Spec.Subnets {
				if igSubnetName != clusterSubnet.Name {
					continue
				}
				if clusterSubnet.IPv6CIDR != "" {
					return true
				}
			}
		}
	}
	return false
}

// Choose between subnets in a zone.
// We have already applied the rules to match internal subnets to internal NLBs and vice-versa for public-facing NLBs.
// For internal NLBs: we prefer the master subnets
// For public facing NLBs: we prefer the utility subnets
func (b *BastionModelBuilder) chooseBestSubnetForNLB(zone string, subnets []*kops.ClusterSubnetSpec) *kops.ClusterSubnetSpec {
	if len(subnets) == 0 {
		return nil
	}
	if len(subnets) == 1 {
		return subnets[0]
	}

	migSubnets := sets.NewString()
	for _, ig := range b.MasterInstanceGroups() {
		for _, subnet := range ig.Spec.Subnets {
			migSubnets.Insert(subnet)
		}
	}

	var scoredSubnets []*scoredSubnet
	for _, subnet := range subnets {
		score := 0

		if migSubnets.Has(subnet.Name) {
			score += 1
		}

		if subnet.Type == kops.SubnetTypeDualStack {
			score += 2
		}

		if subnet.Type == kops.SubnetTypeUtility {
			score += 3
		}

		scoredSubnets = append(scoredSubnets, &scoredSubnet{
			score:  score,
			subnet: subnet,
		})
	}

	sort.Sort(ByScoreDescending(scoredSubnets))

	if scoredSubnets[0].score == scoredSubnets[1].score {
		klog.V(2).Infof("Making arbitrary choice between subnets in zone %q to attach to NLB (%q vs %q)", zone, scoredSubnets[0].subnet.Name, scoredSubnets[1].subnet.Name)
	}

	return scoredSubnets[0].subnet
}
