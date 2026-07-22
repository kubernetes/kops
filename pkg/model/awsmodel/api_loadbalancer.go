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

	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// APILoadBalancerBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerBuilder struct {
	*AWSModelContext

	Lifecycle         fi.Lifecycle
	SecurityLifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &APILoadBalancerBuilder{}

// Build is responsible for building the KubeAPI tasks for the aws model
func (b *APILoadBalancerBuilder) Build(c *fi.CloudupModelBuilderContext) error {
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

	var nlbSubnetMappings []*awstasks.SubnetMapping
	if len(lbSpec.Subnets) != 0 {
		// Subnets have been explicitly set
		for _, subnet := range lbSpec.Subnets {
			for _, clusterSubnet := range b.Cluster.Spec.Networking.Subnets {
				if subnet.Name == clusterSubnet.Name {
					nlbSubnetMapping := &awstasks.SubnetMapping{
						Subnet: b.LinkToSubnet(&clusterSubnet),
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

			nlbSubnetMappings = append(nlbSubnetMappings, &awstasks.SubnetMapping{Subnet: b.LinkToSubnet(subnet)})
		}
	}

	var nlb *awstasks.NetworkLoadBalancer
	{
		var nlbListeners []*awstasks.NetworkLoadBalancerListener

		if lbSpec.SSLCertificate == "" {
			listener443 := &awstasks.NetworkLoadBalancerListener{
				Name:                new(b.NLBListenerName("api", 443)),
				Lifecycle:           b.Lifecycle,
				NetworkLoadBalancer: b.LinkToNLB("api"),
				Port:                443,
				TargetGroup:         b.LinkToTargetGroup("tcp"),
			}
			nlbListeners = append(nlbListeners, listener443)
		} else {
			// When using a custom certificate, we create a secondary listener on 8443, which does _not_ use the custom certificate.
			// This is because client certificates cannot be used in conjunction with custom certificates on NLBs.
			listener8443 := &awstasks.NetworkLoadBalancerListener{
				Name:                new(b.NLBListenerName("api", 8443)),
				Lifecycle:           b.Lifecycle,
				NetworkLoadBalancer: b.LinkToNLB("api"),
				Port:                8443,
				TargetGroup:         b.LinkToTargetGroup("tcp"),
			}
			nlbListeners = append(nlbListeners, listener8443)

			// The primary listener _does_ use the custom certificate.
			listener443 := &awstasks.NetworkLoadBalancerListener{
				Name:                new(b.NLBListenerName("api", 443)),
				Lifecycle:           b.Lifecycle,
				NetworkLoadBalancer: b.LinkToNLB("api"),
				Port:                443,
				TargetGroup:         b.LinkToTargetGroup("tls"),
				SSLCertificateID:    lbSpec.SSLCertificate,
			}
			if lbSpec.SSLPolicy != nil {
				listener443.SSLPolicy = *lbSpec.SSLPolicy
			} else {
				listener443.SSLPolicy = "ELBSecurityPolicy-2016-08" // The AWS default
			}
			nlbListeners = append(nlbListeners, listener443)
		}

		if b.Cluster.UsesLoadBalancerForKopsController() {
			{
				nlbListener := &awstasks.NetworkLoadBalancerListener{
					Name:                new(b.NLBListenerName("api", wellknownports.KopsControllerPort)),
					Lifecycle:           b.Lifecycle,
					NetworkLoadBalancer: b.LinkToNLB("api"),
					Port:                wellknownports.KopsControllerPort,
					TargetGroup:         b.LinkToTargetGroup("kops-controller"),
				}
				nlbListeners = append(nlbListeners, nlbListener)
			}

			if b.Cluster.Spec.Networking.Cilium != nil && b.Cluster.Spec.Networking.Cilium.EtcdManaged {
				nlbListener := &awstasks.NetworkLoadBalancerListener{
					Name:                new(b.NLBListenerName("etcd-cilium", wellknownports.EtcdCiliumClientPort)),
					Lifecycle:           b.Lifecycle,
					NetworkLoadBalancer: b.LinkToNLB("api"),
					Port:                wellknownports.EtcdCiliumClientPort,
					TargetGroup:         b.LinkToTargetGroup("etcd-cilium"),
				}
				nlbListeners = append(nlbListeners, nlbListener)
			}
		}

		if lbSpec.SecurityGroupOverride != nil {
			klog.V(1).Infof("WARNING: You are overwriting the Load Balancers, Security Group. When this is done you are responsible for ensure the correct rules!")
		}

		tags := b.CloudTags("", false)
		for k, v := range b.Cluster.Spec.CloudLabels {
			tags[k] = v
		}
		// Override the returned name to be the expected ELB name
		tags["Name"] = "api." + b.ClusterName()

		nlb = &awstasks.NetworkLoadBalancer{
			Name:      new(b.NLBName("api")),
			Lifecycle: b.Lifecycle,

			LoadBalancerBaseName: new(b.LBName32("api")),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToELBSecurityGroup("api"),
			},
			SubnetMappings:    nlbSubnetMappings,
			Tags:              tags,
			WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.KubeAPIServer},
			VPC:               b.LinkToVPC(),
			Type:              elbv2types.LoadBalancerTypeEnumNetwork,
		}

		// Wait for all load balancer components to be created (including network interfaces needed to
		// bake NLB ENI IPs into worker nodeup configs). Limiting this to clusters that actually need
		// those IPs because load balancer creation is quite slow.
		if b.Cluster.UsesLoadBalancerForKopsController() {
			nlb.SetWaitForLoadBalancerReady(true)
		}

		if b.Cluster.UsesLoadBalancerForKopsController() {
			lbSpec.CrossZoneLoadBalancing = new(true)
		} else if lbSpec.CrossZoneLoadBalancing == nil {
			lbSpec.CrossZoneLoadBalancing = new(false)
		}

		nlb.CrossZoneLoadBalancing = lbSpec.CrossZoneLoadBalancing

		switch lbSpec.Type {
		case kops.LoadBalancerTypeInternal:
			nlb.Scheme = elbv2types.LoadBalancerSchemeEnumInternal
		case kops.LoadBalancerTypePublic:
			nlb.Scheme = elbv2types.LoadBalancerSchemeEnumInternetFacing
		default:
			return fmt.Errorf("unknown load balancer Type: %q", lbSpec.Type)
		}

		if lbSpec.AccessLog != nil {
			nlb.AccessLog = &awstasks.NetworkLoadBalancerAccessLog{
				Enabled:        new(true),
				S3BucketName:   lbSpec.AccessLog.Bucket,
				S3BucketPrefix: lbSpec.AccessLog.BucketPrefix,
			}
		} else {
			nlb.AccessLog = &awstasks.NetworkLoadBalancerAccessLog{
				Enabled: new(false),
			}
		}

		{
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
					Name:                new(groupName),
					Lifecycle:           b.Lifecycle,
					VPC:                 b.LinkToVPC(),
					Tags:                groupTags,
					Protocol:            elbv2types.ProtocolEnumTcp,
					Port:                new(int32(443)),
					Attributes:          groupAttrs,
					Interval:            new(int32(10)),
					HealthyThreshold:    new(int32(2)),
					UnhealthyThreshold:  new(int32(2)),
					HealthCheckProtocol: elbv2types.ProtocolEnumTcp,
					Shared:              new(false),
				}
				tg.CreateNewRevisionsWith(nlb)
				c.AddTask(tg)
			}

			if b.Cluster.UsesLoadBalancerForKopsController() {
				{
					groupName := b.NLBTargetGroupName("kops-controller")
					groupTags := b.CloudTags(groupName, false)

					// Override the returned name to be the expected NLB TG name
					groupTags["Name"] = groupName

					tg := &awstasks.TargetGroup{
						Name:                new(groupName),
						Lifecycle:           b.Lifecycle,
						VPC:                 b.LinkToVPC(),
						Tags:                groupTags,
						Protocol:            elbv2types.ProtocolEnumTcp,
						Port:                new(int32(wellknownports.KopsControllerPort)),
						Attributes:          groupAttrs,
						Interval:            new(int32(10)),
						HealthyThreshold:    new(int32(2)),
						UnhealthyThreshold:  new(int32(2)),
						HealthCheckProtocol: elbv2types.ProtocolEnumHttps,
						HealthCheckPath:     new("/healthz"),
						Shared:              new(false),
					}
					tg.CreateNewRevisionsWith(nlb)

					c.AddTask(tg)
				}

				if b.Cluster.Spec.Networking.Cilium != nil && b.Cluster.Spec.Networking.Cilium.EtcdManaged {
					groupName := b.NLBTargetGroupName("etcd-cilium")
					groupTags := b.CloudTags(groupName, false)

					// Override the returned name to be the expected NLB TG name
					groupTags["Name"] = groupName

					tg := &awstasks.TargetGroup{
						Name:                new(groupName),
						Lifecycle:           b.Lifecycle,
						VPC:                 b.LinkToVPC(),
						Tags:                groupTags,
						Protocol:            elbv2types.ProtocolEnumTcp,
						Port:                new(int32(wellknownports.EtcdCiliumClientPort)),
						Attributes:          groupAttrs,
						Interval:            new(int32(10)),
						HealthyThreshold:    new(int32(2)),
						UnhealthyThreshold:  new(int32(2)),
						HealthCheckProtocol: elbv2types.ProtocolEnumTcp,
						Shared:              new(false),
					}
					tg.CreateNewRevisionsWith(nlb)

					c.AddTask(tg)
				}
			}

			if lbSpec.SSLCertificate != "" {
				tlsGroupName := b.NLBTargetGroupName("tls")
				tlsGroupTags := b.CloudTags(tlsGroupName, false)

				// Override the returned name to be the expected NLB TG name
				tlsGroupTags["Name"] = tlsGroupName
				secondaryTG := &awstasks.TargetGroup{
					Name:                new(tlsGroupName),
					Lifecycle:           b.Lifecycle,
					VPC:                 b.LinkToVPC(),
					Tags:                tlsGroupTags,
					Protocol:            elbv2types.ProtocolEnumTls,
					Port:                new(int32(443)),
					Attributes:          groupAttrs,
					Interval:            new(int32(10)),
					HealthyThreshold:    new(int32(2)),
					UnhealthyThreshold:  new(int32(2)),
					HealthCheckProtocol: elbv2types.ProtocolEnumTcp,
					Shared:              new(false),
				}
				secondaryTG.CreateNewRevisionsWith(nlb)
				c.AddTask(secondaryTG)
			}
			for _, nlbListener := range nlbListeners {
				c.AddTask(nlbListener)
			}
			c.AddTask(nlb)
		}

	}

	var lbSG *awstasks.SecurityGroup
	{
		lbSG = &awstasks.SecurityGroup{
			Name:             new(b.ELBSecurityGroupName("api")),
			Lifecycle:        b.SecurityLifecycle,
			Description:      new("Security group for api ELB"),
			RemoveExtraRules: []string{"port=443"},
			VPC:              b.LinkToVPC(),
		}
		lbSG.Tags = b.CloudTags(*lbSG.Name, false)

		if lbSpec.SecurityGroupOverride != nil {
			lbSG.ID = new(*lbSpec.SecurityGroupOverride)
			lbSG.Shared = new(true)
		}

		c.AddTask(lbSG)
	}

	// Allow traffic from ELB to egress freely
	{
		{
			t := &awstasks.SecurityGroupRule{
				Name:          new("ipv4-api-elb-egress"),
				Lifecycle:     b.SecurityLifecycle,
				CIDR:          new("0.0.0.0/0"),
				Egress:        new(true),
				SecurityGroup: lbSG,
			}
			AddDirectionalGroupRule(c, t)
		}
		{
			t := &awstasks.SecurityGroupRule{
				Name:          new("ipv6-api-elb-egress"),
				Lifecycle:     b.SecurityLifecycle,
				IPv6CIDR:      new("::/0"),
				Egress:        new(true),
				SecurityGroup: lbSG,
			}
			AddDirectionalGroupRule(c, t)
		}
	}

	// Allow traffic into the ELB from KubernetesAPIAccess CIDRs
	{
		for _, cidr := range b.Cluster.Spec.API.Access {
			{
				t := &awstasks.SecurityGroupRule{
					Name:          new("https-api-elb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      new(int32(443)),
					Protocol:      new("tcp"),
					SecurityGroup: lbSG,
					ToPort:        new(int32(443)),
				}
				t.SetCidrOrPrefix(cidr)
				AddDirectionalGroupRule(c, t)
			}

			// If we have opened a secondary listener on 8443, allow it also
			if lbSpec.SSLCertificate != "" {
				t := &awstasks.SecurityGroupRule{
					Name:          new("https-api-elb-8443-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      new(int32(8443)),
					ToPort:        new(int32(8443)),
					Protocol:      new("tcp"),
					SecurityGroup: lbSG,
				}
				lbSG.RemoveExtraRules = append(lbSG.RemoveExtraRules, "port=8443")

				t.SetCidrOrPrefix(cidr)
				AddDirectionalGroupRule(c, t)
			}

			// Allow ICMP traffic required for PMTU discovery
			{
				t := &awstasks.SecurityGroupRule{
					Name:          new("icmpv6-pmtu-api-elb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      new(int32(-1)),
					Protocol:      new("icmpv6"),
					SecurityGroup: lbSG,
					ToPort:        new(int32(-1)),
				}
				t.SetCidrOrPrefix(cidr)
				if t.CIDR == nil {
					c.AddTask(t)
				}
			}
			{
				t := &awstasks.SecurityGroupRule{
					Name:          new("icmp-pmtu-api-elb-" + cidr),
					Lifecycle:     b.SecurityLifecycle,
					FromPort:      new(int32(3)),
					Protocol:      new("icmp"),
					SecurityGroup: lbSG,
					ToPort:        new(int32(4)),
				}
				t.SetCidrOrPrefix(cidr)
				if t.IPv6CIDR == nil {
					c.AddTask(t)
				}
			}
		}
	}

	if b.Cluster.UsesLoadBalancerForKopsController() {
		nodeGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleNode)
		if err != nil {
			return err
		}

		for _, nodeGroup := range nodeGroups {
			suffix := nodeGroup.Suffix
			t := &awstasks.SecurityGroupRule{
				Name:          new(fmt.Sprintf("node%s-to-elb", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				SecurityGroup: lbSG,
				SourceGroup:   nodeGroup.Task,
			}
			c.AddTask(t)
		}
	}

	masterGroups, err := b.GetSecurityGroups(kops.InstanceGroupRoleControlPlane)
	if err != nil {
		return err
	}

	if lbSpec.SSLCertificate != "" {
		for _, masterGroup := range masterGroups {
			suffix := masterGroup.Suffix
			// Allow access to control plane on secondary port through NLB
			t := &awstasks.SecurityGroupRule{
				Name:          new(fmt.Sprintf("tcp-api-cp%s", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      new(int32(8443)),
				Protocol:      new("tcp"),
				SecurityGroup: masterGroup.Task,
				SourceGroup:   lbSG,
				ToPort:        new(int32(8443)),
			}
			c.AddTask(t)
		}
	}

	// Add precreated additional security groups to the ELB
	{
		for _, id := range b.Cluster.Spec.API.LoadBalancer.AdditionalSecurityGroups {
			t := &awstasks.SecurityGroup{
				Name:      new(id),
				Lifecycle: b.SecurityLifecycle,
				ID:        new(id),
				Shared:    new(true),
			}
			c.EnsureTask(t)
			nlb.SecurityGroups = append(nlb.SecurityGroups, t)
		}
	}

	// Allow HTTPS to the control-plane instances from the ELB
	{
		for _, masterGroup := range masterGroups {
			suffix := masterGroup.Suffix
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          new(fmt.Sprintf("https-elb-to-master%s", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      new(int32(443)),
				Protocol:      new("tcp"),
				SecurityGroup: masterGroup.Task,
				SourceGroup:   lbSG,
				ToPort:        new(int32(443)),
			})
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          new(fmt.Sprintf("icmp-pmtu-elb-to-cp%s", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      new(int32(3)),
				Protocol:      new("icmp"),
				SecurityGroup: masterGroup.Task,
				SourceGroup:   lbSG,
				ToPort:        new(int32(4)),
			})
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          new(fmt.Sprintf("icmp-pmtu-cp%s-to-elb", suffix)),
				Lifecycle:     b.SecurityLifecycle,
				FromPort:      new(int32(3)),
				Protocol:      new("icmp"),
				SecurityGroup: lbSG,
				SourceGroup:   masterGroup.Task,
				ToPort:        new(int32(4)),
			})
			if b.Cluster.UsesLoadBalancerForKopsController() {
				{
					nlb.WellKnownServices = append(nlb.WellKnownServices, wellknownservices.KopsController)

					c.AddTask(&awstasks.SecurityGroupRule{
						Name:          new(fmt.Sprintf("kops-controller-elb-to-cp%s", suffix)),
						Lifecycle:     b.SecurityLifecycle,
						FromPort:      new(int32(wellknownports.KopsControllerPort)),
						Protocol:      new("tcp"),
						SecurityGroup: masterGroup.Task,
						ToPort:        new(int32(wellknownports.KopsControllerPort)),
						SourceGroup:   lbSG,
					})
				}

				if b.Cluster.Spec.Networking.Cilium != nil && b.Cluster.Spec.Networking.Cilium.EtcdManaged {
					nlb.WellKnownServices = append(nlb.WellKnownServices, wellknownservices.EtcdCilium)

					c.AddTask(&awstasks.SecurityGroupRule{
						Name:          new(fmt.Sprintf("etcd-cilium-elb-to-cp%s", suffix)),
						Lifecycle:     b.SecurityLifecycle,
						FromPort:      new(int32(wellknownports.EtcdCiliumClientPort)),
						Protocol:      new("tcp"),
						SecurityGroup: masterGroup.Task,
						ToPort:        new(int32(wellknownports.EtcdCiliumClientPort)),
						SourceGroup:   lbSG,
					})
				}
			}
		}
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
		return a[i].score >= a[j].score
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
