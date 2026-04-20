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

type terraformAzureManagedIdentity struct {
	Name              *string                  `cty:"name"`
	Location          *string                  `cty:"location"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	Tags              map[string]string        `cty:"tags"`
}

func (*ManagedIdentity) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ManagedIdentity) error {
	tf := &terraformAzureManagedIdentity{
		Name:              e.Name,
		Location:          fi.PtrTo(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		Tags:              stringMap(e.Tags),
	}
	return t.RenderResource("azurerm_user_assigned_identity", fi.ValueOf(e.Name), tf)
}

// terraformID returns a Literal that resolves to the UAMI resource's `id` attribute.
func (m *ManagedIdentity) terraformID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_user_assigned_identity", fi.ValueOf(m.Name), "id")
}

// terraformPrincipalID returns a Literal that resolves to the UAMI's `principal_id` attribute.
// Matches the signature of VMScaleSet.terraformPrincipalID so RoleAssignment can dispatch
// on either principal source.
func (m *ManagedIdentity) terraformPrincipalID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_user_assigned_identity", fi.ValueOf(m.Name), "principal_id")
}
