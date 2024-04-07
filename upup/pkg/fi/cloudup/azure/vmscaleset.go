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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// VMScaleSetsClient is a client for managing VMSSs.
type VMScaleSetsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, vmScaleSetName string, parameters compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error)
	List(ctx context.Context, resourceGroupName string) ([]*compute.VirtualMachineScaleSet, error)
	Get(ctx context.Context, resourceGroupName string, vmssName string) (*compute.VirtualMachineScaleSet, error)
	Delete(ctx context.Context, resourceGroupName, vmssName string) error
}

type vmScaleSetsClientImpl struct {
	c *compute.VirtualMachineScaleSetsClient
}

var _ VMScaleSetsClient = &vmScaleSetsClientImpl{}

func (c *vmScaleSetsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, vmScaleSetName string, parameters compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, vmScaleSetName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating VMSS: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for VMSS create/update: %w", err)
	}
	return &resp.VirtualMachineScaleSet, nil
}

func (c *vmScaleSetsClientImpl) List(ctx context.Context, resourceGroupName string) ([]*compute.VirtualMachineScaleSet, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*compute.VirtualMachineScaleSet
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing VMSSs: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *vmScaleSetsClientImpl) Get(ctx context.Context, resourceGroupName string, vmssName string) (*compute.VirtualMachineScaleSet, error) {
	opts := &compute.VirtualMachineScaleSetsClientGetOptions{
		Expand: to.Ptr(compute.ExpandTypesForGetVMScaleSetsUserData),
	}
	resp, err := c.c.Get(ctx, resourceGroupName, vmssName, opts)
	if err != nil {
		return nil, fmt.Errorf("getting VMSS: %w", err)
	}
	return &resp.VirtualMachineScaleSet, nil
}

func (c *vmScaleSetsClientImpl) Delete(ctx context.Context, resourceGroupName, vmssName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, vmssName, nil)
	if err != nil {
		return fmt.Errorf("deleting VMSS: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for VMSS deletion completion: %w", err)
	}
	return nil
}

func newVMScaleSetsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*vmScaleSetsClientImpl, error) {
	c, err := compute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSSs client: %w", err)
	}
	return &vmScaleSetsClientImpl{
		c: c,
	}, nil
}
