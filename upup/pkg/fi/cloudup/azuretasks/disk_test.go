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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

const (
	testTagKey   = "key"
	testTagValue = "value"
)

func newTestDisk() *Disk {
	return &Disk{
		Name: to.StringPtr("disk"),
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		SizeGB: to.Int32Ptr(32),
		Tags: map[string]*string{
			testTagKey: to.StringPtr(testTagValue),
		},
	}
}

func TestDiskRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	disk := &Disk{}
	expected := newTestDisk()
	if err := disk.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.DisksClient.Disks[*expected.Name]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Fatalf("unexpected location: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.DiskProperties.DiskSizeGB, *expected.SizeGB; a != e {
		t.Fatalf("unexpected disk size: expected %d, but got %d", e, a)
	}
	if a, e := actual.Tags, expected.Tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %v, but got %v", e, a)
	}
}

func TestDiskFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud: cloud,
	}

	rg := &ResourceGroup{
		Name: to.StringPtr("rg"),
	}
	disk := &Disk{
		Name:          to.StringPtr("disk"),
		ResourceGroup: rg,
	}
	// Find will return nothing if there is no Disk created.
	actual, err := disk.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected Disk found: %+v", actual)
	}

	// Create a Disk.
	var diskSizeGB int32 = 32
	tags := map[string]*string{
		"key": to.StringPtr("value"),
	}
	diskParameters := compute.Disk{
		Location: to.StringPtr(cloud.Location),
		DiskProperties: &compute.DiskProperties{
			CreationData: &compute.CreationData{
				CreateOption: compute.Empty,
			},
			DiskSizeGB: to.Int32Ptr(diskSizeGB),
		},
		Tags: tags,
	}
	if err := cloud.Disk().CreateOrUpdate(context.Background(), *rg.Name, *disk.Name, diskParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}
	// Find again.
	actual, err = disk.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *disk.Name; a != e {
		t.Errorf("unexpected Disk name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.ResourceGroup.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.SizeGB, diskSizeGB; a != e {
		t.Errorf("unexpected disk size: expected %d, but got %d", e, a)
	}
	if a, e := actual.Tags, tags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %v, but got %v", e, a)
	}
}

func TestDiskRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud:  cloud,
		Target: azure.NewAzureAPITarget(cloud),
	}

	vmss := newTestDisk()
	err := vmss.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expectedTags := map[string]*string{
		azure.TagClusterName: to.StringPtr(testClusterName),
		testTagKey:           to.StringPtr(testTagValue),
	}
	if a, e := vmss.Tags, expectedTags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %+v, but got %+v", e, a)
	}
}

func TestDiskCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *Disk
		success       bool
	}{
		{
			a:       nil,
			e:       &Disk{Name: to.StringPtr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &Disk{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &Disk{Name: to.StringPtr("name")},
			changes: &Disk{Name: nil},
			success: true,
		},
		{
			a:       &Disk{Name: to.StringPtr("name")},
			changes: &Disk{Name: to.StringPtr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			d := Disk{}
			err := d.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
