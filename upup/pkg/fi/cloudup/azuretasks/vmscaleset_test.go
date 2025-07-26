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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestSubnetIDParse(t *testing.T) {
	subnetID := &azure.SubnetID{
		SubscriptionID:     "sid",
		ResourceGroupName:  "rg",
		VirtualNetworkName: "vnet",
		SubnetName:         "sub",
	}
	actual, err := azure.ParseSubnetID(subnetID.String())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !reflect.DeepEqual(actual, subnetID) {
		t.Errorf("expected %+v, but got %+v", subnetID, actual)
	}
}

func TestLoadBalancerIDParse(t *testing.T) {
	loadBalancerID := &azure.LoadBalancerID{
		SubscriptionID:    "sid",
		ResourceGroupName: "rg",
		LoadBalancerName:  "lb",
	}
	actual, err := azure.ParseLoadBalancerID(loadBalancerID.String())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !reflect.DeepEqual(actual, loadBalancerID) {
		t.Errorf("expected %+v, but got %+v", loadBalancerID, actual)
	}
}

func newTestVMScaleSet() *VMScaleSet {
	return &VMScaleSet{
		Name:      to.Ptr("vmss"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.Ptr("rg"),
		},
		VirtualNetwork: &VirtualNetwork{
			Name: to.Ptr("vnet"),
		},
		Subnet: &Subnet{
			Name: to.Ptr("sub"),
		},
		LoadBalancer: &LoadBalancer{
			Name: to.Ptr("api-lb"),
		},
		StorageProfile:     &VMScaleSetStorageProfile{},
		RequirePublicIP:    to.Ptr(true),
		SKUName:            to.Ptr("sku"),
		Capacity:           to.Ptr[int64](10),
		ComputerNamePrefix: to.Ptr("cprefix"),
		AdminUser:          to.Ptr("admin"),
		SSHPublicKey:       to.Ptr("ssh"),
		UserData:           fi.NewStringResource("custom"),
		Tags:               map[string]*string{},
		Zones:              []*string{to.Ptr("zone1")},
	}
}

