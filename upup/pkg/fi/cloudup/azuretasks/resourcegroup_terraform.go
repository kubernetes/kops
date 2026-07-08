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

type terraformAzureResourceGroup struct {
	Name     *string           `cty:"name"`
	Location *string           `cty:"location"`
	Tags     map[string]string `cty:"tags"`
}

func (*ResourceGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ResourceGroup) error {
	if fi.ValueOf(e.Shared) {
		return nil
	}

	tf := &terraformAzureResourceGroup{
		Name:     e.Name,
		Location: new(t.Cloud.Region()),
		Tags:     stringMap(e.Tags),
	}
	return t.RenderResource("azurerm_resource_group", fi.ValueOf(e.Name), tf)
}

func (r *ResourceGroup) terraformName() *terraformWriter.Literal {
	if fi.ValueOf(r.Shared) {
		return terraformWriter.LiteralFromStringValue(fi.ValueOf(r.Name))
	}
	return terraformWriter.LiteralProperty("azurerm_resource_group", fi.ValueOf(r.Name), "name")
}
