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
	resources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// ResourceGroupsClient is a client for managing resource groups.
type ResourceGroupsClient interface {
	CreateOrUpdate(ctx context.Context, name string, parameters resources.ResourceGroup) error
	List(ctx context.Context) ([]*resources.ResourceGroup, error)
	Delete(ctx context.Context, name string) error
}

type resourceGroupsClientImpl struct {
	c *resources.ResourceGroupsClient
}

var _ ResourceGroupsClient = &resourceGroupsClientImpl{}

func (c *resourceGroupsClientImpl) CreateOrUpdate(ctx context.Context, name string, parameters resources.ResourceGroup) error {
	_, err := c.c.CreateOrUpdate(ctx, name, parameters, nil)
	return err
}

func (c *resourceGroupsClientImpl) List(ctx context.Context) ([]*resources.ResourceGroup, error) {
	var l []*resources.ResourceGroup
	pager := c.c.NewListPager(nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing resource groups: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *resourceGroupsClientImpl) Delete(ctx context.Context, name string) error {
	future, err := c.c.BeginDelete(ctx, name, nil)
	if err != nil {
		return fmt.Errorf("deleting resource group: %w", err)
	}
	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for resource group deletion completion: %w", err)
	}
	return nil
}

func newResourceGroupsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*resourceGroupsClientImpl, error) {
	c, err := resources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating resource group client: %w", err)
	}
	return &resourceGroupsClientImpl{
		c: c,
	}, nil
}
