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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest"
)

// PublicIPAddressesClient is a client for public ip addresses.
type PublicIPAddressesClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, publicIPAddressName string, parameters network.PublicIPAddress) error
	List(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, error)
	Delete(ctx context.Context, resourceGroupName, publicIPAddressName string) error
}

type publicIPAddressesClientImpl struct {
	c *network.PublicIPAddressesClient
}

var _ PublicIPAddressesClient = &publicIPAddressesClientImpl{}

func (c *publicIPAddressesClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, publicIPAddressName string, parameters network.PublicIPAddress) error {
	_, err := c.c.CreateOrUpdate(ctx, resourceGroupName, publicIPAddressName, parameters)
	return err
}

func (c *publicIPAddressesClientImpl) List(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, error) {
	var l []network.PublicIPAddress
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *publicIPAddressesClientImpl) Delete(ctx context.Context, resourceGroupName, publicIPAddressName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, publicIPAddressName)
	if err != nil {
		return fmt.Errorf("error deleting public ip address: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for public ip address deletion completion: %s", err)
	}
	return nil
}

func newPublicIPAddressesClientImpl(subscriptionID string, authorizer autorest.Authorizer) *publicIPAddressesClientImpl {
	c := network.NewPublicIPAddressesClient(subscriptionID)
	c.Authorizer = authorizer
	return &publicIPAddressesClientImpl{
		c: &c,
	}
}
