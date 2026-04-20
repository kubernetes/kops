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

type terraformAzureFederatedIdentityCredential struct {
	Name              *string                  `cty:"name"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	ParentID          *terraformWriter.Literal `cty:"parent_id"`
	Issuer            *string                  `cty:"issuer"`
	Subject           *string                  `cty:"subject"`
	Audience          []string                 `cty:"audience"`
}

func (*FederatedIdentityCredential) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *FederatedIdentityCredential) error {
	audience := make([]string, 0, len(e.Audiences))
	for _, a := range e.Audiences {
		audience = append(audience, fi.ValueOf(a))
	}
	tf := &terraformAzureFederatedIdentityCredential{
		Name:              e.Name,
		ResourceGroupName: e.ResourceGroup.terraformName(),
		ParentID:          e.ManagedIdentity.terraformID(),
		Issuer:            e.Issuer,
		Subject:           e.Subject,
		Audience:          audience,
	}
	return t.RenderResource("azurerm_federated_identity_credential", fi.ValueOf(e.Name), tf)
}
