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
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// LoadBalancerDefaultIdleTimeout is the default idle time for the ELB
const LoadBalancerDefaultIdleTimeout = 5 * time.Minute

// APILoadBalancerBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerBuilder struct {
	*AWSModelContext

	Lifecycle         fi.Lifecycle
	SecurityLifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &APILoadBalancerBuilder{}

// Build is responsible for building the KubeAPI tasks for the aws model
func (b *APILoadBalancerBuilder) Build(c *fi.ModelBuilderContext) error {
	// Configuration where an ELB fronts the API
	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	lbSpec := b.Cluster.Spec.API.LoadBalancer
	if lbSpec == nil {
		// Skipping API ELB creation; not requested in Spec
		return nil
	}

	switch lbSpec.Type {
	case kops.LoadBalancerTypeInternal, kops.LoadBalancerTypePublic:
	// OK

	default:
		return fmt.Errorf("unhandled LoadBalancer type %q", lbSpec.Type)
	}

	var elbSubnets []*awstasks.Subnet
	var nlbSubnetMappings []*awstasks.SubnetMapping
	if len(lbSpec.Subnets) != 0 {
		// Subnets have been explicitly set
		for _, subnet := range lbSpec.Subnets {
			for _, clusterSubnet := range b.Cluster.Spec.Networking.Subnets {
				if subnet.Name == clusterSubnet.Name {
					elbSubnet := b.LinkToSubnet(&clusterSubnet)
					elbSubnets = append(elbSubnets, elbSubnet)

					nlbSubnetMapping := &awstasks.SubnetMapping{
						Subnet: elbSubnet,
					}
					if subnet.PrivateIPv4Address != nil {
						nlbSubnetMapping.PrivateIPv4Address = subnet.PrivateIPv4Address
					}
					if subnet.AllocationID != nil {
						nlbSubnetMapping.AllocationID = subnet.AllocationID
					}
					nlbSubnetMappings = append(nlbSubnetMappings, nlbSubnetMapping)
					break
				}
			}
		}
	} else {
		// Compute the subnets - only one per zone, and then break ties based on chooseBestSubnetForELB
		subnetsByZone := make(map[string][]*kops.ClusterSubnetSpec)
		for i := range b.Cluster.Spec.Networking.Subnets {
			subnet := &b.Cluster.Spec.Networking.Subnets[i]

			switch subnet.Type {
			case kops.SubnetTypePublic, kops.SubnetTypeUtility:
				if lbSpec.Type != kops.LoadBalancerTypePublic {
					continue
				}

			case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
				if lbSpec.Type != kops.LoadBalancerTypeInternal {
					continue
				}

			default:
				return fmt.Errorf("subnet %q had unknown type %q", subnet.Name, subnet.Type)
			}

			subnetsByZone[subnet.Zone] = append(subnetsByZone[subnet.Zone], subnet)
		}

		for zone, subnets := range subnetsByZone {
			subnet := b.chooseBestSubnetForELB(zone, subnets)

			elbSubnet := b.LinkToSubnet(subnet)
			elbSubnets = append(elbSubnets, elbSubnet)
			nlbSubnetMappings = append(nlbSubnetMappings, &awstasks.SubnetMapping{Subnet: elbSubnet})
		}
	}

	var clb *awstasks.ClassicLoadBalancer
	var nlb *awstasks.NetworkLoadBalancer
	{
		loadBalancerName := b.LBName32("api")

		idleTimeout := LoadBalancerDefaultIdleTimeout
		if lbSpec.IdleTimeoutSeconds != nil {
			idleTimeout = time.Second * time.Duration(*lbSpec.IdleTimeoutSeconds)
		}

		listeners := map[string]*awstasks.ClassicLoadBalancerListener{
			"443": {InstancePort: 443},
		}

		nlbListeners := []*awstasks.NetworkLoadBalancerListener{
			{
				Port:            443,
				TargetGroupName: b.NLBTargetGroupName("tcp"),
			},
		}
		if b.Cluster.UsesNoneDNS() {
			nlbListeners = append(nlbListeners, &awstasks.NetworkLoadBalancerListener{
				Port:            wellknownports.KopsControllerPort,
				TargetGroupName: b.NLBTargetGroupName("kops-controller"),
			})
		}

		if lbSpec.SSLCertificate != "" {
			listeners["443"].SSLCertificateID = lbSpec.SSLCertificate
			nlbListeners[0].Port = 8443

			nlbListener := &awstasks.NetworkLoadBalancerListener{
				Port:             443,
				TargetGroupName:  b.NLBTargetGroupName("tls"),
				SSLCertificateID: lbSpec.SSLCertificate,
			}
			if lbSpec.SSLPolicy != nil {
				nlbListener.SSLPolicy = *lbSpec.SSLPolicy
			}
			nlbListeners = append(nlbListeners, nlbListener)
		}

		if lbSpec.SecurityGroupOverride != nil {
			klog.V(1).Infof("WARNING: You are overwriting the Load Balancers, Security Group. When this is done you are responsible for ensure the correct rules!")
		}

		tags := b.CloudTags(loadBalancerName, false)
		for k, v := range b.Cluster.Spec.CloudLabels {
			tags[k] = v
		}
		// Override the returned name to be the expected ELB name
		tags["Name"] = "api." + b.ClusterName()

		name := b.NLBName("api")
		nlb = &awstasks.NetworkLoadBalancer{
			Name:      &name,
			Lifecycle: b.Lifecycle,

			LoadBalancerName: fi.PtrTo(loadBalancerName),
			CLBName:          fi.PtrTo("api." + b.ClusterName()),
			SubnetMappings:   nlbSubnetMappings,
			Listeners:        nlbListeners,
			TargetGroups:     make([]*awstasks.TargetGroup, 0),

			Tags:          tags,
			VPC:           b.LinkToVPC(),
			Type:          fi.PtrTo("network"),
			IpAddressType: fi.PtrTo("ipv4"),
		}
		if b.UseIPv6ForAPI() {
			nlb.IpAddressType = fi.PtrTo("dualstack")
		}

		clb = &awstasks.ClassicLoadBalancer{
			Name:      fi.PtrTo("api." + b.ClusterName()),
			Lifecycle: b.Lifecycle,

			LoadBalancerName: fi.PtrTo(loadBalancerName),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToELBSecurityGroup("api"),
			},
			Subnets:   elbSubnets,
			Listeners: listeners,

			// Configure fast-recovery health-checks
			HealthCheck: &awstasks.ClassicLoadBalancerHealthCheck{
				Target:             fi.PtrTo("SSL:443"),
				Timeout:            fi.PtrTo(int64(5)),
				Interval:           fi.PtrTo(int64(10)),
				HealthyThreshold:   fi.PtrTo(int64(2)),
				UnhealthyThreshold: fi.PtrTo(int64(2)),
			},

			ConnectionSettings: &awstasks.ClassicLoadBalancerConnectionSettings{
				IdleTimeout: fi.PtrTo(int64(idleTimeout.Seconds())),
			},

			ConnectionDraining: &awstasks.ClassicLoadBalancerConnectionDraining{
				Enabled: fi.PtrTo(true),
				Timeout: fi.PtrTo(int64(300)),
			},

			Tags: tags,
		}

		if b.Cluster.UsesNoneDNS() {
			lbSpec.CrossZoneLoadBalancing = fi.PtrTo(true)
		} else if lbSpec.CrossZoneLoadBalancing == nil {
			lbSpec.CrossZoneLoadBalancing = fi.PtrTo(false)
		}

		clb.CrossZoneLoadBalancing = &awstasks.ClassicLoadBalancerCrossZoneLoadBalancing{
			Enabled: lbSpec.CrossZoneLoadBalancing,
		}

		nlb.CrossZoneLoadBalancing = lbSpec.CrossZoneLoadBalancing

		switch lbSpec.Type {
		case kops.LoadBalancerTypeInternal:
			clb.Scheme = fi.PtrTo("internal")
			nlb.Scheme = fi.PtrTo("internal")
		case kops.LoadBalancerTypePublic:
			clb.Scheme = nil
			nlb.Scheme = nil
		default:
			return fmt.Errorf("unknown load balancer Type: %q", lbSpec.Type)
		}

		if lbSpec.AccessLog != nil {
			clb.AccessLog = &awstasks.ClassicLoadBalancerAccessLog{
				EmitInterval:   fi.PtrTo(int64(lbSpec.AccessLog.Interval)),
				Enabled:        fi.PtrTo(true),
				S3BucketName:   lbSpec.AccessLog.Bucket,
				S3BucketPrefix: lbSpec.AccessLog.BucketPrefix,
			}
			nlb.AccessLog = &awstasks.NetworkLoadBalancerAccessLog{
				Enabled:        fi.PtrTo(true),
				S3BucketName:   lbSpec.AccessLog.Bucket,
				S3BucketPrefix: lbSpec.AccessLog.BucketPrefix,
			}
		} else {
			clb.AccessLog = &awstasks.ClassicLoadBalancerAccessLog{
				Enabled: fi.PtrTo(false),
			}
			nlb.AccessLog = &awstasks.NetworkLoadBalancerAccessLog{
				Enabled: fi.PtrTo(false),
			}
		}

		if b.APILoadBalancerClass() == kops.LoadBalancerClassClassic {
			c.AddTask(clb)
		} else if b.APILoadBalancerClass() == kops.LoadBalancerClassNetwork {
			groupAttrs := map[string]string{
				awstasks.TargetGroupAttributeDeregistrationDelayConnectionTerminationEnabled: "true",
				awstasks.TargetGroupAttributeDeregistrationDelayTimeoutSeconds:               "30",
			}

			{
				groupName := b.NLBTargetGroupName("tcp")
				groupTags := b.CloudTags(groupName, false)

				// Override the returned name to be the expected NLB TG name
				groupTags["Name"] = groupName

				tg := &awstasks.TargetGroup{
					Name:               fi.PtrTo(groupName),
					Lifecycle:          b.Lifecycle,
					VPC:                b.LinkToVPC(),
					Tags:               groupTags,
					Protocol:           fi.PtrTo("TCP"),
					Port:               fi.PtrTo(int64(443)),
					Attributes:         groupAttrs,
					Interval:           fi.PtrTo(int64(10)),
					HealthyThreshold:   fi.PtrTo(int64(2)),
					UnhealthyThreshold: fi.PtrTo(int64(2)),
					Shared:             fi.PtrTo(false),
				}

				c.AddTask(tg)

				nlb.TargetGroups = append(nlb.TargetGroups, tg)
			}

			if b.Cluster.UsesNoneDNS() {
				groupName := b.NLBTargetGroupName("kops-controller")
				groupTags := b.CloudTags(groupName, false)

				// Override the returned name to be the expected NLB TG name
				groupTags["Name"] = groupName

				tg := &awstasks.TargetGroup{
					Name:               fi.PtrTo(groupName),
					Lifecycle:          b.Lifecycle,
					VPC:                b.LinkToVPC(),
					Tags:               groupTags,
					Protocol:           fi.PtrTo("TCP"),
					Port:               fi.PtrTo(int64(wellknownports.KopsControllerPort)),
					Attributes:         groupAttrs,
					Interval:           fi.PtrTo(int64(10)),
					HealthyThreshold:   fi.PtrTo(int64(2)),
					UnhealthyThreshold: fi.PtrTo(int64(2)),
					Shared:             fi.PtrTo(false),
				}

				c.AddTask(tg)

				nlb.TargetGroups = append(nlb.TargetGroups, tg)
			}

			if lbSpec.SSLCertificate != "" {
				tlsGroupName := b.NLBTargetGroupName("tls")
				tlsGroupTags := b.CloudTags(tlsGroupName, false)

				// Override the returned name to be the expected NLB TG name
				tlsGroupTags["Name"] = tlsGroupName
				secondaryTG := &awstasks.TargetGroup{
					Name:               fi.PtrTo(tlsGroupName),
					Lifecycle:          b.Lifecycle,
					VPC:                b.LinkToVPC(),
					Tags:               tlsGroupTags,
					Protocol:           fi.PtrTo("TLS"),
					Port:               fi.PtrTo(int64(443)),
					Attributes:         groupAttrs,
					Interval:           fi.PtrTo(int64(10)),
					HealthyThreshold:   fi.PtrTo(int64(2)),
					UnhealthyThreshold: fi.PtrTo(int64(2)),
					Shared:             fi.PtrTo(false),
				}
				c.AddTask(secondaryTG)
				nlb.TargetGroups = append(nlb.TargetGroups, secondaryTG)
			}
			sort.Stable(awstasks.OrderTargetGroupsByName(nlb.TargetGroups))
			c.AddTask(nlb)
		}

	}

	var lbSG *awstasks.SecurityGroup
	{
		lbSG = &awstasks.SecurityGroup{
			Name:             fi.PtrTo(b.ELBSecurityGroupName("api")),
			Lifecycle:        b.SecurityLifecycle,
			Description:      fi.PtrTo("Security group for api ELB"),
			RemoveExtraRules: []string{"port=443"},
			VPC:              b.LinkToVPC(),
		}
		lbSG.Tags = b.CloudTags(*lbSG.Name, false)

		if lbSpec.SecurityGroupOverride != nil {
			lbSG.ID = fi.PtrTo(*lbSpec.SecurityGroupOverride)
			lbSG.Shared = fi.PtrTo(true)
		}

		c.AddTask(lbSG)
	}

	// Allow traffic from ELB to egress freely
	if b.APILoadBalancerClass() == kops.LoadBalancerClassClassic {
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv4-api-elb-egress"),
				Lifecycle:     b.SecurityLifecycle,
				CIDR:          fi.PtrTo("0.0.0.0/0"),
				Egress:        fi.PtrTo(true),
				SecurityGroup: lbSG,
			}
			AddDirectionalGroupRule(c, t)
		}
		{
			t := &awstasks.SecurityGroupRule{
				Name:          fi.PtrTo("ipv6-api-elb-egress"),
				Lifecycle:     b.SecurityLifecycle,
				IPv6CIDR:      fi.PtrTo("::/0"),
				Egress:        fi.PtrTo(true),
				SecurityGroup: lbSG,
			}
			AddDirectionalGroupRule(c, t)
		}
	}

	// Allow traffic into the ELB from KubernetesAPIAccess CIDRs
	if b.APILoadBalancerClass() == kops.LoadBalancerClassClassic {
		for _, cidr := range b.Cluster.Spec.API.Access {
			{
				t := &awstasks.SecurityGroupRule{
					Name:          fi.PtrTo("https-api-elb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      fi.PtrTo(int64(443)),
					Protocol:      fi.PtrTo("tcp"),
					SecurityGroup: lbSG,
					ToPort:        fi.PtrTo(int64(443)),
				}
				t.SetCidrOrPrefix(cidr)
				AddDirectionalGroupRule(c, t)
			}

			// Allow ICMP traffic required for PMTU discovery
			if utils.IsIPv6CIDR(cidr) {
				c.AddTask(&awstasks.SecurityGroupRule{
					Name:          fi.PtrTo("icmpv6-pmtu-api-elb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					IPv6CIDR:      fi.PtrTo(cidr),
					FromPort:      fi.PtrTo(int64(-1)),
					Protocol:      fi.PtrTo("icmpv6"),
					SecurityGroup: lbSG,
					ToPort:        fi.PtrTo(int64(-1)),
				})
			} else {
				c.AddTask(&awstasks.SecurityGroupRule{
					Name:          fi.PtrTo("icmp-pmtu-api-elb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					CIDR:          fi.PtrTo(cidr),
					FromPort:      fi.PtrTo(int64(3)),
					Protocol:      fi.PtrTo("icmp"),
					SecurityGroup: lbSG,
					ToPort:        fi.PtrTo(int64(4)),
				})
			}
		}
	}

	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleControlPlane)
	if err != nil {
		return err
	}

	if b.APILoadBalancerClass() == kops.LoadBalancerClassNetwork {
		for _, cidr := range b.Cluster.Spec.API.Access {
			for _, masterGroup := range masterGroups {
				{
					t := &awstasks.SecurityGroupRule{
						Name:          fi.PtrTo(fmt.Sprintf("https-api-elb-%s", cidr)),
						Lifecycle:     b.SecurityLifecycle,
						FromPort:      fi.PtrTo(int64(443)),
						Protocol:      fi.PtrTo("tcp"),
						SecurityGroup: masterGroup.Task,
						ToPort:        fi.PtrTo(int64(443)),
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
						Name:          fi.PtrTo("icmpv6-pmtu-api-elb-" + cidr),
						Lifecycle:     b.SecurityLifecycle,
						FromPort:      fi.PtrTo(int64(-1)),
						Protocol:      fi.PtrTo("icmpv6"),
						SecurityGroup: masterGroup.Task,
						ToPort:        fi.PtrTo(int64(-1)),
					}
					t.SetCidrOrPrefix(cidr)
					c.AddTask(t)
				} else {
					t := &awstasks.SecurityGroupRule{
						Name:          fi.PtrTo("icmp-pmtu-api-elb-" + cidr),
						Lifecycle:     b.SecurityLifecycle,
						FromPort:      fi.PtrTo(int64(3)),
						Protocol:      fi.PtrTo("icmp"),
						SecurityGroup: masterGroup.Task,
						ToPort:        fi.PtrTo(int64(4)),
					}
					t.SetCidrOrPrefix(cidr)
					c.AddTask(t)
				}

				if b.Cluster.Spec.API.LoadBalancer != nil && b.Cluster.Spec.API.LoadBalancer.SSLCertificate != "" {
					// Allow access to masters on secondary port through NLB
					t := &awstasks.SecurityGroupRule{
						Name:          fi.PtrTo(fmt.Sprintf("tcp-api-%s", cidr)),
						Lifecycle:     b.SecurityLifecycle,
						FromPort:      fi.PtrTo(int64(8443)),
						Protocol:      fi.PtrTo("tcp"),
						SecurityGroup: masterGroup.Task,
						ToPort:        fi.PtrTo(int64(8443)),
					}
					t.SetCidrOrPrefix(cidr)
					c.AddTask(t)
				}
			}
		}
	}

	// Add precreated additional security groups to the ELB
	if b.APILoadBalancerClass() == kops.LoadBalancerClassClassic {
		for _, id := range b.Cluster.Spec.API.LoadBalancer.AdditionalSecurityGroups {
			t := &awstasks.SecurityGroup{
				Name:      fi.PtrTo(id),
				Lifecycle: b.SecurityLifecycle,
				ID:        fi.PtrTo(id),
				Shared:    fi.PtrTo(true),
			}
			if err := c.EnsureTask(t); err != nil {
				return err
			}
			clb.SecurityGroups = append(clb.SecurityGroups, t)
		}
	}

	// Allow HTTPS to the master instances from the ELB
	if b.APILoadBalancerClass() == kops.LoadBalancerClassClassic {
		for _, masterGroup := range masterGroups {
			suffix := masterGroup.Suffix
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          fi.PtrTo(fmt.Sprintf("https-elb-to-master%s", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      fi.PtrTo(int64(443)),
				Protocol:      fi.PtrTo("tcp"),
				SecurityGroup: masterGroup.Task,
				SourceGroup:   lbSG,
				ToPort:        fi.PtrTo(int64(443)),
			})
		}
	} else if b.APILoadBalancerClass() == kops.LoadBalancerClassNetwork {
		for _, masterGroup := range masterGroups {
			suffix := masterGroup.Suffix
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          fi.PtrTo(fmt.Sprintf("https-elb-to-master%s", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      fi.PtrTo(int64(443)),
				Protocol:      fi.PtrTo("tcp"),
				SecurityGroup: masterGroup.Task,
				ToPort:        fi.PtrTo(int64(443)),
				CIDR:          fi.PtrTo(b.Cluster.Spec.Networking.NetworkCIDR),
			})
			for _, cidr := range b.Cluster.Spec.Networking.AdditionalNetworkCIDRs {
				c.AddTask(&awstasks.SecurityGroupRule{
					Name:          fi.PtrTo(fmt.Sprintf("https-lb-to-master%s-%s", suffix, cidr)),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      fi.PtrTo(int64(443)),
					Protocol:      fi.PtrTo("tcp"),
					SecurityGroup: masterGroup.Task,
					ToPort:        fi.PtrTo(int64(443)),
					CIDR:          fi.PtrTo(cidr),
				})
			}
		}
	}

	// Allow kops-controller to the master instances from the ELB
	if b.Cluster.UsesNoneDNS() && b.APILoadBalancerClass() == kops.LoadBalancerClassNetwork {
		for _, masterGroup := range masterGroups {
			suffix := masterGroup.Suffix
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          fi.PtrTo(fmt.Sprintf("kops-controller-lb-to-master%s", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      fi.PtrTo(int64(wellknownports.KopsControllerPort)),
				Protocol:      fi.PtrTo("tcp"),
				SecurityGroup: masterGroup.Task,
				ToPort:        fi.PtrTo(int64(wellknownports.KopsControllerPort)),
				CIDR:          fi.PtrTo(b.Cluster.Spec.Networking.NetworkCIDR),
			})
			for _, cidr := range b.Cluster.Spec.Networking.AdditionalNetworkCIDRs {
				c.AddTask(&awstasks.SecurityGroupRule{
					Name:          fi.PtrTo(fmt.Sprintf("kops-controller-lb-to-master%s-%s", suffix, cidr)),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      fi.PtrTo(int64(wellknownports.KopsControllerPort)),
					Protocol:      fi.PtrTo("tcp"),
					SecurityGroup: masterGroup.Task,
					ToPort:        fi.PtrTo(int64(wellknownports.KopsControllerPort)),
					CIDR:          fi.PtrTo(cidr),
				})
			}
		}
	}

	if b.Cluster.IsGossip() || b.Cluster.UsesPrivateDNS() || b.Cluster.UsesNoneDNS() {
		// Ensure the LB hostname is included in the TLS certificate,
		// if we're not going to use an alias for it
		clb.ForAPIServer = true
		nlb.ForAPIServer = true
	}

	return nil
}

type scoredSubnet struct {
	score  int
	subnet *kops.ClusterSubnetSpec
}

type ByScoreDescending []*scoredSubnet

func (a ByScoreDescending) Len() int      { return len(a) }
func (a ByScoreDescending) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByScoreDescending) Less(i, j int) bool {
	if a[i].score != a[j].score {
		// ! to sort highest score first
		return !(a[i].score < a[j].score)
	}
	// Use name to break ties consistently
	return a[i].subnet.Name < a[j].subnet.Name
}

// Choose between subnets in a zone.
// We have already applied the rules to match internal subnets to internal ELBs and vice-versa for public-facing ELBs.
// For internal ELBs: we prefer dual stack and the master subnets
// For public facing ELBs: we prefer the utility subnets
func (b *APILoadBalancerBuilder) chooseBestSubnetForELB(zone string, subnets []*kops.ClusterSubnetSpec) *kops.ClusterSubnetSpec {
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
		klog.V(2).Infof("Making arbitrary choice between subnets in zone %q to attach to ELB (%q vs %q)", zone, scoredSubnets[0].subnet.Name, scoredSubnets[1].subnet.Name)
	}

	return scoredSubnets[0].subnet
}
