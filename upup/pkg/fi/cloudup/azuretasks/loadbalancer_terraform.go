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

package azuretasks

import (
	"fmt"

	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

const (
	terraformAzureLoadBalancerFrontendName = "LoadBalancerFrontEnd"
	terraformAzureLoadBalancerBackendName  = "LoadBalancerBackEnd"
)

type terraformAzureLoadBalancerFrontendIPConfiguration struct {
	Name                      *string                  `cty:"name"`
	PublicIPAddressID         *terraformWriter.Literal `cty:"public_ip_address_id"`
	PrivateIPAllocationMethod *string                  `cty:"private_ip_address_allocation"`
	SubnetID                  *terraformWriter.Literal `cty:"subnet_id"`
}

type terraformAzureLoadBalancer struct {
	Name                    *string                                              `cty:"name"`
	Location                *string                                              `cty:"location"`
	ResourceGroupName       *terraformWriter.Literal                             `cty:"resource_group_name"`
	SKU                     *string                                              `cty:"sku"`
	FrontendIPConfiguration []*terraformAzureLoadBalancerFrontendIPConfiguration `cty:"frontend_ip_configuration"`
	Tags                    map[string]string                                    `cty:"tags"`
}

type terraformAzureLoadBalancerBackendAddressPool struct {
	Name           *string                  `cty:"name"`
	LoadBalancerID *terraformWriter.Literal `cty:"loadbalancer_id"`
}

type terraformAzureLoadBalancerProbe struct {
	Name              *string                  `cty:"name"`
	LoadBalancerID    *terraformWriter.Literal `cty:"loadbalancer_id"`
	Protocol          *string                  `cty:"protocol"`
	Port              *int32                   `cty:"port"`
	RequestPath       *string                  `cty:"request_path"`
	IntervalInSeconds *int32                   `cty:"interval_in_seconds"`
	NumberOfProbes    *int32                   `cty:"number_of_probes"`
}

type terraformAzureLoadBalancerRule struct {
	Name                        *string                    `cty:"name"`
	LoadBalancerID              *terraformWriter.Literal   `cty:"loadbalancer_id"`
	Protocol                    *string                    `cty:"protocol"`
	FrontendPort                *int32                     `cty:"frontend_port"`
	BackendPort                 *int32                     `cty:"backend_port"`
	FrontendIPConfigurationName *string                    `cty:"frontend_ip_configuration_name"`
	BackendAddressPoolIDs       []*terraformWriter.Literal `cty:"backend_address_pool_ids"`
	ProbeID                     *terraformWriter.Literal   `cty:"probe_id"`
	IdleTimeoutInMinutes        *int32                     `cty:"idle_timeout_in_minutes"`
	FloatingIPEnabled           *bool                      `cty:"floating_ip_enabled"`
	LoadDistribution            *string                    `cty:"load_distribution"`
}

func (*LoadBalancer) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *LoadBalancer) error {
	sku := string(e.SKU)
	tf := &terraformAzureLoadBalancer{
		Name:              e.Name,
		Location:          new(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		SKU:               &sku,
		Tags:              stringMap(e.Tags),
	}

	frontend := &terraformAzureLoadBalancerFrontendIPConfiguration{
		Name: new(terraformAzureLoadBalancerFrontendName),
	}
	if fi.ValueOf(e.External) {
		frontend.PublicIPAddressID = e.PublicIPAddress.terraformID()
	} else {
		allocationMethod := string(network.IPAllocationMethodDynamic)
		frontend.PrivateIPAllocationMethod = &allocationMethod
		subnetID, err := e.Subnet.terraformID(t)
		if err != nil {
			return err
		}
		frontend.SubnetID = subnetID
	}
	tf.FrontendIPConfiguration = []*terraformAzureLoadBalancerFrontendIPConfiguration{frontend}

	if err := t.RenderResource("azurerm_lb", fi.ValueOf(e.Name), tf); err != nil {
		return err
	}

	backendName := terraformAzureLoadBalancerBackendName
	if err := t.RenderResource("azurerm_lb_backend_address_pool", fmt.Sprintf("%s-backend-pool", fi.ValueOf(e.Name)), &terraformAzureLoadBalancerBackendAddressPool{
		Name:           &backendName,
		LoadBalancerID: e.terraformID(),
	}); err != nil {
		return err
	}

	for _, probe := range e.Probes {
		probeResourceName := fmt.Sprintf("%s-%s", fi.ValueOf(e.Name), probe.Name)
		if err := t.RenderResource("azurerm_lb_probe", probeResourceName, &terraformAzureLoadBalancerProbe{
			Name:              new(probe.Name),
			LoadBalancerID:    e.terraformID(),
			Protocol:          new(string(probe.Protocol)),
			Port:              new(probe.Port),
			RequestPath:       probe.RequestPath,
			IntervalInSeconds: new(probe.IntervalInSeconds),
			NumberOfProbes:    new(probe.NumberOfProbes),
		}); err != nil {
			return err
		}
	}

	for _, rule := range e.Rules {
		ruleResourceName := fmt.Sprintf("%s-%s", fi.ValueOf(e.Name), rule.Name)
		probeResourceName := fmt.Sprintf("%s-%s", fi.ValueOf(e.Name), rule.ProbeName)
		ruleProtocol := string(rule.Protocol)
		loadDistribution := string(rule.LoadDistribution)
		if err := t.RenderResource("azurerm_lb_rule", ruleResourceName, &terraformAzureLoadBalancerRule{
			Name:                        new(rule.Name),
			LoadBalancerID:              e.terraformID(),
			Protocol:                    &ruleProtocol,
			FrontendPort:                new(rule.Port),
			BackendPort:                 new(rule.Port),
			FrontendIPConfigurationName: new(terraformAzureLoadBalancerFrontendName),
			BackendAddressPoolIDs:       []*terraformWriter.Literal{e.terraformBackendAddressPoolID()},
			ProbeID:                     terraformWriter.LiteralProperty("azurerm_lb_probe", probeResourceName, "id"),
			IdleTimeoutInMinutes:        new(rule.IdleTimeoutInMinutes),
			FloatingIPEnabled:           new(rule.EnableFloatingIP),
			LoadDistribution:            &loadDistribution,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (lb *LoadBalancer) terraformID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_lb", fi.ValueOf(lb.Name), "id")
}

func (lb *LoadBalancer) terraformBackendAddressPoolID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_lb_backend_address_pool", fmt.Sprintf("%s-backend-pool", fi.ValueOf(lb.Name)), "id")
}
