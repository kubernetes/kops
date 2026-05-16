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
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/google/uuid"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestRoleAssignmentRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	ra := &RoleAssignment{}
	expected := &RoleAssignment{
		Name:  to.Ptr("ra"),
		Scope: to.Ptr("scope"),
		ManagedIdentity: &ManagedIdentity{
			Name:        to.Ptr("mi"),
			PrincipalID: to.Ptr("pid"),
		},
		RoleDefID: to.Ptr("rdid0"),
	}

	if err := ra.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if expected.ID == nil {
		t.Fatalf("id must be set")
	}
	actual := cloud.RoleAssignmentsClient.RAs[*expected.ID]
	if a, e := *actual.Properties.PrincipalID, *expected.ManagedIdentity.PrincipalID; a != e {
		t.Errorf("unexpected principal ID: expected %s, but got %s", e, a)
	}
}

func TestRoleAssignmentFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	// Create a managed identity.
	miName := "test-identity"
	miResult, err := cloud.ManagedIdentity().CreateOrUpdate(context.TODO(), "rg", miName, armmsi.Identity{
		Location: to.Ptr("eastus"),
	})
	if err != nil {
		t.Fatalf("failed to create managed identity: %s", err)
	}

	scope := "/subscriptions/sub/resourceGroups/rg"
	roleDefID := "rdid0"
	ra := &RoleAssignment{
		Name:  to.Ptr(miName),
		Scope: &scope,
		ManagedIdentity: &ManagedIdentity{
			Name:        to.Ptr(miName),
			PrincipalID: miResult.Properties.PrincipalID,
		},
		RoleDefID: &roleDefID,
	}
	// Find will return nothing if there is no Role Assignment created.
	actual, err := ra.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected role assignment found: %+v", actual)
	}

	// Create Role Assignments. One of them has irrelevant (different role definition ID).
	roleAssignmentName := uuid.New().String()
	roleAssignment := authz.RoleAssignmentCreateParameters{
		Properties: &authz.RoleAssignmentProperties{
			RoleDefinitionID: to.Ptr(roleDefID),
			PrincipalID:      miResult.Properties.PrincipalID,
		},
	}
	if _, err := cloud.RoleAssignment().Create(context.TODO(), scope, roleAssignmentName, roleAssignment); err != nil {
		t.Fatalf("failed to create: %s", err)
	}

	irrelevant := authz.RoleAssignmentCreateParameters{
		Properties: &authz.RoleAssignmentProperties{
			RoleDefinitionID: to.Ptr("irrelevant"),
			PrincipalID:      miResult.Properties.PrincipalID,
		},
	}
	if _, err := cloud.RoleAssignment().Create(context.TODO(), scope, uuid.New().String(), irrelevant); err != nil {
		t.Fatalf("failed to create: %s", err)
	}

	// Find again.
	actual, err = ra.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual == nil {
		t.Fatalf("expected role assignment to be found")
	}
	if a, e := *actual.Name, *ra.Name; a != e {
		t.Errorf("unexpected Role name Assignment: expected %+v, but got %+v", e, a)
	}
	if a, e := *actual.RoleDefID, roleDefID; a != e {
		t.Errorf("unexpected Role def ID: expected %s, but got %s", e, a)
	}
}

// TestRoleAssignmentFind_NoPrincipalID verifies that Find doesn't find any Role Assignment
// when the principal ID of the Managed Identity hasn't yet been set.
func TestRoleAssignmentFind_NoPrincipalID(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	// Create a dummy Role Assignment to ensure that this won't be returned by Find.
	roleAssignment := authz.RoleAssignmentCreateParameters{
		Properties: &authz.RoleAssignmentProperties{
			RoleDefinitionID: to.Ptr("rdid0"),
		},
	}
	if _, err := cloud.RoleAssignment().Create(context.TODO(), "scope", uuid.New().String(), roleAssignment); err != nil {
		t.Fatalf("failed to create Role Assignment: %s", err)
	}

	scope := "scope"
	ra := &RoleAssignment{
		Name:  to.Ptr("test"),
		Scope: to.Ptr(scope),
		ManagedIdentity: &ManagedIdentity{
			Name: to.Ptr("mi"),
			// Do not set principal ID.
		},
	}
	actual, err := ra.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected Role Assignment found: %+v", actual)
	}
}

func TestRoleAssignmentCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *RoleAssignment
		success       bool
	}{
		{
			a:       nil,
			e:       &RoleAssignment{Name: to.Ptr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &RoleAssignment{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &RoleAssignment{Name: to.Ptr("name")},
			changes: &RoleAssignment{Name: nil},
			success: true,
		},
		{
			a:       &RoleAssignment{Name: to.Ptr("name")},
			changes: &RoleAssignment{Name: to.Ptr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			ra := RoleAssignment{}
			err := ra.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
