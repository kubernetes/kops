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
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// LoadBalancersClient is a client for managing load balancers.
type LoadBalancersClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, loadBalancerName string, parameters network.LoadBalancer) (*network.LoadBalancer, error)
	List(ctx context.Context, resourceGroupName string) ([]*network.LoadBalancer, error)
	Get(ctx context.Context, resourceGroupName string, loadBalancerName string) (*network.LoadBalancer, error)
	Delete(ctx context.Context, resourceGroupName, loadBalancerName string) error
}

type loadBalancersClientImpl struct {
	c *network.LoadBalancersClient
}

var _ LoadBalancersClient = &loadBalancersClientImpl{}

func (c *loadBalancersClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, loadBalancerName string, parameters network.LoadBalancer) (*network.LoadBalancer, error) {
	future, err := c.c.BeginCreateOrUpdate(ctx, resourceGroupName, loadBalancerName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("creating/updating load balancer: %w", err)
	}
	resp, err := future.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for load balancer create/update: %w", err)
	}
	return &resp.LoadBalancer, nil
}

func (c *loadBalancersClientImpl) List(ctx context.Context, resourceGroupName string) ([]*network.LoadBalancer, error) {
	if resourceGroupName == "" {
		return nil, nil
	}

	var l []*network.LoadBalancer
	pager := c.c.NewListPager(resourceGroupName, nil)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) && respErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
			return nil, fmt.Errorf("listing load balancers: %w", err)
		}
		l = append(l, resp.Value...)
	}
	return l, nil
}

func (c *loadBalancersClientImpl) Get(ctx context.Context, resourceGroupName string, loadBalancerName string) (*network.LoadBalancer, error) {
	opts := &network.LoadBalancersClientGetOptions{
		Expand: to.Ptr("frontendIpConfigurations/publicIpAddress"),
	}
	resp, err := c.c.Get(ctx, resourceGroupName, loadBalancerName, opts)
	if err != nil {
		return nil, fmt.Errorf("getting load balancer: %w", err)
	}
	return &resp.LoadBalancer, nil
}

func (c *loadBalancersClientImpl) Delete(ctx context.Context, resourceGroupName, loadBalancerName string) error {
	future, err := c.c.BeginDelete(ctx, resourceGroupName, loadBalancerName, nil)
	if err != nil {
		return fmt.Errorf("deleting load balancer: %w", err)
	}
	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for load balancer deletion completion: %w", err)
	}
	return nil
}

func newLoadBalancersClientImpl(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*loadBalancersClientImpl, error) {
	c, err := network.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating load balancers client: %w", err)
	}
	return &loadBalancersClientImpl{
		c: c,
	}, nil
}
