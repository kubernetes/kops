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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type terraformAzureNatGateway struct {
	Name              *string                  `cty:"name"`
	Location          *string                  `cty:"location"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	SKUName           *string                  `cty:"sku_name"`
	Tags              map[string]string        `cty:"tags"`
}

type terraformAzureNatGatewayPublicIPAssociation struct {
	NatGatewayID      *terraformWriter.Literal `cty:"nat_gateway_id"`
	PublicIPAddressID *terraformWriter.Literal `cty:"public_ip_address_id"`
}

func (*NatGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NatGateway) error {
	skuName := string(e.SKU)
	tf := &terraformAzureNatGateway{
		Name:              e.Name,
		Location:          new(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		SKUName:           &skuName,
		Tags:              stringMap(e.Tags),
	}
	if err := t.RenderResource("azurerm_nat_gateway", fi.ValueOf(e.Name), tf); err != nil {
		return err
	}

	for _, pip := range e.PublicIPAddresses {
		assoc := &terraformAzureNatGatewayPublicIPAssociation{
			NatGatewayID:      e.terraformID(),
			PublicIPAddressID: pip.terraformID(),
		}
		if err := t.RenderResource("azurerm_nat_gateway_public_ip_association", e.terraformAssociationName(pip), assoc); err != nil {
			return err
		}
	}

	return nil
}

func (ngw *NatGateway) terraformID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_nat_gateway", fi.ValueOf(ngw.Name), "id")
}

func (ngw *NatGateway) terraformAssociationName(pip *PublicIPAddress) string {
	return fmt.Sprintf("%s-%s", fi.ValueOf(ngw.Name), fi.ValueOf(pip.Name))
}
