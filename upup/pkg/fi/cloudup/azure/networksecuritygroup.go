/*
Copyright 2023 The Kubernetes Authors.

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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest"
)

// NetworkSecurityGroupsClient is a client for managing Network Security Groups.
type NetworkSecurityGroupsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string, parameters network.SecurityGroup) (*network.SecurityGroup, error)
	List(ctx context.Context, resourceGroupName string) ([]network.SecurityGroup, error)
	Delete(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string) error
}

type NetworkSecurityGroupsClientImpl struct {
	c *network.SecurityGroupsClient
}

var _ NetworkSecurityGroupsClient = &NetworkSecurityGroupsClientImpl{}

func (c *NetworkSecurityGroupsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string, parameters network.SecurityGroup) (*network.SecurityGroup, error) {
	future, err := c.c.CreateOrUpdate(ctx, resourceGroupName, NetworkSecurityGroupName, parameters)
	if err != nil {
		return nil, fmt.Errorf("error creating/updating Network Security Group: %w", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return nil, fmt.Errorf("error waiting for Network Security Group create/update completion: %w", err)
	}
	asg, err := future.Result(*c.c)
	if err != nil {
		return nil, fmt.Errorf("error obtaining result for Network Security Group create/update: %w", err)
	}
	return &asg, err
}

func (c *NetworkSecurityGroupsClientImpl) List(ctx context.Context, resourceGroupName string) ([]network.SecurityGroup, error) {
	var l []network.SecurityGroup
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *NetworkSecurityGroupsClientImpl) Delete(ctx context.Context, resourceGroupName, NetworkSecurityGroupName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, NetworkSecurityGroupName)
	if err != nil {
		return fmt.Errorf("error deleting Network Security Group: %w", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for Network Security Group deletion completion: %w", err)
	}
	return nil
}

func newNetworkSecurityGroupsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *NetworkSecurityGroupsClientImpl {
	c := network.NewSecurityGroupsClient(subscriptionID)
	c.Authorizer = authorizer
	return &NetworkSecurityGroupsClientImpl{
		c: &c,
	}
}
