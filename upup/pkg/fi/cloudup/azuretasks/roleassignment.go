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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
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

	Scope           *string
	ManagedIdentity *ManagedIdentity
	ID              *string
	RoleDefID       *string
}

var (
	_ fi.CloudupTask   = &RoleAssignment{}
	_ fi.CompareWithID = &RoleAssignment{}
	// RoleAssignment does not implement CloudupTaskNormalize because Azure role assignments do not support tags.
)

// CompareWithID returns the Name of the RoleAssignment.
func (r *RoleAssignment) CompareWithID() *string {
	return r.Name
}

// Find discovers the RoleAssignment in the cloud provider.
func (r *RoleAssignment) Find(c *fi.CloudupContext) (*RoleAssignment, error) {
	if r.ManagedIdentity.PrincipalID == nil {
		// PrincipalID of the Managed Identity hasn't yet been
		// populated. No corresponding Role Assignment
		// should exist in Cloud.
		return nil, nil
	}

	cloud := c.T.Cloud.(azure.AzureCloud)
	rs, err := cloud.RoleAssignment().List(context.TODO(), *r.Scope)
	if err != nil {
		return nil, err
	}

	principalID := *r.ManagedIdentity.PrincipalID
	var found *authz.RoleAssignment
	for i := range rs {
		ra := rs[i]
		if ra.Properties == nil || ra.Properties.RoleDefinitionID == nil || ra.Properties.PrincipalID == nil {
			continue
		}
		parsed, err := arm.ParseResourceID(*ra.Properties.RoleDefinitionID)
		if err != nil {
			continue
		}
		if *ra.Properties.PrincipalID == principalID && parsed.Name == *r.RoleDefID {
			found = ra
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	r.ID = found.ID
	return &RoleAssignment{
		Name:      r.Name,
		Lifecycle: r.Lifecycle,
		Scope:     found.Properties.Scope,
		ManagedIdentity: &ManagedIdentity{
			Name: r.ManagedIdentity.Name,
		},
		ID:        found.ID,
		RoleDefID: r.RoleDefID,
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
			PrincipalID:      e.ManagedIdentity.PrincipalID,
			// PrincipalType must be set to avoid PrincipalNotFound errors caused by
			// Entra ID replication delay after managed identity creation.
			PrincipalType: to.Ptr(authz.PrincipalTypeServicePrincipal),
		},
	}
	ra, err := t.Cloud.RoleAssignment().Create(context.TODO(), scope, roleAssignmentName, roleAssignment)
	if err != nil {
		return err
	}
	e.ID = ra.ID
	return nil
}
