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

type terraformAzureUserAssignedIdentity struct {
	Name              *string                  `cty:"name"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	Location          *string                  `cty:"location"`
	Tags              map[string]string        `cty:"tags"`
}

func (*ManagedIdentity) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ManagedIdentity) error {
	tf := &terraformAzureUserAssignedIdentity{
		Name:              e.Name,
		ResourceGroupName: e.ResourceGroup.terraformName(),
		Location:          fi.PtrTo(t.Cloud.Region()),
		Tags:              stringMap(e.Tags),
	}
	return t.RenderResource("azurerm_user_assigned_identity", fi.ValueOf(e.Name), tf)
}

func (m *ManagedIdentity) terraformID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_user_assigned_identity", fi.ValueOf(m.Name), "id")
}

func (m *ManagedIdentity) terraformPrincipalID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_user_assigned_identity", fi.ValueOf(m.Name), "principal_id")
}
