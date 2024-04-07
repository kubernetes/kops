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
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// NatGatewaysClient is a client for managing Nat Gateways.
type NatGatewaysClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, natGatewayName string, parameters network.NatGateway) (*network.NatGateway, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.NatGateway, error)
	Delete(ctx context.Context, resourceGroupName, natGatewayName string) error
}

type NatGatewaysClientImpl struct {
	c *network.NatGatewaysClient
}

var _ NatGatewaysClient = &NatGatewaysClientImpl{}

func (c *NatGatewaysClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, natGatewayName string, parameters network.NatGateway) (*network.NatGateway, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, natGatewayName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating nat gateway: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for nat gateway create/update: %w", err)
	}
	return &resp.NatGateway, err
}

func (c *NatGatewaysClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.NatGateway, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*network.NatGateway
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing nat gateways: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *NatGatewaysClientImpl) Delete(ctx context.Context, resourceGroupName, natGatewayName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, natGatewayName, nil)
	if err != nil {
		return fmt.Errorf("deleting nat gateway: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for nat gateway deletion completion: %w", err)
	}
	return nil
}

func newNatGatewaysClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*NatGatewaysClientImpl, error) {
	c, err := network.NewNatGatewaysClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating nat gateways client: %w", err)
	}
	return &NatGatewaysClientImpl{
		c: c,
	}, nil
}
