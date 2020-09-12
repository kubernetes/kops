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

// VMScaleSetsClient is a client for managing VM Scale Set.
type VMScaleSetsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, vmScaleSetName string, parameters compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error)
	List(ctx context.Context, resourceGroupName string) ([]compute.VirtualMachineScaleSet, error)
	Delete(ctx context.Context, resourceGroupName, vmssName string) error
}

type vmScaleSetsClientImpl struct {
	c *compute.VirtualMachineScaleSetsClient
}

var _ VMScaleSetsClient = &vmScaleSetsClientImpl{}

func (c *vmScaleSetsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, vmScaleSetName string, parameters compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error) {
	future, err := c.c.CreateOrUpdate(ctx, resourceGroupName, vmScaleSetName, parameters)
	if err != nil {
		return nil, fmt.Errorf("error creating/updating VM Scale Set: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return nil, fmt.Errorf("error waiting for VM Scale Set create/update completion: %s", err)
	}
	vmss, err := future.Result(*c.c)
	if err != nil {
		return nil, fmt.Errorf("error obtaining result for VM Scale Set create/update: %s", err)
	}
	return &vmss, nil
}

func (c *vmScaleSetsClientImpl) List(ctx context.Context, resourceGroupName string) ([]compute.VirtualMachineScaleSet, error) {
	var l []compute.VirtualMachineScaleSet
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *vmScaleSetsClientImpl) Delete(ctx context.Context, resourceGroupName, vmssName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, vmssName)
	if err != nil {
		return fmt.Errorf("error deleting VM Scale Set: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for VM Scale Set deletion completion: %s", err)
	}
	return nil
}

func newVMScaleSetsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *vmScaleSetsClientImpl {
	c := compute.NewVirtualMachineScaleSetsClient(subscriptionID)
	c.Authorizer = authorizer
	return &vmScaleSetsClientImpl{
		c: &c,
	}
}
