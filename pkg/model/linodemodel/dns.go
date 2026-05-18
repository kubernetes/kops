/*
Copyright 2026 The Kubernetes Authors.

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

package linodemodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

// DNSModelBuilder builds DNS tasks for the Linode (Akamai) API load balancer.
type DNSModelBuilder struct {
	*LinodeModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &DNSModelBuilder{}

// Build creates DNS tasks for the API load balancer
func (b *DNSModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.Cluster.PublishesDNSRecords() {
		return nil
	}

	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	lbTask, found := c.Tasks["LoadBalancer/api."+b.ClusterName()]
	if !found {
		return nil
	}
	apiLoadBalancer, ok := lbTask.(*linodetasks.LoadBalancer)
	if !ok {
		return nil
	}

	if b.Cluster.Spec.API.PublicName != "" {
		c.AddTask(&linodetasks.DNSRecord{
			Name:         fi.PtrTo(b.Cluster.Spec.API.PublicName),
			ResourceName: fi.PtrTo(b.Cluster.Spec.API.PublicName),
			Lifecycle:    b.Lifecycle,
			RecordType:   fi.PtrTo("A"),
			Target:       apiLoadBalancer,
		})
	}

	if b.UseLoadBalancerForInternalAPI() {
		c.AddTask(&linodetasks.DNSRecord{
			Name:         fi.PtrTo(b.Cluster.APIInternalName()),
			ResourceName: fi.PtrTo(b.Cluster.APIInternalName()),
			Lifecycle:    b.Lifecycle,
			RecordType:   fi.PtrTo("A"),
			Target:       apiLoadBalancer,
		})
	}

	return nil
}
