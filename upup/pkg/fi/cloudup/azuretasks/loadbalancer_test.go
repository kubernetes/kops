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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func newTestLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		Name:      to.Ptr("loadbalancer"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.Ptr("rg"),
		},
		Subnet: &Subnet{
			Name:      to.Ptr("subnet"),
			Lifecycle: fi.LifecycleSync,
			VirtualNetwork: &VirtualNetwork{
				Name: to.Ptr("vnet"),
			},
		},
		External:     to.Ptr(true),
		ForAPIServer: true,
		Tags: map[string]*string{
			testTagKey: to.Ptr(testTagValue),
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
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	rg := &ResourceGroup{
		Name: to.Ptr("rg"),
	}
	loadBalancer := &LoadBalancer{
		Name: to.Ptr("loadbalancer"),
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
		PrivateIPAllocationMethod: to.Ptr(network.IPAllocationMethodDynamic),
		Subnet: &network.Subnet{
			Name: to.Ptr("subnet"),
			ID:   to.Ptr("id"),
		},
	}
	feConfigProperties.PublicIPAddress = &network.PublicIPAddress{
		ID: to.Ptr("id"),
	}

	// Create a Loadbalancer.
	loadBalancerParameters := network.LoadBalancer{
		Location: to.Ptr("eastus"),
		Properties: &network.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*network.FrontendIPConfiguration{
				{
					Properties: feConfigProperties,
				},
			},
		},
	}
	_, err = cloud.LoadBalancer().CreateOrUpdate(context.Background(), *rg.Name, *loadBalancer.Name, loadBalancerParameters)
	if err != nil {
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
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
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
		azure.TagClusterName: to.Ptr(testClusterName),
		testTagKey:           to.Ptr(testTagValue),
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
			e:       &LoadBalancer{Name: to.Ptr("name")},
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
			a:       &LoadBalancer{Name: to.Ptr("name")},
			changes: &LoadBalancer{Name: nil},
			success: true,
		},
		{
			a:       &LoadBalancer{Name: to.Ptr("name")},
			changes: &LoadBalancer{Name: to.Ptr("newName")},
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
