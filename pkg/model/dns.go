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
