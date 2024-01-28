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

// SubnetsClient is a client for managing subnets.
type SubnetsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string, parameters network.Subnet) (*network.Subnet, error)
	List(ctx context.Context, resourceGroupName, virtualNetworkName string) ([]*network.Subnet, error)
	Delete(ctx context.Context, resourceGroupName, vnetName, subnetName string) error
}

type subnetsClientImpl struct {
	c *network.SubnetsClient
}

var _ SubnetsClient = &subnetsClientImpl{}

func (c *subnetsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string, parameters network.Subnet) (*network.Subnet, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating subnet: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for subnet create/update completion: %w", err)
	}
	return &resp.Subnet, err
}

func (c *subnetsClientImpl) List(ctx context.Context, resourceGroupName, virtualNetworkName string) ([]*network.Subnet, error) {
	var l []*network.Subnet
	pager := c.c.NewListPager(resourceGroupName, virtualNetworkName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing subnets: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *subnetsClientImpl) Delete(ctx context.Context, resourceGroupName, vnetName, subnetName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, vnetName, subnetName, nil)
	if err != nil {
		return fmt.Errorf("deleting subnet: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for subnet deletion completion: %w", err)
	}
	return nil
}

func newSubnetsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*subnetsClientImpl, error) {
	c, err := network.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating subnets client: %w", err)
	}
	return &subnetsClientImpl{
		c: c,
	}, nil
}
