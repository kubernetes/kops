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
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	containerregistry "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// ContainerRegistriesClient is a client for managing container registries.
type ContainerRegistriesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, registryName string, parameters containerregistry.Registry) (*containerregistry.Registry, error)
	List(ctx context.Context, resourceGroupName string) ([]*containerregistry.Registry, error)
	Delete(ctx context.Context, resourceGroupName, registryName string) error
}

type containerRegistriesClientImpl struct {
	c *containerregistry.RegistriesClient
}

var _ ContainerRegistriesClient = (*containerRegistriesClientImpl)(nil)

func (c *containerRegistriesClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, registryName string, parameters containerregistry.Registry) (*containerregistry.Registry, error) {
	future, err := c.c.BeginCreate(ctx, resourceGroupName, registryName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating container registry: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for container registry creation completion: %w", err)
	}
	return &resp.Registry, nil
}

func (c *containerRegistriesClientImpl) List(ctx context.Context, resourceGroupName string) ([]*containerregistry.Registry, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*containerregistry.Registry
	pager := c.c.NewListByResourceGroupPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing container registries: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *containerRegistriesClientImpl) Delete(ctx context.Context, resourceGroupName, registryName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, registryName, nil)
	if err != nil {
		return fmt.Errorf("deleting container registry: %w", err)
	}
	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for container registry deletion completion: %w", err)
	}
	return nil
}

func newContainerRegistriesClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*containerRegistriesClientImpl, error) {
	c, err := containerregistry.NewRegistriesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating container registries client: %w", err)
	}
	return &containerRegistriesClientImpl{
		c: c,
	}, nil
}
