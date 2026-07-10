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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// APILoadBalancerModelBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerModelBuilder struct {
	*AzureModelContext

	Lifecycle         fi.Lifecycle
	SecurityLifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &APILoadBalancerModelBuilder{}

// Build builds tasks for creating a K8s API server for Azure.
func (b *APILoadBalancerModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
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
		Name:              new(b.NameForLoadBalancer()),
		Lifecycle:         b.Lifecycle,
		ResourceGroup:     b.LinkToResourceGroup(),
		Tags:              map[string]*string{},
		SKU:               network.LoadBalancerSKUNameStandard,
		WellKnownServices: []wellknownservices.WellKnownService{wellknownservices.KubeAPIServer},
	}

	// API server probe: TCP on 443
	lb.Probes = append(lb.Probes, azuretasks.LoadBalancerProbe{
		Name:              fmt.Sprintf("Health-TCP-%d", wellknownports.KubeAPIServer),
		Protocol:          network.ProbeProtocolTCP,
		Port:              wellknownports.KubeAPIServer,
		IntervalInSeconds: 15,
		NumberOfProbes:    4,
	})
	lb.Rules = append(lb.Rules, azuretasks.LoadBalancerRule{
		Name:                 fmt.Sprintf("TCP-%d", wellknownports.KubeAPIServer),
		Port:                 wellknownports.KubeAPIServer,
		ProbeName:            fmt.Sprintf("Health-TCP-%d", wellknownports.KubeAPIServer),
		Protocol:             network.TransportProtocolTCP,
		IdleTimeoutInMinutes: 4,
		LoadDistribution:     network.LoadDistributionDefault,
	})

	switch lbSpec.Type {
	case kops.LoadBalancerTypeInternal:
		lb.External = to.Ptr(false)
		subnet, err := b.subnetForLoadBalancer()
		if err != nil {
			return err
		}
		lb.Subnet = b.LinkToAzureSubnet(subnet)
	case kops.LoadBalancerTypePublic:
		lb.External = to.Ptr(true)

		// Create Public IP Address for Public Loadbalacer
		p := &azuretasks.PublicIPAddress{
			Name:             new(b.NameForLoadBalancer()),
			Lifecycle:        b.Lifecycle,
			ResourceGroup:    b.LinkToResourceGroup(),
			IPVersion:        network.IPVersionIPv4,
			AllocationMethod: network.IPAllocationMethodStatic,
			SKU:              network.PublicIPAddressSKUNameStandard,
			Tags:             map[string]*string{},
		}
		c.AddTask(p)
		lb.PublicIPAddress = p
	default:
		return fmt.Errorf("unknown load balancer Type: %q", lbSpec.Type)
	}

	if b.Cluster.UsesLoadBalancerForKopsController() {
		lb.WellKnownServices = append(lb.WellKnownServices, wellknownservices.KopsController)

		// kops-controller probe: HTTPS on 3988 with /healthz
		lb.Probes = append(lb.Probes, azuretasks.LoadBalancerProbe{
			Name:              fmt.Sprintf("Health-HTTPS-%d", wellknownports.KopsControllerPort),
			Protocol:          network.ProbeProtocolHTTPS,
			Port:              wellknownports.KopsControllerPort,
			RequestPath:       new("/healthz"),
			IntervalInSeconds: 15,
			NumberOfProbes:    4,
		})
		lb.Rules = append(lb.Rules, azuretasks.LoadBalancerRule{
			Name:                 fmt.Sprintf("TCP-%d", wellknownports.KopsControllerPort),
			Port:                 wellknownports.KopsControllerPort,
			ProbeName:            fmt.Sprintf("Health-HTTPS-%d", wellknownports.KopsControllerPort),
			Protocol:             network.TransportProtocolTCP,
			IdleTimeoutInMinutes: 4,
			LoadDistribution:     network.LoadDistributionDefault,
		})
	}

	c.AddTask(lb)

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
