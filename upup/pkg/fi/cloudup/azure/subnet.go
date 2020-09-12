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

// SubnetsClient is a client for managing Subnets.
type SubnetsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string, parameters network.Subnet) error
	List(ctx context.Context, resourceGroupName, virtualNetworkName string) ([]network.Subnet, error)
	Delete(ctx context.Context, resourceGroupName, vnetName, subnetName string) error
}

type subnetsClientImpl struct {
	c *network.SubnetsClient
}

var _ SubnetsClient = &subnetsClientImpl{}

func (c *subnetsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string, parameters network.Subnet) error {
	_, err := c.c.CreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, parameters)
	return err
}

func (c *subnetsClientImpl) List(ctx context.Context, resourceGroupName, virtualNetworkName string) ([]network.Subnet, error) {
	var l []network.Subnet
	for iter, err := c.c.ListComplete(ctx, resourceGroupName, virtualNetworkName); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *subnetsClientImpl) Delete(ctx context.Context, resourceGroupName, vnetName, subnetName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, vnetName, subnetName)
	if err != nil {
		return fmt.Errorf("error deleting subnet: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for subnet deletion completion: %s", err)
	}
	return nil
}

func newSubnetsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *subnetsClientImpl {
	c := network.NewSubnetsClient(subscriptionID)
	c.Authorizer = authorizer
	return &subnetsClientImpl{
		c: &c,
	}
}
