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

// ApplicationSecurityGroupsClient is a client for managing application security groups.
type ApplicationSecurityGroupsClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, applicationSecurityGroupName string, parameters network.ApplicationSecurityGroup) (*network.ApplicationSecurityGroup, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.ApplicationSecurityGroup, error)
	Delete(ctx context.Context, resourceGroupName, applicationSecurityGroupName string) error
}

type ApplicationSecurityGroupsClientImpl struct {
	c *network.ApplicationSecurityGroupsClient
}

var _ ApplicationSecurityGroupsClient = &ApplicationSecurityGroupsClientImpl{}

func (c *ApplicationSecurityGroupsClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, applicationSecurityGroupName string, parameters network.ApplicationSecurityGroup) (*network.ApplicationSecurityGroup, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, applicationSecurityGroupName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating application security group: %w", err)
	}
	asg, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for application security group create/update completion: %w", err)
	}
	return &asg.ApplicationSecurityGroup, err
}

func (c *ApplicationSecurityGroupsClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.ApplicationSecurityGroup, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*network.ApplicationSecurityGroup
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing application security groups: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *ApplicationSecurityGroupsClientImpl) Delete(ctx context.Context, resourceGroupName, applicationSecurityGroupName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, applicationSecurityGroupName, nil)
	if err != nil {
		return fmt.Errorf("deleting application security group: %w", err)
	}
	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for application security group deletion completion: %w", err)
	}
	return nil
}

func newApplicationSecurityGroupsClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*ApplicationSecurityGroupsClientImpl, error) {
	c, err := network.NewApplicationSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating application security groups client: %w", err)
	}
	return &ApplicationSecurityGroupsClientImpl{
		c: c,
	}, nil
}
