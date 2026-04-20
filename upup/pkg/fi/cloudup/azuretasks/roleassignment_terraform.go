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

type terraformAzureRoleAssignment struct {
	Scope                        *string                  `cty:"scope"`
	RoleDefinitionID             *string                  `cty:"role_definition_id"`
	PrincipalID                  *terraformWriter.Literal `cty:"principal_id"`
	SkipServicePrincipalAADCheck *bool                    `cty:"skip_service_principal_aad_check"`
}

func (*RoleAssignment) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RoleAssignment) error {
	principalID, err := e.terraformPrincipalID()
	if err != nil {
		return err
	}
	roleDefinitionID := fmt.Sprintf("%s/providers/Microsoft.Authorization/roleDefinitions/%s", fi.ValueOf(e.Scope), fi.ValueOf(e.RoleDefID))
	tf := &terraformAzureRoleAssignment{
		Scope:                        e.Scope,
		RoleDefinitionID:             &roleDefinitionID,
		PrincipalID:                  principalID,
		SkipServicePrincipalAADCheck: fi.PtrTo(true),
	}
	return t.RenderResource("azurerm_role_assignment", fi.ValueOf(e.Name), tf)
}

// terraformPrincipalID resolves the principal ID Literal from whichever
// principal source (VMScaleSet or ManagedIdentity) is set. The fi-task-level
// invariant is that exactly one is set; if that's violated we return an error
// rather than silently picking one.
func (r *RoleAssignment) terraformPrincipalID() (*terraformWriter.Literal, error) {
	switch {
	case r.ManagedIdentity != nil && r.VMScaleSet != nil:
		return nil, fmt.Errorf("RoleAssignment %q has both ManagedIdentity and VMScaleSet set; only one is allowed", fi.ValueOf(r.Name))
	case r.ManagedIdentity != nil:
		return r.ManagedIdentity.terraformPrincipalID(), nil
	case r.VMScaleSet != nil:
		return r.VMScaleSet.terraformPrincipalID(), nil
	default:
		return nil, fmt.Errorf("RoleAssignment %q has no principal source set", fi.ValueOf(r.Name))
	}
}
