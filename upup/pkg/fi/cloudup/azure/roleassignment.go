/*
Copyright 2024 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	authz "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
)

// RoleAssignmentsClient is a client for managing role assignments
type RoleAssignmentsClient interface {
	Create(ctx context.Context, scope, roleAssignmentName string, parameters authz.RoleAssignmentCreateParameters) (*authz.RoleAssignment, error)
	List(ctx context.Context, scope string) ([]*authz.RoleAssignment, error)
	Delete(ctx context.Context, scope, raName string) error
}

type roleAssignmentsClientImpl struct {
	c *authz.RoleAssignmentsClient
}

var _ RoleAssignmentsClient = &roleAssignmentsClientImpl{}

func (c *roleAssignmentsClientImpl) Create(
	ctx context.Context,
	scope string,
	roleAssignmentName string,
	parameters authz.RoleAssignmentCreateParameters,
) (*authz.RoleAssignment, error) {
	resp, err := c.c.Create(ctx, scope, roleAssignmentName, parameters, nil)
	return &resp.RoleAssignment, err
}

func (c *roleAssignmentsClientImpl) List(ctx context.Context, scope string) ([]*authz.RoleAssignment, error) {
	var l []*authz.RoleAssignment
	pager := c.c.NewListForScopePager(scope, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing role assignments: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *roleAssignmentsClientImpl) Delete(ctx context.Context, scope, raName string) error {
	_, err := c.c.Delete(ctx, scope, raName, nil)
	if err != nil {
		return fmt.Errorf("deleting role assignment: %w", err)
	}
	return nil
}

func newRoleAssignmentsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*roleAssignmentsClientImpl, error) {
	c, err := authz.NewRoleAssignmentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating role assignments client: %w", err)
	}
	return &roleAssignmentsClientImpl{
		c: c,
	}, nil
}
