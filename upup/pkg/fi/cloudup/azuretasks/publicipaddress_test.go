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

func newTestPublicIPAddress() *PublicIPAddress {
	return &PublicIPAddress{
		Name:      to.StringPtr("publicIPAddress"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		Tags: map[string]*string{
			testTagKey: to.StringPtr(testTagValue),
		},
	}
}

func TestPublicIPAddressRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	publicIPAddress := &PublicIPAddress{}
	expected := newTestPublicIPAddress()
	if err := publicIPAddress.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.PublicIPAddressesClient.PubIPs[*expected.Name]
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected Name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Fatalf("unexpected location: expected %s, but got %s", e, a)
	}
}

func TestPublicIPAddressFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud: cloud,
	}

	rg := &ResourceGroup{
		Name: to.StringPtr("rg"),
	}
	publicIPAddress := &PublicIPAddress{
		Name: to.StringPtr("publicIPAddress"),
		ResourceGroup: &ResourceGroup{
			Name: rg.Name,
		},
	}
	// Find will return nothing if there is no public ip address created.
	actual, err := publicIPAddress.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected publicIPAddress found: %+v", actual)
	}

	// Create a public ip address.
	publicIPAddressParameters := network.PublicIPAddress{
		Location: to.StringPtr("eastus"),
		Name:     to.StringPtr("publicIPAddress"),
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAddressVersion:   network.IPv4,
			PublicIPAllocationMethod: network.Dynamic,
		},
	}
	if err := cloud.PublicIPAddress().CreateOrUpdate(context.Background(), *rg.Name, *publicIPAddress.Name, publicIPAddressParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = publicIPAddress.Find(ctx)
	t.Log(actual)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *publicIPAddress.Name; a != e {
		t.Errorf("unexpected publicIPAddress name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.ResourceGroup.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
}

func TestPublicIPAddressRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud:  cloud,
		Target: azure.NewAzureAPITarget(cloud),
	}

	lb := newTestPublicIPAddress()
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

func TestPublicIPAddressCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *PublicIPAddress
		success       bool
	}{
		{
			a:       nil,
			e:       &PublicIPAddress{Name: to.StringPtr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &PublicIPAddress{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &PublicIPAddress{Name: to.StringPtr("name")},
			changes: &PublicIPAddress{Name: nil},
			success: true,
		},
		{
			a:       &PublicIPAddress{Name: to.StringPtr("name")},
			changes: &PublicIPAddress{Name: to.StringPtr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			publicIPAddress := PublicIPAddress{}
			err := publicIPAddress.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
