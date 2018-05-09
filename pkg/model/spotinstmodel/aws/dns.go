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

package aws

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/upup/pkg/fi"
	awstasks "k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks/aws"
)

// DNSModelBuilder builds DNS related model objects
type DNSModelBuilder struct {
	*ModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// Add a HostedZone if we are going to publish a dns record that depends on it
	if b.UsePrivateDNS() {
		// Check to see if we are using a bastion DNS record that points to the hosted zone
		// If we are, we need to make sure we include the hosted zone as a task

		if err := b.ensureDNSZone(c); err != nil {
			return err
		}
	} else {
		// We now create the DNS Zone for AWS even in the case of public zones;
		// it has to exist for the IAM record anyway.
		// TODO: We can now rationalize the code paths
		if !dns.IsGossipHostname(b.Cluster.Name) {
			if err := b.ensureDNSZone(c); err != nil {
				return err
			}
		}
	}

	if b.UseLoadBalancerForAPI() {
		// This will point our DNS to the load balancer, and put the pieces
		// together for kubectl to be work

		if !dns.IsGossipHostname(b.Cluster.Name) {
			if err := b.ensureDNSZone(c); err != nil {
				return err
			}

			apiDnsName := &awstasks.DNSName{
				Name:               fi.String(b.Cluster.Spec.MasterPublicName),
				Lifecycle:          b.Lifecycle,
				Zone:               b.LinkToDNSZone(),
				ResourceType:       fi.String("A"),
				TargetLoadBalancer: b.LinkToELB("api"),
			}
			c.AddTask(apiDnsName)
		}
	}

	if b.UsesBastionDns() {
		// Pulling this down into it's own if statement. The DNS configuration here
		// is similar to others, but I would like to keep it on it's own in case we need
		// to change anything.

		if err := b.ensureDNSZone(c); err != nil {
			return err
		}
	}

	return nil
}

func (b *DNSModelBuilder) ensureDNSZone(c *fi.ModelBuilderContext) error {
	if dns.IsGossipHostname(b.Cluster.Name) {
		return nil
	}

	// Configuration for a DNS zone
	dnsZone := &awstasks.DNSZone{
		Name:      fi.String(b.NameForDNSZone()),
		Lifecycle: b.Lifecycle,
	}

	topology := b.Cluster.Spec.Topology
	if topology != nil && topology.DNS != nil {
		switch topology.DNS.Type {
		case kops.DNSTypePublic:
			// Ignore

		case kops.DNSTypePrivate:
			dnsZone.Private = fi.Bool(true)
			dnsZone.PrivateVPC = b.LinkToVPC()

		default:
			return fmt.Errorf("Unknown DNS type %q", topology.DNS.Type)
		}
	}

	if !strings.Contains(b.Cluster.Spec.DNSZone, ".") {
		// Looks like a hosted zone ID
		dnsZone.ZoneID = fi.String(b.Cluster.Spec.DNSZone)
	} else {
		// Looks like a normal DNS name
		dnsZone.DNSName = fi.String(b.Cluster.Spec.DNSZone)
	}

	return c.EnsureTask(dnsZone)
}
