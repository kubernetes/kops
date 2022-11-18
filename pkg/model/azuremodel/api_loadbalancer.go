/*
Copyright 2020 The Kubernetes Authors.

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

package azuremodel

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// APILoadBalancerModelBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerModelBuilder struct {
	*AzureModelContext

	Lifecycle         fi.Lifecycle
	SecurityLifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &APILoadBalancerModelBuilder{}

// Build builds tasks for creating a K8s API server for Azure.
func (b *APILoadBalancerModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	lbSpec := b.Cluster.Spec.API.LoadBalancer
	if lbSpec == nil {
		// Skipping API LB creation; not requested in Spec
		return nil
	}

	// Create LoadBalancer for API ELB
	lb := &azuretasks.LoadBalancer{
		Name:          fi.PtrTo(b.NameForLoadBalancer()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
	}

	switch lbSpec.Type {
	case kops.LoadBalancerTypeInternal:
		lb.External = to.BoolPtr(false)
		subnet, err := b.subnetForLoadBalancer()
		if err != nil {
			return err
		}
		lb.Subnet = b.LinkToAzureSubnet(subnet)
	case kops.LoadBalancerTypePublic:
		lb.External = to.BoolPtr(true)

		// Create Public IP Address for Public Loadbalacer
		p := &azuretasks.PublicIPAddress{
			Name:          fi.PtrTo(b.NameForLoadBalancer()),
			Lifecycle:     b.Lifecycle,
			ResourceGroup: b.LinkToResourceGroup(),
			Tags:          map[string]*string{},
		}
		c.AddTask(p)
	default:
		return fmt.Errorf("unknown load balancer Type: %q", lbSpec.Type)
	}

	c.AddTask(lb)

	if b.Cluster.IsGossip() || b.Cluster.UsesPrivateDNS() || b.Cluster.UsesNoneDNS() {
		lb.ForAPIServer = true
	}

	return nil
}

// subnetForLoadBalancer returns the subnet the loadbalancer will use.
func (c *AzureModelContext) subnetForLoadBalancer() (*kops.ClusterSubnetSpec, error) {
	// Get all master instance group subnets
	for _, ig := range c.MasterInstanceGroups() {
		subnets, err := c.GatherSubnets(ig)
		if err != nil {
			return nil, err
		}
		if len(subnets) != 1 {
			return nil, fmt.Errorf("expected exactly one subnet for InstanceGroup %q; subnets was %s", ig.Name, ig.Spec.Subnets)
		}
		if subnets[0].Type != kops.SubnetTypeDualStack && subnets[0].Type != kops.SubnetTypePrivate {
			continue
		}
		return subnets[0], nil
	}

	return nil, fmt.Errorf("no suitable subnets found")
}
