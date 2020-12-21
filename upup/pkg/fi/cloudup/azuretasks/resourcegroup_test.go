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

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestResourceGroupRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	rg := &ResourceGroup{}
	expected := &ResourceGroup{
		Name: to.StringPtr("rg"),
		Tags: map[string]*string{
			"key": to.StringPtr("value"),
		},
	}
	if err := rg.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.ResourceGroupsClient.RGs[*expected.Name]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Errorf("unexpected location: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected name: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, expected.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}

	// Call Render again to update tags.
	current := expected
	expected = &ResourceGroup{
		Name: to.StringPtr("rg"),
		Tags: map[string]*string{
			"key2": to.StringPtr("value2"),
		},
	}
	changes := &ResourceGroup{
		Tags: map[string]*string{
			"key2": to.StringPtr("value2"),
		},
	}
	if err := rg.RenderAzure(apiTarget, current, expected, changes); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	actual = cloud.ResourceGroupsClient.RGs[*expected.Name]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Errorf("unexpected location: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected name: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, expected.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}
}

func TestResourceGroupFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud: cloud,
	}

	rg := &ResourceGroup{
		Name:   to.StringPtr("rg"),
		Shared: to.BoolPtr(true),
	}
	// Find will return nothing if there is no Resource Group created.
	actual, err := rg.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected resource group found: %+v", actual)
	}

	// Create a Resource Group.
	rgParameters := resources.Group{
		Location: to.StringPtr(cloud.Location),
		Tags: map[string]*string{
			"key": to.StringPtr("value"),
		},
	}
	if err := cloud.ResourceGroup().CreateOrUpdate(context.Background(), *rg.Name, rgParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = rg.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := actual.Tags, rgParameters.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags expected %+v, but got %+v", e, a)
	}
	if a, e := *actual.Shared, *rg.Shared; a != e {
		t.Errorf("unexpected shared: %+v, but got %+v", e, a)
	}

	// Call Find with an invalid resource group name.
	rg = &ResourceGroup{
		Name: to.StringPtr("invalid"),
	}
	actual, err = rg.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected resource group found: %+v", actual)
	}
}

func TestResourceGroupRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud:  cloud,
		Target: azure.NewAzureAPITarget(cloud),
	}

	const (
		key = "key"
		val = "val"
	)
	rg := &ResourceGroup{
		Name: to.StringPtr("rg"),
		Tags: map[string]*string{
			key: to.StringPtr(val),
		},
	}
	err := rg.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	e := map[string]*string{
		azure.TagClusterName: to.StringPtr(testClusterName),
		key:                  to.StringPtr(val),
	}
	if a := rg.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %+v, but got %+v", e, a)
	}
}

func TestResourceGroupCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *ResourceGroup
		success       bool
	}{
		{
			a:       nil,
			e:       &ResourceGroup{Name: to.StringPtr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &ResourceGroup{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &ResourceGroup{Name: to.StringPtr("name")},
			changes: &ResourceGroup{Name: nil},
			success: true,
		},
		{
			a:       &ResourceGroup{Name: to.StringPtr("name")},
			changes: &ResourceGroup{Name: to.StringPtr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			rg := ResourceGroup{}
			err := rg.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
