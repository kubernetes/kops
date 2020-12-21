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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

// VMScaleSetVMsClient is a client for managing VMs in VM Scale Sets.
type VMScaleSetVMsClient interface {
	List(ctx context.Context, resourceGroupName, vmssName string) ([]compute.VirtualMachineScaleSetVM, error)
}

type vmScaleSetVMsClientImpl struct {
	c *compute.VirtualMachineScaleSetVMsClient
}

var _ VMScaleSetVMsClient = &vmScaleSetVMsClientImpl{}

func (c *vmScaleSetVMsClientImpl) List(ctx context.Context, resourceGroupName, vmssName string) ([]compute.VirtualMachineScaleSetVM, error) {
	var l []compute.VirtualMachineScaleSetVM
	for iter, err := c.c.ListComplete(ctx, resourceGroupName, vmssName, "", "", ""); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func newVMScaleSetVMsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *vmScaleSetVMsClientImpl {
	c := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID)
	c.Authorizer = authorizer
	return &vmScaleSetVMsClientImpl{
		c: &c,
	}
}
