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

func TestVirtualNetworkRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	vnet := &VirtualNetwork{}
	expected := &VirtualNetwork{
		Name: to.StringPtr("vnet"),
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		CIDR: to.StringPtr("10.0.0.0/8"),
		Tags: map[string]*string{
			"key": to.StringPtr("val"),
		},
	}
	if err := vnet.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.VirtualNetworksClient.VNets[*expected.Name]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Fatalf("unexpected location: expected %s, but got %s", e, a)
	}
	addrPrefixes := *actual.AddressSpace.AddressPrefixes
	if a, e := len(addrPrefixes), 1; a != e {
		t.Fatalf("unexpected number of addess prefixes: expected %d, but got %d", e, a)
	}
	if a, e := addrPrefixes[0], *expected.CIDR; a != e {
		t.Errorf("unexpected CIDR: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, expected.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}
}

func TestVirtualNetworkFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud: cloud,
	}

	rg := &ResourceGroup{
		Name: to.StringPtr("rg"),
	}
	vnet := &VirtualNetwork{
		Name:          to.StringPtr("vnet"),
		ResourceGroup: rg,
	}
	// Find will return nothing if there is no Virtual Network created.
	actual, err := vnet.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected vnet found: %+v", actual)
	}

	// Create a Virtual Network.
	cidr := "10.0.0.0/8"
	vnetParameters := network.VirtualNetwork{
		Location: to.StringPtr(cloud.Location),
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{cidr},
			},
		},
		Tags: map[string]*string{
			"key": to.StringPtr("val"),
		},
	}
	if err := cloud.VirtualNetwork().CreateOrUpdate(context.Background(), *rg.Name, *vnet.Name, vnetParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = vnet.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *vnet.Name; a != e {
		t.Errorf("unexpected Virtual Network name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.ResourceGroup.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.CIDR, cidr; a != e {
		t.Errorf("unexpected CIDR: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, vnetParameters.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}
}

func TestVirtualNetworkRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud:  cloud,
		Target: azure.NewAzureAPITarget(cloud),
	}

	const (
		key = "key"
		val = "val"
	)
	vnet := &VirtualNetwork{
		Name: to.StringPtr("rg"),
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		CIDR: to.StringPtr("10.0.0.0/8"),
		Tags: map[string]*string{
			key: to.StringPtr(val),
		},
	}
	err := vnet.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	e := map[string]*string{
		azure.TagClusterName: to.StringPtr(testClusterName),
		key:                  to.StringPtr(val),
	}
	if a := vnet.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %+v, but got %+v", e, a)
	}
}

func TestVirtualNetworkCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *VirtualNetwork
		success       bool
	}{
		{
			a:       nil,
			e:       &VirtualNetwork{Name: to.StringPtr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &VirtualNetwork{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &VirtualNetwork{Name: to.StringPtr("name")},
			changes: &VirtualNetwork{Name: nil},
			success: true,
		},
		{
			a:       &VirtualNetwork{Name: to.StringPtr("name")},
			changes: &VirtualNetwork{Name: to.StringPtr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			vnet := VirtualNetwork{}
			err := vnet.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
