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
	resources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestResourceGroupRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	rg := &ResourceGroup{}
	expected := &ResourceGroup{
		Name: to.Ptr("rg"),
		Tags: map[string]*string{
			"key": to.Ptr("value"),
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
		Name: to.Ptr("rg"),
		Tags: map[string]*string{
			"key2": to.Ptr("value2"),
		},
	}
	changes := &ResourceGroup{
		Tags: map[string]*string{
			"key2": to.Ptr("value2"),
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
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	rg := &ResourceGroup{
		Name:   to.Ptr("rg"),
		Shared: to.Ptr(true),
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
	rgParameters := resources.ResourceGroup{
		Location: to.Ptr(cloud.Location),
		Tags: map[string]*string{
			"key": to.Ptr("value"),
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
		Name: to.Ptr("invalid"),
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
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
		Target: azure.NewAzureAPITarget(cloud),
	}

	const (
		key = "key"
		val = "val"
	)
	rg := &ResourceGroup{
		Name:      to.Ptr("rg"),
		Lifecycle: fi.LifecycleSync,
		Tags: map[string]*string{
			key: to.Ptr(val),
		},
	}
	err := rg.Normalize(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = rg.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	e := map[string]*string{
		azure.TagClusterName: to.Ptr(testClusterName),
		key:                  to.Ptr(val),
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
			e:       &ResourceGroup{Name: to.Ptr("name")},
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
			a:       &ResourceGroup{Name: to.Ptr("name")},
			changes: &ResourceGroup{Name: nil},
			success: true,
		},
		{
			a:       &ResourceGroup{Name: to.Ptr("name")},
			changes: &ResourceGroup{Name: to.Ptr("newName")},
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
