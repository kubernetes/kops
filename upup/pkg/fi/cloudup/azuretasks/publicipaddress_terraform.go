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

type terraformAzurePublicIPAddress struct {
	Name              *string                  `cty:"name"`
	Location          *string                  `cty:"location"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	AllocationMethod  *string                  `cty:"allocation_method"`
	SKU               *string                  `cty:"sku"`
	IPVersion         *string                  `cty:"ip_version"`
	Tags              map[string]string        `cty:"tags"`
}

func (*PublicIPAddress) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *PublicIPAddress) error {
	allocationMethod := string(e.AllocationMethod)
	sku := string(e.SKU)
	ipVersion := string(e.IPVersion)
	tf := &terraformAzurePublicIPAddress{
		Name:              e.Name,
		Location:          fi.PtrTo(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		AllocationMethod:  &allocationMethod,
		SKU:               &sku,
		IPVersion:         &ipVersion,
		Tags:              stringMap(e.Tags),
	}
	return t.RenderResource("azurerm_public_ip", fi.ValueOf(e.Name), tf)
}

func (pip *PublicIPAddress) terraformID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_public_ip", fi.ValueOf(pip.Name), "id")
}

func (pip *PublicIPAddress) terraformIPAddress() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_public_ip", fi.ValueOf(pip.Name), "ip_address")
}
