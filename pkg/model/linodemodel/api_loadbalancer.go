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
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	cloudlinode "k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

// APILoadBalancerModelBuilder builds a Linode (Akamai) load balancer for API access.
type APILoadBalancerModelBuilder struct {
	*LinodeModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &APILoadBalancerModelBuilder{}

func (b *APILoadBalancerModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	if len(b.Cluster.Spec.Networking.Subnets) == 0 || b.Cluster.Spec.Networking.Subnets[0].Region == "" {
		return fmt.Errorf("linode API load balancer requires at least one subnet with a region")
	}

	lb := &linodetasks.LoadBalancer{
		Name:      fi.PtrTo("api." + b.ClusterName()),
		Lifecycle: b.Lifecycle,
		Region:    fi.PtrTo(b.Cluster.Spec.Networking.Subnets[0].Region),
		Tags: []string{
			cloudlinode.BuildLinodeTag(kops.LabelClusterName, b.ClusterName()),
		},
		WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.KubeAPIServer},
	}

	if dns.IsGossipClusterName(b.Cluster.Name) || b.Cluster.UsesPrivateDNS() || b.Cluster.UsesNoneDNS() {
		lb.WellKnownServices = append(lb.WellKnownServices, wellknownservices.KopsController)
	}

	c.AddTask(lb)

	backendReconcile := &linodetasks.LoadBalancerBackends{
		Name:         fi.PtrTo("backends." + fi.ValueOf(lb.Name)),
		Lifecycle:    b.Lifecycle,
		LoadBalancer: lb,
	}
	c.AddTask(backendReconcile)

	return nil
}
