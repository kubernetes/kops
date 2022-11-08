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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest"
)

// LoadBalancersClient is a client for connecting to the kubernetes api.
type LoadBalancersClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName, loadBalancerName string, parameters network.LoadBalancer) error
	List(ctx context.Context, resourceGroupName string) ([]network.LoadBalancer, error)
	Delete(ctx context.Context, resourceGroupName, loadBalancerName string) error
}

type loadBalancersClientImpl struct {
	c *network.LoadBalancersClient
}

var _ LoadBalancersClient = &loadBalancersClientImpl{}

func (c *loadBalancersClientImpl) CreateOrUpdate(ctx context.Context, resourceGroupName, loadBalancerName string, parameters network.LoadBalancer) error {
	_, err := c.c.CreateOrUpdate(ctx, resourceGroupName, loadBalancerName, parameters)
	return err
}

func (c *loadBalancersClientImpl) List(ctx context.Context, resourceGroupName string) ([]network.LoadBalancer, error) {
	var l []network.LoadBalancer
	for iter, err := c.c.ListComplete(ctx, resourceGroupName); iter.NotDone(); err = iter.Next() {
		if err != nil {
			return nil, err
		}
		l = append(l, iter.Value())
	}
	return l, nil
}

func (c *loadBalancersClientImpl) Delete(ctx context.Context, resourceGroupName, loadBalancerName string) error {
	future, err := c.c.Delete(ctx, resourceGroupName, loadBalancerName)
	if err != nil {
		return fmt.Errorf("error deleting loadbalancer: %s", err)
	}
	if err := future.WaitForCompletionRef(ctx, c.c.Client); err != nil {
		return fmt.Errorf("error waiting for loadbalancer deletion completion: %s", err)
	}
	return nil
}

func newLoadBalancersClientImpl(subscriptionID string, authorizer autorest.Authorizer) *loadBalancersClientImpl {
	c := network.NewLoadBalancersClient(subscriptionID)
	c.Authorizer = authorizer
	return &loadBalancersClientImpl{
		c: &c,
	}
}
