/*
Copyright 2020 The Kubernetes Authors.

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
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/google/uuid"
)

// RoleAssignment is an Azure Role Assignment.
// +kops:fitask
type RoleAssignment struct {
	// Name is the name of the RoleAssignment task. This is
	// different from a name of Role Assignment, which is GUID.
	// As kops cannot give a fixed name to the Role Assignment
	// name generated from kops cluster config, we keep a task
	// name and a Role Assignment name separate.
	Name      *string
	Lifecycle fi.Lifecycle

	Scope      *string
	VMScaleSet *VMScaleSet
	ID         *string
	RoleDefID  *string
}

var (
	_ fi.CloudupTask   = &RoleAssignment{}
	_ fi.CompareWithID = &RoleAssignment{}
)

// CompareWithID returns the Name of the VM Scale Set.
func (r *RoleAssignment) CompareWithID() *string {
	return r.Name
}

// Find discovers the RoleAssignment in the cloud provider.
func (r *RoleAssignment) Find(c *fi.CloudupContext) (*RoleAssignment, error) {
	if r.VMScaleSet.PrincipalID == nil {
		// PrincipalID of the VM Scale Set hasn't yet been
		// populated. No corresponding Role Assignment
		// shouldn't exist in Cloud.
		return nil, nil
	}

	cloud := c.T.Cloud.(azure.AzureCloud)
	rs, err := cloud.RoleAssignment().List(context.TODO(), *r.Scope)
	if err != nil {
		return nil, err
	}

	principalID := *r.VMScaleSet.PrincipalID
	var found *authz.RoleAssignment
	for i := range rs {
		ra := rs[i]
		if ra.Properties == nil {
			continue
		}
		// Use a name constructed by VMSS and Role definition ID to find a Role Assignment. We cannot use ra.Name
		// as it is set to a randomly generated GUID.
		l := strings.Split(*ra.Properties.RoleDefinitionID, "/")
		roleDefID := l[len(l)-1]
		if *ra.Properties.PrincipalID == principalID && roleDefID == *r.RoleDefID {
			found = ra
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	// Query VM Scale Sets and find one that has matching Principal ID.
	vs, err := cloud.VMScaleSet().List(context.TODO(), *r.VMScaleSet.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var foundVMSS *compute.VirtualMachineScaleSet
	for _, v := range vs {
		if v.Identity == nil {
			continue
		}
		if *v.Identity.PrincipalID == principalID {
			foundVMSS = v
			break
		}
	}
	if foundVMSS == nil {
		return nil, fmt.Errorf("corresponding VM Scale Set not found for Role Assignment: %s", *found.ID)
	}

	r.ID = found.ID
	return &RoleAssignment{
		Name:      r.Name,
		Lifecycle: r.Lifecycle,
		Scope:     found.Properties.Scope,
		VMScaleSet: &VMScaleSet{
			Name: foundVMSS.Name,
		},
		ID:        found.ID,
		RoleDefID: to.Ptr(filepath.Base(*found.Properties.RoleDefinitionID)),
	}, nil
}

// Run implements fi.Task.Run.
func (r *RoleAssignment) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(r, c)
}

// CheckChanges returns an error if a change is not allowed.
func (r *RoleAssignment) CheckChanges(a, e, changes *RoleAssignment) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchangeable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a Role Assignment.
func (*RoleAssignment) RenderAzure(t *azure.AzureAPITarget, a, e, changes *RoleAssignment) error {
	if a == nil {
		return createNewRoleAssignment(t, e)
	}
	if changes.ID != nil && changes.RoleDefID != nil {
		return errors.New("updating Role Assignment is not yet implemented")
	}
	return nil
}

func createNewRoleAssignment(t *azure.AzureAPITarget, e *RoleAssignment) error {
	// We generate the name of Role Assignment here. It must be a valid GUID.
	roleAssignmentName := uuid.New().String()

	scope := *e.Scope
	roleDefID := fmt.Sprintf("%s/providers/Microsoft.Authorization/roleDefinitions/%s", scope, *e.RoleDefID)
	roleAssignment := authz.RoleAssignmentCreateParameters{
		Properties: &authz.RoleAssignmentProperties{
			RoleDefinitionID: to.Ptr(roleDefID),
			PrincipalID:      e.VMScaleSet.PrincipalID,
		},
	}
	ra, err := t.Cloud.RoleAssignment().Create(context.TODO(), scope, roleAssignmentName, roleAssignment)
	if err != nil {
		return err
	}
	e.ID = ra.ID
	return nil
}
