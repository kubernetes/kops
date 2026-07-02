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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type terraformAzureNetworkSecurityRule struct {
	Name                                   *string                    `cty:"name"`
	Priority                               *int32                     `cty:"priority"`
	Access                                 *string                    `cty:"access"`
	Direction                              *string                    `cty:"direction"`
	Protocol                               *string                    `cty:"protocol"`
	SourceAddressPrefix                    *terraformWriter.Literal   `cty:"source_address_prefix"`
	SourceAddressPrefixes                  []string                   `cty:"source_address_prefixes"`
	SourceApplicationSecurityGroupIDs      []*terraformWriter.Literal `cty:"source_application_security_group_ids"`
	SourcePortRange                        *string                    `cty:"source_port_range"`
	DestinationAddressPrefix               *string                    `cty:"destination_address_prefix"`
	DestinationAddressPrefixes             []string                   `cty:"destination_address_prefixes"`
	DestinationApplicationSecurityGroupIDs []*terraformWriter.Literal `cty:"destination_application_security_group_ids"`
	DestinationPortRange                   *string                    `cty:"destination_port_range"`
}

type terraformAzureNetworkSecurityGroup struct {
	Name              *string                              `cty:"name"`
	Location          *string                              `cty:"location"`
	ResourceGroupName *terraformWriter.Literal             `cty:"resource_group_name"`
	SecurityRule      []*terraformAzureNetworkSecurityRule `cty:"security_rule"`
	Tags              map[string]string                    `cty:"tags"`
}

func (*NetworkSecurityGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NetworkSecurityGroup) error {
	tf := &terraformAzureNetworkSecurityGroup{
		Name:              e.Name,
		Location:          fi.PtrTo(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		Tags:              stringMap(e.Tags),
	}

	for _, rule := range e.SecurityRules {
		tf.SecurityRule = append(tf.SecurityRule, rule.toTerraform())
	}

	return t.RenderResource("azurerm_network_security_group", fi.ValueOf(e.Name), tf)
}

func (nsg *NetworkSecurityGroup) terraformID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_network_security_group", fi.ValueOf(nsg.Name), "id")
}

func (rule *NetworkSecurityRule) toTerraform() *terraformAzureNetworkSecurityRule {
	access := string(rule.Access)
	direction := string(rule.Direction)
	protocol := string(rule.Protocol)
	var sourceAddressPrefix *terraformWriter.Literal
	if rule.SourcePublicIPAddress != nil {
		sourceAddressPrefix = terraformWriter.LiteralProperty("azurerm_public_ip", fi.ValueOf(rule.SourcePublicIPAddress.Name), "ip_address")
	} else if rule.SourceAddressPrefix != nil {
		sourceAddressPrefix = terraformWriter.LiteralFromStringValue(*rule.SourceAddressPrefix)
	}
	return &terraformAzureNetworkSecurityRule{
		Name:                                   rule.Name,
		Priority:                               rule.Priority,
		Access:                                 &access,
		Direction:                              &direction,
		Protocol:                               &protocol,
		SourceAddressPrefix:                    sourceAddressPrefix,
		SourceAddressPrefixes:                  stringSlice(rule.SourceAddressPrefixes),
		SourceApplicationSecurityGroupIDs:      applicationSecurityGroupNameIDs(rule.SourceApplicationSecurityGroupNames),
		SourcePortRange:                        rule.SourcePortRange,
		DestinationAddressPrefix:               rule.DestinationAddressPrefix,
		DestinationAddressPrefixes:             stringSlice(rule.DestinationAddressPrefixes),
		DestinationApplicationSecurityGroupIDs: applicationSecurityGroupNameIDs(rule.DestinationApplicationSecurityGroupNames),
		DestinationPortRange:                   rule.DestinationPortRange,
	}
}
