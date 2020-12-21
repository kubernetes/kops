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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/go-autorest/autorest"
)

// RouteTablesClient is a client for managing Virtual Networks.
type RouteTablesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, routeTableName string, parameters network.RouteTable) error
	List(ctx context.Context, resourceGroupName string) ([]network.RouteTable, error)
	Delete(ctx context.Context, resourceGroupName, vnetName string) error
}

type routeTablesClientImpl struct {
	c *network.RouteTablesClient
}

var _ RouteTablesClient = &routeTablesClientImpl{}

func (c *routeTablesClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, routeTableName string, parameters network.RouteTable) error {
	_, err := c.c.CreateOrUpdate(ctx, resourceGroupName, routeTableName, parameters)
	return err
}

func (c *routeTablesClientImpl) List(ctx context.Context, resourceGroupName string) ([]network.RouteTable, error) {
	var l []network.RouteTable
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *routeTablesClientImpl) Delete(ctx context.Context, resourceGroupName, vnetName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, vnetName)
	if err != nil {
		return fmt.Errorf("error deleting virtual network: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for virtual network deletion completion: %s", err)
	}
	return nil
}

func newRouteTablesClientImpl(subscriptionID string, authorizer autorest.Authorizer) *routeTablesClientImpl {
	c := network.NewRouteTablesClient(subscriptionID)
	c.Authorizer = authorizer
	return &routeTablesClientImpl{
		c: &c,
	}
}
