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

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// VMScaleSetVMsClient is a client for managing VMs in VM Scale Sets.
type VMScaleSetVMsClient interface {
	List(ctx context.Context, resourceGroupName, vmssName string) ([]*compute.VirtualMachineScaleSetVM, error)
	Delete(ctx context.Context, resourceGroupName, vmssName, instanceId string) error
}

type vmScaleSetVMsClientImpl struct {
	c *compute.VirtualMachineScaleSetVMsClient
}

var _ VMScaleSetVMsClient = &vmScaleSetVMsClientImpl{}

func (c *vmScaleSetVMsClientImpl) List(ctx context.Context, resourceGroupName, vmssName string) ([]*compute.VirtualMachineScaleSetVM, error) {
	var l []*compute.VirtualMachineScaleSetVM
	pager := c.c.NewListPager(resourceGroupName, vmssName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing VMSS VMs: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *vmScaleSetVMsClientImpl) Delete(ctx context.Context, resourceGroupName, vmssName, instanceId string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, vmssName, instanceId, nil)
	if err != nil {
		return fmt.Errorf("deleting VMSS VM: %w", err)
	}
	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for VMSS VM deletion completion: %w", err)
	}
	return nil
}

func newVMScaleSetVMsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*vmScaleSetVMsClientImpl, error) {
	c, err := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSS VMs client: %w", err)
	}
	return &vmScaleSetVMsClientImpl{
		c: c,
	}, nil
}
