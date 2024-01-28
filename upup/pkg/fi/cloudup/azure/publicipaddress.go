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

// PublicIPAddressesClient is a client for public IP addresses.
type PublicIPAddressesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, publicIPAddressName string, parameters network.PublicIPAddress) (*network.PublicIPAddress, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.PublicIPAddress, error)
	Delete(ctx context.Context, resourceGroupName, publicIPAddressName string) error
}

type publicIPAddressesClientImpl struct {
	c *network.PublicIPAddressesClient
}

var _ PublicIPAddressesClient = &publicIPAddressesClientImpl{}

func (c *publicIPAddressesClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, publicIPAddressName string, parameters network.PublicIPAddress) (*network.PublicIPAddress, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, publicIPAddressName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating public ip address: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for public ip address create/update completion: %w", err)
	}
	return &resp.PublicIPAddress, err
}

func (c *publicIPAddressesClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.PublicIPAddress, error) {
	var l []*network.PublicIPAddress
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing public ip addresses: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *publicIPAddressesClientImpl) Delete(ctx context.Context, resourceGroupName, publicIPAddressName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, publicIPAddressName, nil)
	if err != nil {
		return fmt.Errorf("deleting public ip address: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for public ip address deletion completion: %w", err)
	}
	return nil
}

func newPublicIPAddressesClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*publicIPAddressesClientImpl, error) {
	c, err := network.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating public ip addresses client: %w", err)
	}
	return &publicIPAddressesClientImpl{
		c: c,
	}, nil
}