func TestVMScaleSetRenderAzure(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	apiTarget := azure.NewAzureAPITarget(cloud)
	vmss := &VMScaleSet{}
	expected := newTestVMScaleSet()
	if err := vmss.RenderAzure(apiTarget, nil, expected, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	actual := cloud.VMScaleSetsClient.VMSSes[*expected.Name]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Fatalf("unexpected location: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Name, *expected.Name; a != e {
		t.Errorf("unexpected name: expected %s, but got %s", e, a)
	}
	// Check other major fields.
	if a, e := *actual.SKU.Name, *expected.SKUName; a != e {
		t.Errorf("unexpected SKU name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.SKU.Capacity, *expected.Capacity; a != e {
		t.Errorf("unexpected SKU Capacity: expected %d, but got %d", e, a)
	}
	actualUserData, err := base64.StdEncoding.DecodeString(
		*actual.Properties.VirtualMachineProfile.UserData)
	if err != nil {
		t.Fatalf("failed to decode user data: %s", err)
	}
	expectedUserData, err := fi.ResourceAsBytes(expected.UserData)
	if err != nil {
		t.Fatalf("failed to get user data: %s", err)
	}
	if !bytes.Equal(actualUserData, expectedUserData) {
		t.Errorf("unexpected user data: expected %v, but got %v", expectedUserData, actualUserData)
	}

	if expected.PrincipalID == nil {
		t.Errorf("unexpected nil principalID")
	}

	if a, e := actual.Zones, expected.Zones; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected Zone: expected %+v, but got %+v", e, a)
	}
}

func TestVMScaleSetFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
	}

	rg := &ResourceGroup{
		Name: to.Ptr("rg"),
	}
	vmss := &VMScaleSet{
		Name:          to.Ptr("vmss"),
		ResourceGroup: rg,
	}
	// Find will return nothing if there is no VM ScaleSet created.
	actual, err := vmss.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if actual != nil {
		t.Errorf("unexpected vmss found: %+v", actual)
	}

	// Create a VM ScaleSet.
	userData := []byte("custom")
	osProfile := &compute.VirtualMachineScaleSetOSProfile{
		ComputerNamePrefix: to.Ptr("prefix"),
		AdminUsername:      to.Ptr("admin"),
		LinuxConfiguration: &compute.LinuxConfiguration{
			SSH: &compute.SSHConfiguration{
				PublicKeys: []*compute.SSHPublicKey{
					{
						Path:    to.Ptr("path"),
						KeyData: to.Ptr("ssh"),
					},
				},
			},
			DisablePasswordAuthentication: to.Ptr(true),
		},
	}
	storageProfile := &compute.VirtualMachineScaleSetStorageProfile{
		ImageReference: &compute.ImageReference{
			Publisher: to.Ptr("publisher"),
			Offer:     to.Ptr("offer"),
			SKU:       to.Ptr("sku"),
			Version:   to.Ptr("version"),
		},
		OSDisk: &compute.VirtualMachineScaleSetOSDisk{
			OSType:       to.Ptr(compute.OperatingSystemTypesLinux),
			CreateOption: to.Ptr(compute.DiskCreateOptionTypesFromImage),
			DiskSizeGB:   to.Ptr[int32](2),
			ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
				StorageAccountType: to.Ptr(compute.StorageAccountTypesPremiumLRS),
			},
			Caching: to.Ptr(compute.CachingTypes(compute.HostCachingReadWrite)),
		},
	}
	subnetID := azure.SubnetID{
		SubscriptionID:     "subID",
		ResourceGroupName:  *rg.Name,
		VirtualNetworkName: "vnet",
		SubnetName:         "sub",
	}
	loadBalancerID := azure.LoadBalancerID{
		SubscriptionID:    "subID",
		ResourceGroupName: *rg.Name,
		LoadBalancerName:  "api-lb",
	}
	ipConfigProperties := &compute.VirtualMachineScaleSetIPConfigurationProperties{
		Subnet: &compute.APIEntityReference{
			ID: to.Ptr(subnetID.String()),
		},
		Primary:                 to.Ptr(true),
		PrivateIPAddressVersion: to.Ptr(compute.IPVersionIPv4),
	}
	ipConfigProperties.PublicIPAddressConfiguration = &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
		Name: to.Ptr("vmss"),
		Properties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
			PublicIPAddressVersion: to.Ptr(compute.IPVersionIPv4),
		},
	}
	ipConfigProperties.LoadBalancerBackendAddressPools = []*compute.SubResource{
		{
			ID: to.Ptr(loadBalancerID.String()),
		},
	}
	networkConfig := &compute.VirtualMachineScaleSetNetworkConfiguration{
		Name: to.Ptr("vmss"),
		Properties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
			Primary:            to.Ptr(true),
			EnableIPForwarding: to.Ptr(true),
			IPConfigurations: []*compute.VirtualMachineScaleSetIPConfiguration{
				{
					Name:       to.Ptr("vmss"),
					Properties: ipConfigProperties,
				},
			},
		},
	}

	vmssParameters := compute.VirtualMachineScaleSet{
		Location: to.Ptr(cloud.Location),
		SKU: &compute.SKU{
			Name:     to.Ptr("sku"),
			Capacity: to.Ptr[int64](2),
		},
		Properties: &compute.VirtualMachineScaleSetProperties{
			UpgradePolicy: &compute.UpgradePolicy{
				Mode: to.Ptr(compute.UpgradeModeManual),
			},
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				OSProfile:      osProfile,
				StorageProfile: storageProfile,
				UserData:       to.Ptr(base64.RawStdEncoding.EncodeToString(userData)),
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: []*compute.VirtualMachineScaleSetNetworkConfiguration{
						networkConfig,
					},
				},
			},
		},
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type: to.Ptr(compute.ResourceIdentityTypeSystemAssigned),
		},
		Zones: []*string{to.Ptr("zone1")},
	}
	if _, err := cloud.VMScaleSet().CreateOrUpdate(context.Background(), *rg.Name, *vmss.Name, vmssParameters); err != nil {
		t.Fatalf("failed to create: %s", err)
	}

	// Find again.
	actual, err = vmss.Find(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if a, e := *actual.Name, *vmss.Name; a != e {
		t.Errorf("unexpected Virtual Network name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.ResourceGroup.Name, *rg.Name; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.VirtualNetwork.Name, subnetID.VirtualNetworkName; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Subnet.ID, subnetID.String(); a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.LoadBalancer.Name, loadBalancerID.LoadBalancerName; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	// Check other major fields.
	if a, e := *actual.SKUName, *vmssParameters.SKU.Name; a != e {
		t.Errorf("unexpected SKU name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Capacity, *vmssParameters.SKU.Capacity; a != e {
		t.Errorf("unexpected SKU Capacity: expected %d, but got %d", e, a)
	}
	if !*actual.RequirePublicIP {
		t.Errorf("unexpected require public IP")
	}
	if a, e := actual.Zones, vmssParameters.Zones; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected Zone: expected %v, but got %v", e, a)
	}
}

func TestVMScaleSetRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.CloudupContext{
		T: fi.CloudupSubContext{
			Cloud: cloud,
		},
		Target: azure.NewAzureAPITarget(cloud),
	}

	vmss := newTestVMScaleSet()
	err := vmss.Normalize(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = vmss.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expectedTags := map[string]*string{
		azure.TagClusterName: fi.PtrTo(testClusterName),
	}
	if a, e := vmss.Tags, expectedTags; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected tags: expected %+v, but got %+v", e, a)
	}
}

func TestVMScaleSetCheckChanges(t *testing.T) {
	testCases := []struct {
		a, e, changes *VMScaleSet
		success       bool
	}{
		{
			a:       nil,
			e:       &VMScaleSet{Name: to.Ptr("name")},
			changes: nil,
			success: true,
		},
		{
			a:       nil,
			e:       &VMScaleSet{Name: nil},
			changes: nil,
			success: false,
		},
		{
			a:       &VMScaleSet{Name: to.Ptr("name")},
			changes: &VMScaleSet{Name: nil},
			success: true,
		},
		{
			a:       &VMScaleSet{Name: to.Ptr("name")},
			changes: &VMScaleSet{Name: to.Ptr("newName")},
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			vmss := VMScaleSet{}
			err := vmss.CheckChanges(tc.a, tc.e, tc.changes)
			if tc.success != (err == nil) {
				t.Errorf("expected success=%t, but got err=%v", tc.success, err)
			}
		})
	}
}
