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
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestSubnetRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	subnet := &Subnet{}
	expected := &Subnet{
		Name: to.StringPtr("vnet"),
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		VirtualNetwork: &VirtualNetwork{
			Name: to.StringPtr("vnet"),
		},
		CIDR: to.StringPtr("10.0.0.0/8"),
	}
	if err := subnet.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.SubnetsClient.Subnets[*expected.Name]
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.AddressPrefix, *expected.CIDR; a != e {
		t.Errorf("unexpected CIDR: expected %s, but got %s", e, a)
	}
}

func TestSubnetFind(t *testing.T) {
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
	subnet := &Subnet{
		Name:           to.StringPtr("sub"),
		ResourceGroup:  rg,
		VirtualNetwork: vnet,
	}
	// Find will return nothing if there is no Subnet created.
	actual, err := subnet.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected subnet found: %+v", actual)
	}

	// Create a Subnet.
	cidr := "10.0.0.0/8"
	subnetParameters := network.Subnet{
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr(cidr),
		},
	}
	if err := cloud.Subnet().CreateOrUpdate(context.Background(), *rg.Name, *vnet.Name, *subnet.Name, subnetParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = subnet.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *subnet.Name; a != e {
		t.Errorf("unexpected Virtual Network name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.ResourceGroup.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.VirtualNetwork.Name, *vnet.Name; a != e {
		t.Errorf("unexpected Virtual Network name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.CIDR, cidr; a != e {
		t.Errorf("unexpected CIDR: expected %s, but got %s", e, a)
	}
}

func TestSubnetCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *Subnet
		success       bool
	}{
		{
			a:       nil,
			e:       &Subnet{Name: to.StringPtr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &Subnet{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &Subnet{Name: to.StringPtr("name")},
			changes: &Subnet{Name: nil},
			success: true,
		},
		{
			a:       &Subnet{Name: to.StringPtr("name")},
			changes: &Subnet{Name: to.StringPtr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			subnet := Subnet{}
			err := subnet.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
