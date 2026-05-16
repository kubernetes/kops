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

package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
)

// ManagedIdentitiesClient is a client for managing user-assigned managed identities.
type ManagedIdentitiesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, name string, parameters armmsi.Identity) (*armmsi.Identity, error)
	Get(ctx context.Context, resourceGroupName, name string) (*armmsi.Identity, error)
	List(ctx context.Context, resourceGroupName string) ([]*armmsi.Identity, error)
	Delete(ctx context.Context, resourceGroupName, name string) error
}

type managedIdentitiesClientImpl struct {
	c *armmsi.UserAssignedIdentitiesClient
}

var _ ManagedIdentitiesClient = (*managedIdentitiesClientImpl)(nil)

func (c *managedIdentitiesClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, name string, parameters armmsi.Identity) (*armmsi.Identity, error) {
	resp, err := c.c.CreateOrUpdate(ctx, resourceGroupName, name, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Identity, nil
}

func (c *managedIdentitiesClientImpl) Get(ctx context.Context, resourceGroupName, name string) (*armmsi.Identity, error) {
	resp, err := c.c.Get(ctx, resourceGroupName, name, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Identity, nil
}

func (c *managedIdentitiesClientImpl) List(ctx context.Context, resourceGroupName string) ([]*armmsi.Identity, error) {
	var l []*armmsi.Identity
	pager := c.c.NewListByResourceGroupPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing managed identities: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *managedIdentitiesClientImpl) Delete(ctx context.Context, resourceGroupName, name string) error {
	_, err := c.c.Delete(ctx, resourceGroupName, name, nil)
	return err
}

func newManagedIdentitiesClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*managedIdentitiesClientImpl, error) {
	c, err := armmsi.NewUserAssignedIdentitiesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating managed identities client: %w", err)
	}
	return &managedIdentitiesClientImpl{
		c: c,
	}, nil
}
