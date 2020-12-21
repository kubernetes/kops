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

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources"
	"github.com/Azure/go-autorest/autorest"
)

// ResourceGroupsClient is a client for managing Resource Groups.
type ResourceGroupsClient interface {
	CreateOrUpdate(ctx context.Context, name string, parameters resources.Group) error
	List(ctx context.Context, filter string) ([]resources.Group, error)
	Delete(ctx context.Context, name string) error
}

type resourceGroupsClientImpl struct {
	c *resources.GroupsClient
}

var _ ResourceGroupsClient = &resourceGroupsClientImpl{}

func (c *resourceGroupsClientImpl) CreateOrUpdate(ctx context.Context, name string, parameters resources.Group) error {
	_, err := c.c.CreateOrUpdate(ctx, name, parameters)
	return err
}

func (c *resourceGroupsClientImpl) List(ctx context.Context, filter string) ([]resources.Group, error) {
	var l []resources.Group
	for iter, err := c.c.ListComplete(ctx, filter, nil /* top */); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *resourceGroupsClientImpl) Delete(ctx context.Context, name string) error {
	future, err := c.c.Delete(ctx, name)
	if err != nil {
		return fmt.Errorf("error deleting resource group: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for resource group deletion completion: %s", err)
	}
	return nil
}

func newResourceGroupsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *resourceGroupsClientImpl {
	c := resources.NewGroupsClient(subscriptionID)
	c.Authorizer = authorizer
	return &resourceGroupsClientImpl{
		c: &c,
	}
}
