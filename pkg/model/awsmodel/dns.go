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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// DNSModelBuilder builds DNS related model objects
type DNSModelBuilder struct {
	*AWSModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) ensureDNSZone(c *fi.ModelBuilderContext) error {
	if b.Cluster.IsGossip() || b.Cluster.UsesNoneDNS() {
		return nil
	}

	// Configuration for a DNS zone
	dnsZone := &awstasks.DNSZone{
		Name:      fi.PtrTo(b.NameForDNSZone()),
		Lifecycle: b.Lifecycle,
	}

	topology := b.Cluster.Spec.Networking.Topology
	if topology != nil {
		switch topology.DNS {
		case kops.DNSTypePublic:
		// Ignore

		case kops.DNSTypePrivate:
			dnsZone.Private = fi.PtrTo(true)
			dnsZone.PrivateVPC = b.LinkToVPC()

		default:
			return fmt.Errorf("unknown DNS type %q", topology.DNS)
		}
	}

	if !strings.Contains(b.Cluster.Spec.DNSZone, ".") {
		// Looks like a hosted zone ID
		dnsZone.ZoneID = fi.PtrTo(b.Cluster.Spec.DNSZone)
	} else {
		// Looks like a normal DNS name
		dnsZone.DNSName = fi.PtrTo(b.Cluster.Spec.DNSZone)
	}

	return c.EnsureTask(dnsZone)
}

func (b *DNSModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// Add a HostedZone if we are going to publish a dns record that depends on it
	if !b.Cluster.IsGossip() && !b.Cluster.UsesNoneDNS() {
		if err := b.ensureDNSZone(c); err != nil {
			return err
		}
	}

	var targetLoadBalancer awstasks.DNSTarget

	if b.UseLoadBalancerForAPI() || b.UseLoadBalancerForInternalAPI() {
		lbSpec := b.Cluster.Spec.API.LoadBalancer
		switch lbSpec.Class {
		case kops.LoadBalancerClassClassic, "":
			targetLoadBalancer = awstasks.DNSTarget(b.LinkToCLB("api"))
		case kops.LoadBalancerClassNetwork:
			targetLoadBalancer = awstasks.DNSTarget(b.LinkToNLB("api"))
		}
	}

	if b.UseLoadBalancerForAPI() {
		// This will point our external DNS record to the load balancer, and put the
		// pieces together for kubectl to work

		if !b.Cluster.IsGossip() && !b.Cluster.UsesNoneDNS() {
			if err := b.ensureDNSZone(c); err != nil {
				return err
			}

			c.AddTask(&awstasks.DNSName{
				Name:               fi.PtrTo(b.Cluster.Spec.API.PublicName),
				ResourceName:       fi.PtrTo(b.Cluster.Spec.API.PublicName),
				Lifecycle:          b.Lifecycle,
				Zone:               b.LinkToDNSZone(),
				ResourceType:       fi.PtrTo("A"),
				TargetLoadBalancer: targetLoadBalancer,
			})
			if b.UseIPv6ForAPI() {
				c.AddTask(&awstasks.DNSName{
					Name:               fi.PtrTo(b.Cluster.Spec.API.PublicName + "-AAAA"),
					ResourceName:       fi.PtrTo(b.Cluster.Spec.API.PublicName),
					Lifecycle:          b.Lifecycle,
					Zone:               b.LinkToDNSZone(),
					ResourceType:       fi.PtrTo("AAAA"),
					TargetLoadBalancer: targetLoadBalancer,
				})
			}
		}
	}

	if b.UseLoadBalancerForInternalAPI() {
		// This will point the internal API DNS record to the load balancer.
		// This means kubelet connections go via the load balancer and are more HA.

		if !b.Cluster.IsGossip() && !b.Cluster.UsesNoneDNS() {
			if err := b.ensureDNSZone(c); err != nil {
				return err
			}

			// Using EnsureTask as APIInternalName() and APIPublicName could be the same
			{
				err := c.EnsureTask(&awstasks.DNSName{
					Name:               fi.PtrTo(b.Cluster.APIInternalName()),
					ResourceName:       fi.PtrTo(b.Cluster.APIInternalName()),
					Lifecycle:          b.Lifecycle,
					Zone:               b.LinkToDNSZone(),
					ResourceType:       fi.PtrTo("A"),
					TargetLoadBalancer: targetLoadBalancer,
				})
				if err != nil {
					return err
				}
			}
			if b.UseIPv6ForAPI() {
				err := c.EnsureTask(&awstasks.DNSName{
					Name:               fi.PtrTo(b.Cluster.APIInternalName() + "-AAAA"),
					ResourceName:       fi.PtrTo(b.Cluster.APIInternalName()),
					Lifecycle:          b.Lifecycle,
					Zone:               b.LinkToDNSZone(),
					ResourceType:       fi.PtrTo("AAAA"),
					TargetLoadBalancer: targetLoadBalancer,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	if b.UsesBastionDns() {
		// Pulling this down into it's own if statement. The DNS configuration here
		// is similar to others, but I would like to keep it on it's own in case we need
		// to change anything.

		if !b.Cluster.IsGossip() && !b.Cluster.UsesNoneDNS() {
			if err := b.ensureDNSZone(c); err != nil {
				return err
			}
		}
	}

	return nil
}
