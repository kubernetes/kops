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

// ApplicationSecurityGroupsClient is a client for managing Application Security Groups.
type ApplicationSecurityGroupsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, applicationSecurityGroupName string, parameters network.ApplicationSecurityGroup) (*network.ApplicationSecurityGroup, error)
	List(ctx context.Context, resourceGroupName string) ([]network.ApplicationSecurityGroup, error)
	Delete(ctx context.Context, resourceGroupName, applicationSecurityGroupName string) error
}

type ApplicationSecurityGroupsClientImpl struct {
	c *network.ApplicationSecurityGroupsClient
}

var _ ApplicationSecurityGroupsClient = &ApplicationSecurityGroupsClientImpl{}

func (c *ApplicationSecurityGroupsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, applicationSecurityGroupName string, parameters network.ApplicationSecurityGroup) (*network.ApplicationSecurityGroup, error) {
	future, err := c.c.CreateOrUpdate(ctx, resourceGroupName, applicationSecurityGroupName, parameters)
	if err != nil {
		return nil, fmt.Errorf("creating/updating Application Security Group: %w", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return nil, fmt.Errorf("waiting for Application Security Group create/update completion: %w", err)
	}
	asg, err := future.Result(*c.c)
	if err != nil {
		return nil, fmt.Errorf("obtaining result for Application Security Group create/update: %w", err)
	}
	return &asg, err
}

func (c *ApplicationSecurityGroupsClientImpl) List(ctx context.Context, resourceGroupName string) ([]network.ApplicationSecurityGroup, error) {
	var l []network.ApplicationSecurityGroup
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.NextWithContext(ctx) {
		if err != nil {
			return nil, fmt.Errorf("listing application security groups: %w", err)
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *ApplicationSecurityGroupsClientImpl) Delete(ctx context.Context, resourceGroupName, applicationSecurityGroupName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, applicationSecurityGroupName)
	if err != nil {
		return fmt.Errorf("deleting application security group: %w", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("waiting for application security group deletion completion: %w", err)
	}
	return nil
}

func newApplicationSecurityGroupsClientImpl(subscriptionID string, authorizer autorest.Authorizer) *ApplicationSecurityGroupsClientImpl {
	c := network.NewApplicationSecurityGroupsClient(subscriptionID)
	c.Authorizer = authorizer
	return &ApplicationSecurityGroupsClientImpl{
		c: &c,
	}
}
