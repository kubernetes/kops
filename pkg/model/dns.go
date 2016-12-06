package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"fmt"
)

// DNSModelBuilder builds DNS related model objects
type DNSModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.ModelBuilderContext) error {
	//
	//# MASTER_DNS
	//
	//# Configuration for a DNS name for the master
	//
	//dnsZone/{{ .DNSZone }}: {}
	//
	//
	//# MASTER_LB
	//
	//
	//# Master name -> ELB
	//dnsName/{{ .MasterPublicName }}:
	//Zone: dnsZone/{{ .DNSZone }}
	//ResourceType: "A"
	//TargetLoadBalancer: loadBalancer/api.{{ ClusterName }}
	//
	//
	//# PRIVATE TOPOLOGY
	//
	//# ---------------------------------------------------------------
	//# DNS - Api
	//#
	//# This will point our DNS to the load balancer, and put the pieces
	//# together for kubectl to be work
	//# ---------------------------------------------------------------
	//dnsZone/{{ .DNSZone }}: {}
	//dnsName/{{ .MasterPublicName }}:
	//Zone: dnsZone/{{ .DNSZone }}
	//ResourceType: "A"
	//TargetLoadBalancer: loadBalancer/api.{{ ClusterName }}

	return fmt.Errorf("dns.go NOT IMPLEMENTED")
}
