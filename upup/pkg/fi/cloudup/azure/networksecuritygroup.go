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
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// NetworkSecurityGroupsClient is a client for managing network security groups.
type NetworkSecurityGroupsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string, parameters network.SecurityGroup) (*network.SecurityGroup, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.SecurityGroup, error)
	Delete(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string) error
}

type NetworkSecurityGroupsClientImpl struct {
	c *network.SecurityGroupsClient
}

var _ NetworkSecurityGroupsClient = &NetworkSecurityGroupsClientImpl{}

func (c *NetworkSecurityGroupsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string, parameters network.SecurityGroup) (*network.SecurityGroup, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, NetworkSecurityGroupName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating network security group: %w", err)
	}
	asg, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for network security group create/update completion: %w", err)
	}
	return &asg.SecurityGroup, err
}

func (c *NetworkSecurityGroupsClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.SecurityGroup, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*network.SecurityGroup
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing network security groups: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *NetworkSecurityGroupsClientImpl) Delete(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, NetworkSecurityGroupName, nil)
	if err != nil {
		return fmt.Errorf("deleting network security group: %w", err)
	}
	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for network security group deletion completion: %w", err)
	}
	return nil
}

func newNetworkSecurityGroupsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*NetworkSecurityGroupsClientImpl, error) {
	c, err := network.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating network security groups client: %w", err)
	}
	return &NetworkSecurityGroupsClientImpl{
		c: c,
	}, nil
}
