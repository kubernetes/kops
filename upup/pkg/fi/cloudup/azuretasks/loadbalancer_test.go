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

package azuretasks

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func newTestLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		Name:      to.StringPtr("loadbalancer"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		Subnet: &Subnet{
			Name:      to.StringPtr("subnet"),
			Lifecycle: fi.LifecycleSync,
			VirtualNetwork: &VirtualNetwork{
				Name: to.StringPtr("vnet"),
			},
		},
		External:     to.BoolPtr(true),
		ForAPIServer: true,
		Tags: map[string]*string{
			testTagKey: to.StringPtr(testTagValue),
		},
	}
}

func TestLoadBalancerRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	loadbalancer := &LoadBalancer{}
	expected := newTestLoadBalancer()
	if err := loadbalancer.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.LoadBalancersClient.LBs[*expected.Name]
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected Name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Fatalf("unexpected location: expected %s, but got %s", e, a)
	}
}

func TestLoadBalancerFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud: cloud,
	}

	rg := &ResourceGroup{
		Name: to.StringPtr("rg"),
	}
	loadBalancer := &LoadBalancer{
		Name: to.StringPtr("loadbalancer"),
		ResourceGroup: &ResourceGroup{
			Name: rg.Name,
		},
	}
	// Find will return nothing if there is no Loadbalancer created.
	actual, err := loadBalancer.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected loadbalancer found: %+v", actual)
	}

	feConfigProperties := &network.FrontendIPConfigurationPropertiesFormat{
		PrivateIPAllocationMethod: network.Dynamic,
		Subnet: &network.Subnet{
			Name: to.StringPtr("subnet"),
			ID:   to.StringPtr("id"),
		},
	}
	feConfigProperties.PublicIPAddress = &network.PublicIPAddress{
		ID: to.StringPtr("id"),
	}

	// Create a Loadbalancer.
	loadBalancerParameters := network.LoadBalancer{
		Location: to.StringPtr("eastus"),
		LoadBalancerPropertiesFormat: &network.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: &[]network.FrontendIPConfiguration{
				{
					FrontendIPConfigurationPropertiesFormat: feConfigProperties,
				},
			},
		},
	}
	if err := cloud.LoadBalancer().CreateOrUpdate(context.Background(), *rg.Name, *loadBalancer.Name, loadBalancerParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = loadBalancer.Find(ctx)
	t.Log(actual)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *loadBalancer.Name; a != e {
		t.Errorf("unexpected Loadbalancer name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.ResourceGroup.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Subnet.Name, *feConfigProperties.Subnet.Name; a != e {
		t.Errorf("unexpected Subnet name: expected %s, but got %s", e, a)
	}
	if !*actual.External {
		t.Errorf("unexpected require public IP")
	}
}

func TestLoadBalancerRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud:  cloud,
		Target: azure.NewAzureAPITarget(cloud),
	}

	lb := newTestLoadBalancer()
	err := lb.Normalize(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = lb.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	e := map[string]*string{
		azure.TagClusterName: to.StringPtr(testClusterName),
		testTagKey:           to.StringPtr(testTagValue),
	}
	if a := lb.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %+v, but got %+v", e, a)
	}
}

func TestLoadBalancerCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *LoadBalancer
		success       bool
	}{
		{
			a:       nil,
			e:       &LoadBalancer{Name: to.StringPtr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &LoadBalancer{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &LoadBalancer{Name: to.StringPtr("name")},
			changes: &LoadBalancer{Name: nil},
			success: true,
		},
		{
			a:       &LoadBalancer{Name: to.StringPtr("name")},
			changes: &LoadBalancer{Name: to.StringPtr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			loadBalancer := LoadBalancer{}
			err := loadBalancer.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
