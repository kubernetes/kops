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
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// DisksClient is a client for managing disks.
type DisksClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, diskName string, parameters compute.Disk) (*compute.Disk, error)
	List(ctx context.Context, resourceGroupName string) ([]*compute.Disk, error)
	Delete(ctx context.Context, resourceGroupName, diskname string) error
}

type disksClientImpl struct {
	c *compute.DisksClient
}

var _ DisksClient = &disksClientImpl{}

func (c *disksClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, diskName string, parameters compute.Disk) (*compute.Disk, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, diskName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating disk: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for disk create/update completion: %w", err)
	}
	return &resp.Disk, err
}

func (c *disksClientImpl) List(ctx context.Context, resourceGroupName string) ([]*compute.Disk, error) {
	var l []*compute.Disk
	pager := c.c.NewListPager(nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing disks: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *disksClientImpl) Delete(ctx context.Context, resourceGroupName, diskName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, diskName, nil)
	if err != nil {
		return fmt.Errorf("deleting disk: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for disk deletion completion: %w", err)
	}
	return nil
}

func newDisksClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*disksClientImpl, error) {
	c, err := compute.NewDisksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating disks client: %w", err)
	}
	return &disksClientImpl{
		c: c,
	}, nil
}
