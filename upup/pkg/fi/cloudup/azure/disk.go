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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

// DisksClient is a client for managing VM Scale Set.
type DisksClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, diskName string, parameters compute.Disk) error
	List(ctx context.Context, resourceGroupName string) ([]compute.Disk, error)
	Delete(ctx context.Context, resourceGroupName, diskname string) error
}

type disksClientImpl struct {
	c *compute.DisksClient
}

var _ DisksClient = &disksClientImpl{}

func (c *disksClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, diskName string, parameters compute.Disk) error {
	_, err := c.c.CreateOrUpdate(ctx, resourceGroupName, diskName, parameters)
	return err
}

func (c *disksClientImpl) List(ctx context.Context, resourceGroupName string) ([]compute.Disk, error) {
	var l []compute.Disk
	for iter, err := c.c.ListByResourceGroupComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *disksClientImpl) Delete(ctx context.Context, resourceGroupName, diskName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, diskName)
	if err != nil {
		return fmt.Errorf("error deleting disk: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for disk deletion completion: %s", err)
	}
	return nil
}

func newDisksClientImpl(subscriptionID string, authorizer autorest.Authorizer) *disksClientImpl {
	c := compute.NewDisksClient(subscriptionID)
	c.Authorizer = authorizer
	return &disksClientImpl{
		c: &c,
	}
}
