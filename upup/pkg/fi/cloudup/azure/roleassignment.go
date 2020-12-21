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

package azure

import (
	"context"

	// Use 2018-01-01-preview API as we need the version to create
	// a role assignment with Data Actions (https://github.com/Azure/azure-sdk-for-go/issues/1895).
	// The non-preview version of the authorization API (2015-07-01)
	// doesn't support Data Actions.
	authz "github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/go-autorest/autorest"
)

// RoleAssignmentsClient is a client for managing Role Assignments
type RoleAssignmentsClient interface {
	Create(ctx context.Context, scope, roleAssignmentName string, parameters authz.RoleAssignmentCreateParameters) (*authz.RoleAssignment, error)
	List(ctx context.Context, resourceGroupName string) ([]authz.RoleAssignment, error)
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
	ra, err := c.c.Create(ctx, scope, roleAssignmentName, parameters)
	return &ra, err
}

func (c *roleAssignmentsClientImpl) List(ctx context.Context, resourceGroupName string) ([]authz.RoleAssignment, error) {
	var l []authz.RoleAssignment
	for iter, err := c.c.ListForResourceGroupComplete(ctx, resourceGroupName, "" /* filter */); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *roleAssignmentsClientImpl) Delete(ctx context.Context, scope, raName string) error {
	_, err := c.c.Delete(ctx, scope, raName)
	return err
}

func newRoleAssignmentsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *roleAssignmentsClientImpl {
	c := authz.NewRoleAssignmentsClient(subscriptionID)
	c.Authorizer = authorizer
	return &roleAssignmentsClientImpl{
		c: &c,
	}
}
