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
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// VirtualNetworksClient is a client for managing Virtual Networks.
type VirtualNetworksClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName string, parameters network.VirtualNetwork) (*network.VirtualNetwork, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.VirtualNetwork, error)
	Delete(ctx context.Context, resourceGroupName, vnetName string) error
}

type virtualNetworksClientImpl struct {
	c *network.VirtualNetworksClient
}

var _ VirtualNetworksClient = &virtualNetworksClientImpl{}

func (c *virtualNetworksClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName string, parameters network.VirtualNetwork) (*network.VirtualNetwork, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating virtual network: %w", err)
	}
	vnet, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for virtual network create/update completion: %w", err)
	}
	return &vnet.VirtualNetwork, err
}

func (c *virtualNetworksClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.VirtualNetwork, error) {
	var l []*network.VirtualNetwork
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing virtual networks: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *virtualNetworksClientImpl) Delete(ctx context.Context, resourceGroupName, vnetName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, vnetName, nil)
	if err != nil {
		return fmt.Errorf("deleting virtual network: %w", err)
	}
	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for virtual network deletion completion: %w", err)
	}
	return nil
}

func newVirtualNetworksClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*virtualNetworksClientImpl, error) {
	c, err := network.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating virtual networks client: %w", err)
	}
	return &virtualNetworksClientImpl{
		c: c,
	}, nil
}
