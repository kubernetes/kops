package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// DNSModelBuilder builds DNS related model objects
type DNSModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.UseLoadBalancerForAPI() {
		// This will point our DNS to the load balancer, and put the pieces
		// together for kubectl to be work

		// Configuration for a DNS name for the master
		dnsZone := &awstasks.DNSZone{
			Name: s(b.Cluster.Spec.DNSZone),
		}
		c.AddTask(dnsZone)

		dnsName := &awstasks.DNSName{
			Name:               s(b.Cluster.Spec.MasterPublicName),
			Zone:               dnsZone,
			ResourceType:       s("A"),
			TargetLoadBalancer: b.LinkToELB("api"),
		}
		c.AddTask(dnsName)
	}
	return nil
}
