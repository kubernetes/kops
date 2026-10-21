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

package azuretasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// LoadBalancerProbe defines a health probe for the load balancer.
type LoadBalancerProbe struct {
	// Name is the probe name, e.g. "Health-HTTPS-3988".
	Name string
	// Protocol is the probe protocol, e.g. network.ProbeProtocolTCP.
	Protocol network.ProbeProtocol
	// Port is the port to probe.
	Port int32
	// RequestPath is the path for HTTP/HTTPS probes (nil for TCP).
	RequestPath *string
	// IntervalInSeconds is the probe interval.
	IntervalInSeconds int32
	// NumberOfProbes is the number of probes before marking unhealthy.
	NumberOfProbes int32
}

var _ fi.CloudupHasDependencies = (*LoadBalancerProbe)(nil)

func (e *LoadBalancerProbe) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

// LoadBalancerRule defines a load balancing rule.
type LoadBalancerRule struct {
	// Name is the rule name, e.g. "TCP-443".
	Name string
	// Port is both the frontend and backend port.
	Port int32
	// ProbeName references the probe by name.
	ProbeName string
	// Protocol is the transport protocol, e.g. network.TransportProtocolTCP.
	Protocol network.TransportProtocol
	// IdleTimeoutInMinutes is the idle timeout for the rule.
	IdleTimeoutInMinutes int32
	// EnableFloatingIP enables Direct Server Return.
	EnableFloatingIP bool
	// LoadDistribution is the load distribution policy.
	LoadDistribution network.LoadDistribution
}

var _ fi.CloudupHasDependencies = (*LoadBalancerRule)(nil)

func (e *LoadBalancerRule) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

// LoadBalancer is an Azure Cloud LoadBalancer
// +kops:fitask
type LoadBalancer struct {
	Name          *string
	Lifecycle     fi.Lifecycle
	ResourceGroup *ResourceGroup
	Subnet        *Subnet

	// External is set to true when the loadbalancer is used for external traffic
	External *bool

	// PublicIPAddress is the public IP address for external load balancers.
	PublicIPAddress *PublicIPAddress

	Tags map[string]*string

	// SKU is the load balancer SKU, e.g. network.LoadBalancerSKUNameStandard.
	SKU network.LoadBalancerSKUName

	// WellKnownServices indicates which services are supported by this resource.
	// This field is internal and is not rendered to the cloud.
	WellKnownServices []wellknownservices.WellKnownService

	// Probes defines the health probes for the load balancer.
	Probes []LoadBalancerProbe

	// Rules defines the load balancing rules.
	Rules []LoadBalancerRule
}

var (
	_ fi.CloudupTask          = &LoadBalancer{}
	_ fi.CompareWithID        = &LoadBalancer{}
	_ fi.CloudupTaskNormalize = &LoadBalancer{}
	_ fi.HasAddress           = &LoadBalancer{}
)

// CompareWithID returns the Name of the LoadBalancer
func (lb *LoadBalancer) CompareWithID() *string {
	return lb.Name
}

// GetWellKnownServices implements fi.HasAddress::GetWellKnownServices.
// It indicates which services we support with this load balancer.
func (lb *LoadBalancer) GetWellKnownServices() []wellknownservices.WellKnownService {
	return lb.WellKnownServices
}

func (lb *LoadBalancer) FindAddresses(c *fi.CloudupContext) ([]string, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	loadbalancer, err := cloud.LoadBalancer().Get(context.TODO(), *lb.ResourceGroup.Name, *lb.Name)
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return nil, err
	}

	if loadbalancer != nil && loadbalancer.Properties != nil && loadbalancer.Properties.FrontendIPConfigurations != nil && len(loadbalancer.Properties.FrontendIPConfigurations) > 0 {
		var addresses []string
		for _, fipc := range loadbalancer.Properties.FrontendIPConfigurations {
			if fipc.Properties == nil {
				continue
			}
			if fipc.Properties.PrivateIPAddress != nil {
				addresses = append(addresses, *fipc.Properties.PrivateIPAddress)
			}
			if fipc.Properties.PublicIPAddress != nil && fipc.Properties.PublicIPAddress.Properties != nil && fipc.Properties.PublicIPAddress.Properties.IPAddress != nil {
				addresses = append(addresses, *fipc.Properties.PublicIPAddress.Properties.IPAddress)
			}
		}
		return addresses, nil
	}

	return nil, nil
}

