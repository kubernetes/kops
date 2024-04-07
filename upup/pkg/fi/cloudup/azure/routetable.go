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

// RouteTablesClient is a client for managing route tables.
type RouteTablesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, routeTableName string, parameters network.RouteTable) (*network.RouteTable, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.RouteTable, error)
	Delete(ctx context.Context, resourceGroupName, vnetName string) error
}

type routeTablesClientImpl struct {
	c *network.RouteTablesClient
}

var _ RouteTablesClient = &routeTablesClientImpl{}

func (c *routeTablesClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, routeTableName string, parameters network.RouteTable) (*network.RouteTable, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, routeTableName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating route table: %w", err)
	}
	rt, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for route table create/update completion: %w", err)
	}
	return &rt.RouteTable, err
}

func (c *routeTablesClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.RouteTable, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*network.RouteTable
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing route tables: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *routeTablesClientImpl) Delete(ctx context.Context, resourceGroupName, vnetName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, vnetName, nil)
	if err != nil {
		return fmt.Errorf("deleting route table: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for route table deletion completion: %w", err)
	}
	return nil
}

func newRouteTablesClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*routeTablesClientImpl, error) {
	c, err := network.NewRouteTablesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating route tables client: %w", err)
	}
	return &routeTablesClientImpl{
		c: c,
	}, nil
}
