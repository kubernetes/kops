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
)

type mockClient struct {
	vmsses map[string]compute.VirtualMachineScaleSet
}

var _ vmssGetter = &mockClient{}

func (c *mockClient) getVMScaleSet(ctx context.Context, vmssName string) (compute.VirtualMachineScaleSet, error) {
	vmss, ok := c.vmsses[vmssName]
	if !ok {
		return compute.VirtualMachineScaleSet{}, fmt.Errorf("no VM ScaleSet found for %s", vmssName)
	}
	return vmss, nil
}