// Find discovers the LoadBalancer in the cloud provider
func (lb *LoadBalancer) Find(c *fi.CloudupContext) (*LoadBalancer, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	l, err := cloud.LoadBalancer().List(context.TODO(), *lb.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.LoadBalancer
	for _, v := range l {
		if *v.Name == *lb.Name {
			found = v
			break
		}
	}
	if found == nil {
		return nil, nil
	}
	if found.Properties != nil && found.Properties.ProvisioningState != nil && *found.Properties.ProvisioningState == network.ProvisioningStateFailed {
		klog.Warningf("found load balancer %q in failed provisioning state", *lb.Name)
		return nil, nil
	}

	lbProperties := found.Properties

	feConfigs := lbProperties.FrontendIPConfigurations
	if len(feConfigs) != 1 {
		return nil, fmt.Errorf("unexpected number of frontend configs found for LoadBalancer %s: %d", *lb.Name, len(feConfigs))
	}
	feConfig := feConfigs[0]
	subnet := feConfig.Properties.Subnet

	actual := &LoadBalancer{
		Name:              lb.Name,
		Lifecycle:         lb.Lifecycle,
		WellKnownServices: lb.WellKnownServices,
		ResourceGroup: &ResourceGroup{
			Name: lb.ResourceGroup.Name,
		},
		External: to.Ptr(feConfig.Properties.PublicIPAddress != nil),
		Tags:     found.Tags,
	}
	if found.SKU != nil {
		actual.SKU = fi.ValueOf(found.SKU.Name)
	}
	if subnet != nil {
		actual.Subnet = &Subnet{
			Name: subnet.Name,
		}
	}
	if feConfig.Properties.PublicIPAddress != nil {
		actual.PublicIPAddress = &PublicIPAddress{
			ID: feConfig.Properties.PublicIPAddress.ID,
		}
	}

	for _, probe := range lbProperties.Probes {
		if probe.Properties == nil {
			continue
		}
		p := LoadBalancerProbe{
			Name:              fi.ValueOf(probe.Name),
			Protocol:          fi.ValueOf(probe.Properties.Protocol),
			Port:              fi.ValueOf(probe.Properties.Port),
			IntervalInSeconds: fi.ValueOf(probe.Properties.IntervalInSeconds),
			NumberOfProbes:    fi.ValueOf(probe.Properties.NumberOfProbes),
		}
		if probe.Properties.RequestPath != nil {
			p.RequestPath = probe.Properties.RequestPath
		}
		actual.Probes = append(actual.Probes, p)
	}

	for _, rule := range lbProperties.LoadBalancingRules {
		if rule.Properties == nil {
			continue
		}
		r := LoadBalancerRule{
			Name:                 fi.ValueOf(rule.Name),
			Port:                 fi.ValueOf(rule.Properties.FrontendPort),
			Protocol:             fi.ValueOf(rule.Properties.Protocol),
			IdleTimeoutInMinutes: fi.ValueOf(rule.Properties.IdleTimeoutInMinutes),
			EnableFloatingIP:     fi.ValueOf(rule.Properties.EnableFloatingIP),
			LoadDistribution:     fi.ValueOf(rule.Properties.LoadDistribution),
		}
		if rule.Properties.Probe != nil && rule.Properties.Probe.ID != nil {
			// Extract probe name from the full resource ID
			parts := strings.Split(fi.ValueOf(rule.Properties.Probe.ID), "/")
			r.ProbeName = parts[len(parts)-1]
		}
		actual.Rules = append(actual.Rules, r)
	}

	return actual, nil
}

func (lb *LoadBalancer) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(lb.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (lb *LoadBalancer) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(lb, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchanegable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a Loadbalancer.
func (*LoadBalancer) RenderAzure(t *azure.AzureAPITarget, a, e, changes *LoadBalancer) error {
	if a == nil {
		klog.Infof("Creating a new Loadbalancer with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Loadbalancer with name: %s", fi.ValueOf(e.Name))
	}

	idPrefix := fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network", t.Cloud.SubscriptionID(), *e.ResourceGroup.Name)
	feConfigProperties := &network.FrontendIPConfigurationPropertiesFormat{}
	if *e.External {
		feConfigProperties.PublicIPAddress = &network.PublicIPAddress{
			ID: e.PublicIPAddress.ID,
		}
	} else {
		feConfigProperties.PrivateIPAllocationMethod = to.Ptr(network.IPAllocationMethodDynamic)
		feConfigProperties.Subnet = &network.Subnet{
			ID: to.Ptr(fmt.Sprintf("/%s/virtualNetworks/%s/subnets/%s", idPrefix, *e.Subnet.VirtualNetwork.Name, *e.Subnet.Name)),
		}
	}

	lb := network.LoadBalancer{
		Location: to.Ptr(t.Cloud.Region()),
		SKU: &network.LoadBalancerSKU{
			Name: to.Ptr(e.SKU),
		},
		Properties: &network.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*network.FrontendIPConfiguration{
				{
					Name:       to.Ptr("LoadBalancerFrontEnd"),
					Properties: feConfigProperties,
				},
			},
			BackendAddressPools: []*network.BackendAddressPool{
				{
					Name: to.Ptr("LoadBalancerBackEnd"),
				},
			},
		},
		Tags: e.Tags,
	}

	for _, probe := range e.Probes {
		p := &network.Probe{
			Name: to.Ptr(probe.Name),
			Properties: &network.ProbePropertiesFormat{
				Protocol:          to.Ptr(probe.Protocol),
				Port:              to.Ptr(probe.Port),
				IntervalInSeconds: to.Ptr(probe.IntervalInSeconds),
				NumberOfProbes:    to.Ptr(probe.NumberOfProbes),
			},
		}
		if probe.RequestPath != nil {
			p.Properties.RequestPath = probe.RequestPath
		}
		lb.Properties.Probes = append(lb.Properties.Probes, p)
	}

	for _, rule := range e.Rules {
		lb.Properties.LoadBalancingRules = append(lb.Properties.LoadBalancingRules, &network.LoadBalancingRule{
			Name: to.Ptr(rule.Name),
			Properties: &network.LoadBalancingRulePropertiesFormat{
				Protocol:             to.Ptr(rule.Protocol),
				FrontendPort:         to.Ptr(rule.Port),
				BackendPort:          to.Ptr(rule.Port),
				IdleTimeoutInMinutes: to.Ptr(rule.IdleTimeoutInMinutes),
				EnableFloatingIP:     to.Ptr(rule.EnableFloatingIP),
				LoadDistribution:     to.Ptr(rule.LoadDistribution),
				FrontendIPConfiguration: &network.SubResource{
					ID: to.Ptr(fmt.Sprintf("/%s/loadbalancers/%s/frontendIPConfigurations/%s", idPrefix, *e.Name, "LoadBalancerFrontEnd")),
				},
				BackendAddressPool: &network.SubResource{
					ID: to.Ptr(fmt.Sprintf("/%s/loadbalancers/%s/backendAddressPools/%s", idPrefix, *e.Name, "LoadBalancerBackEnd")),
				},
				Probe: &network.SubResource{
					ID: to.Ptr(fmt.Sprintf("/%s/loadbalancers/%s/probes/%s", idPrefix, *e.Name, rule.ProbeName)),
				},
			},
		})
	}

	_, err := t.Cloud.LoadBalancer().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		lb)

	return err
}
