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
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// NetworkInterfacesClient is a client for managing Network Interfaces.
type NetworkInterfacesClient interface {
	ListScaleSetsNetworkInterfaces(ctx context.Context, resourceGroupName, vmssName string) ([]*network.Interface, error)
}

type networkInterfacesClientImpl struct {
	c *network.InterfacesClient
}

var _ NetworkInterfacesClient = &networkInterfacesClientImpl{}

func (c *networkInterfacesClientImpl) ListScaleSetsNetworkInterfaces(ctx context.Context, resourceGroupName, vmssName string) ([]*network.Interface, error) {
	var l []*network.Interface
	pager := c.c.NewListVirtualMachineScaleSetNetworkInterfacesPager(resourceGroupName, vmssName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing network interfaces: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func newNetworkInterfacesClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*networkInterfacesClientImpl, error) {
	c, err := network.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating network interfaces client: %w", err)
	}
	return &networkInterfacesClientImpl{
		c: c,
	}, nil
}
