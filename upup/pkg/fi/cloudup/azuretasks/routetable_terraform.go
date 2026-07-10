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

type terraformAzureRouteTable struct {
	Name              *string                  `cty:"name"`
	Location          *string                  `cty:"location"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	Tags              map[string]string        `cty:"tags"`
}

func (*RouteTable) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RouteTable) error {
	if fi.ValueOf(e.Shared) {
		return nil
	}

	tf := &terraformAzureRouteTable{
		Name:              e.Name,
		Location:          new(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		Tags:              stringMap(e.Tags),
	}
	return t.RenderResource("azurerm_route_table", fi.ValueOf(e.Name), tf)
}

func (rt *RouteTable) terraformID() *terraformWriter.Literal {
	if fi.ValueOf(rt.Shared) {
		return terraformWriter.LiteralFromStringValue(fi.ValueOf(rt.Name))
	}
	return terraformWriter.LiteralProperty("azurerm_route_table", fi.ValueOf(rt.Name), "id")
}
