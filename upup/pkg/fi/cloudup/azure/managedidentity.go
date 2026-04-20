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

// ManagedIdentityClient is a client for managing user-assigned managed identities.
type ManagedIdentityClient interface {
	Get(ctx context.Context, resourceGroupName, identityName string) (*armmsi.Identity, error)
	CreateOrUpdate(ctx context.Context, resourceGroupName, identityName string, parameters armmsi.Identity) (*armmsi.Identity, error)
	Delete(ctx context.Context, resourceGroupName, identityName string) error
}

type managedIdentityClientImpl struct {
	c *armmsi.UserAssignedIdentitiesClient
}

var _ ManagedIdentityClient = (*managedIdentityClientImpl)(nil)

func (c *managedIdentityClientImpl) Get(ctx context.Context, resourceGroupName, identityName string) (*armmsi.Identity, error) {
	resp, err := c.c.Get(ctx, resourceGroupName, identityName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Identity, nil
}

func (c *managedIdentityClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, identityName string, parameters armmsi.Identity) (*armmsi.Identity, error) {
	resp, err := c.c.CreateOrUpdate(ctx, resourceGroupName, identityName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Identity, nil
}

func (c *managedIdentityClientImpl) Delete(ctx context.Context, resourceGroupName, identityName string) error {
	_, err := c.c.Delete(ctx, resourceGroupName, identityName, nil)
	if err != nil {
		return fmt.Errorf("deleting managed identity: %w", err)
	}
	return nil
}

func newManagedIdentityClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*managedIdentityClientImpl, error) {
	c, err := armmsi.NewUserAssignedIdentitiesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating managed identity client: %w", err)
	}
	return &managedIdentityClientImpl{c: c}, nil
}

// FederatedIdentityCredentialClient is a client for managing federated identity credentials.
type FederatedIdentityCredentialClient interface {
	Get(ctx context.Context, resourceGroupName, identityName, credentialName string) (*armmsi.FederatedIdentityCredential, error)
	CreateOrUpdate(ctx context.Context, resourceGroupName, identityName, credentialName string, parameters armmsi.FederatedIdentityCredential) (*armmsi.FederatedIdentityCredential, error)
	Delete(ctx context.Context, resourceGroupName, identityName, credentialName string) error
	List(ctx context.Context, resourceGroupName, identityName string) ([]*armmsi.FederatedIdentityCredential, error)
}

type federatedIdentityCredentialClientImpl struct {
	c *armmsi.FederatedIdentityCredentialsClient
}

var _ FederatedIdentityCredentialClient = (*federatedIdentityCredentialClientImpl)(nil)

func (c *federatedIdentityCredentialClientImpl) Get(ctx context.Context, resourceGroupName, identityName, credentialName string) (*armmsi.FederatedIdentityCredential, error) {
	resp, err := c.c.Get(ctx, resourceGroupName, identityName, credentialName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FederatedIdentityCredential, nil
}

func (c *federatedIdentityCredentialClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, identityName, credentialName string, parameters armmsi.FederatedIdentityCredential) (*armmsi.FederatedIdentityCredential, error) {
	resp, err := c.c.CreateOrUpdate(ctx, resourceGroupName, identityName, credentialName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FederatedIdentityCredential, nil
}

func (c *federatedIdentityCredentialClientImpl) Delete(ctx context.Context, resourceGroupName, identityName, credentialName string) error {
	_, err := c.c.Delete(ctx, resourceGroupName, identityName, credentialName, nil)
	if err != nil {
		return fmt.Errorf("deleting federated identity credential: %w", err)
	}
	return nil
}

func (c *federatedIdentityCredentialClientImpl) List(ctx context.Context, resourceGroupName, identityName string) ([]*armmsi.FederatedIdentityCredential, error) {
	var l []*armmsi.FederatedIdentityCredential
	pager := c.c.NewListPager(resourceGroupName, identityName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing federated identity credentials: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func newFederatedIdentityCredentialClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*federatedIdentityCredentialClientImpl, error) {
	c, err := armmsi.NewFederatedIdentityCredentialsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating federated identity credential client: %w", err)
	}
	return &federatedIdentityCredentialClientImpl{c: c}, nil
}
