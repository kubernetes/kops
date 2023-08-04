/*
Copyright 2023 The Kubernetes Authors.

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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest"
)

// NatGatewaysClient is a client for managing Nat Gateways.
type NatGatewaysClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, natGatewayName string, parameters network.NatGateway) (*network.NatGateway, error)
	List(ctx context.Context, resourceGroupName string) ([]network.NatGateway, error)
	Delete(ctx context.Context, resourceGroupName, natGatewayName string) error
}

type NatGatewaysClientImpl struct {
	c *network.NatGatewaysClient
}

var _ NatGatewaysClient = &NatGatewaysClientImpl{}

func (c *NatGatewaysClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, natGatewayName string, parameters network.NatGateway) (*network.NatGateway, error) {
	future, err := c.c.CreateOrUpdate(ctx, resourceGroupName, natGatewayName, parameters)
	if err != nil {
		return nil, fmt.Errorf("creating/updating nat gateway: %w", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return nil, fmt.Errorf("waiting for nat gateway create/update completion: %w", err)
	}
	asg, err := future.Result(*c.c)
	if err != nil {
		return nil, fmt.Errorf("obtaining result for nat gateway create/update: %w", err)
	}
	return &asg, err
}

func (c *NatGatewaysClientImpl) List(ctx context.Context, resourceGroupName string) ([]network.NatGateway, error) {
	var l []network.NatGateway
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.NextWithContext(ctx) {
		if err != nil {
			return nil, fmt.Errorf("listing nat gateways: %w", err)
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *NatGatewaysClientImpl) Delete(ctx context.Context, resourceGroupName, natGatewayName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, natGatewayName)
	if err != nil {
		return fmt.Errorf("deleting nat gateway: %w", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("waiting for nat gateway deletion completion: %w", err)
	}
	return nil
}

func newNatGatewaysClientImpl(subscriptionID string, authorizer autorest.Authorizer) *NatGatewaysClientImpl {
	c := network.NewNatGatewaysClient(subscriptionID)
	c.Authorizer = authorizer
	return &NatGatewaysClientImpl{
		c: &c,
	}
}
