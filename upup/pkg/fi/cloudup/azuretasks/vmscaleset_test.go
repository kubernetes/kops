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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

func TestSubnetIDParse(t *testing.T) {
	subnetID := &SubnetID{
		SubscriptionID:     "sid",
		ResourceGroupName:  "rg",
		VirtualNetworkName: "vnet",
		SubnetName:         "sub",
	}
	actual, err := ParseSubnetID(subnetID.String())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !reflect.DeepEqual(actual, subnetID) {
		t.Errorf("expected %+v, but got %+v", subnetID, actual)
	}
}

func TestLoadBalancerIDParse(t *testing.T) {
	loadBalancerID := &loadBalancerID{
		SubscriptionID:    "sid",
		ResourceGroupName: "rg",
		LoadBalancerName:  "lb",
	}
	actual, err := parseLoadBalancerID(loadBalancerID.String())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !reflect.DeepEqual(actual, loadBalancerID) {
		t.Errorf("expected %+v, but got %+v", loadBalancerID, actual)
	}
}

func newTestVMScaleSet() *VMScaleSet {
	return &VMScaleSet{
		Name:      to.StringPtr("vmss"),
		Lifecycle: fi.LifecycleSync,
		ResourceGroup: &ResourceGroup{
			Name: to.StringPtr("rg"),
		},
		VirtualNetwork: &VirtualNetwork{
			Name: to.StringPtr("vnet"),
		},
		Subnet: &Subnet{
			Name: to.StringPtr("sub"),
		},
		LoadBalancer: &LoadBalancer{
			Name: to.StringPtr("api-lb"),
		},
		StorageProfile:     &VMScaleSetStorageProfile{},
		RequirePublicIP:    to.BoolPtr(true),
		SKUName:            to.StringPtr("sku"),
		Capacity:           to.Int64Ptr(10),
		ComputerNamePrefix: to.StringPtr("cprefix"),
		AdminUser:          to.StringPtr("admin"),
		SSHPublicKey:       to.StringPtr("ssh"),
		CustomData:         fi.NewStringResource("custom"),
		Tags:               map[string]*string{},
		Zones:              []string{"zone1"},
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
	if a, e := *actual.Sku.Name, *expected.SKUName; a != e {
		t.Errorf("unexpected SKU name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Sku.Capacity, *expected.Capacity; a != e {
		t.Errorf("unexpected SKU Capacity: expected %d, but got %d", e, a)
	}
	actualCData, err := base64.StdEncoding.DecodeString(
		*actual.VirtualMachineProfile.OsProfile.CustomData)
	if err != nil {
		t.Fatalf("failed to decode custom data: %s", err)
	}
	expectedCData, err := fi.ResourceAsBytes(expected.CustomData)
	if err != nil {
		t.Fatalf("failed to get custom data: %s", err)
	}
	if !bytes.Equal(actualCData, expectedCData) {
		t.Errorf("unexpected custom data: expected %v, but got %v", expectedCData, actualCData)
	}

	if expected.PrincipalID == nil {
		t.Errorf("unexpected nil principalID")
	}

	if a, e := *actual.Zones, expected.Zones; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected Zone: expected %s, but got %s", e, a)
	}
}

func TestVMScaleSetFind(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud: cloud,
	}

	rg := &ResourceGroup{
		Name: to.StringPtr("rg"),
	}
	vmss := &VMScaleSet{
		Name:          to.StringPtr("vmss"),
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
	customData := []byte("custom")
	osProfile := &compute.VirtualMachineScaleSetOSProfile{
		ComputerNamePrefix: to.StringPtr("prefix"),
		AdminUsername:      to.StringPtr("admin"),
		CustomData:         to.StringPtr(base64.RawStdEncoding.EncodeToString(customData)),
		LinuxConfiguration: &compute.LinuxConfiguration{
			SSH: &compute.SSHConfiguration{
				PublicKeys: &[]compute.SSHPublicKey{
					{
						Path:    to.StringPtr("path"),
						KeyData: to.StringPtr("ssh"),
					},
				},
			},
			DisablePasswordAuthentication: to.BoolPtr(true),
		},
	}
	storageProfile := &compute.VirtualMachineScaleSetStorageProfile{
		ImageReference: &compute.ImageReference{
			Publisher: to.StringPtr("publisher"),
			Offer:     to.StringPtr("offer"),
			Sku:       to.StringPtr("sku"),
			Version:   to.StringPtr("version"),
		},
		OsDisk: &compute.VirtualMachineScaleSetOSDisk{
			OsType:       compute.OperatingSystemTypes(compute.Linux),
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			DiskSizeGB:   to.Int32Ptr(2),
			ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
				StorageAccountType: compute.StorageAccountTypesPremiumLRS,
			},
			Caching: compute.CachingTypes(compute.HostCachingReadWrite),
		},
	}
	subnetID := SubnetID{
		SubscriptionID:     "subID",
		ResourceGroupName:  *rg.Name,
		VirtualNetworkName: "vnet",
		SubnetName:         "sub",
	}
	loadBalancerID := loadBalancerID{
		SubscriptionID:    "subID",
		ResourceGroupName: *rg.Name,
		LoadBalancerName:  "api-lb",
	}
	ipConfigProperties := &compute.VirtualMachineScaleSetIPConfigurationProperties{
		Subnet: &compute.APIEntityReference{
			ID: to.StringPtr(subnetID.String()),
		},
		Primary:                 to.BoolPtr(true),
		PrivateIPAddressVersion: compute.IPv4,
	}
	ipConfigProperties.PublicIPAddressConfiguration = &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
		Name: to.StringPtr("vmss-publicipconfig"),
		VirtualMachineScaleSetPublicIPAddressConfigurationProperties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
			PublicIPAddressVersion: compute.IPv4,
		},
	}
	ipConfigProperties.LoadBalancerBackendAddressPools = &[]compute.SubResource{
		{
			ID: to.StringPtr(loadBalancerID.String()),
		},
	}
	networkConfig := compute.VirtualMachineScaleSetNetworkConfiguration{
		Name: to.StringPtr("vmss-netconfig"),
		VirtualMachineScaleSetNetworkConfigurationProperties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
			Primary:            to.BoolPtr(true),
			EnableIPForwarding: to.BoolPtr(true),
			IPConfigurations: &[]compute.VirtualMachineScaleSetIPConfiguration{
				{
					Name: to.StringPtr("vmss-ipconfig"),
					VirtualMachineScaleSetIPConfigurationProperties: ipConfigProperties,
				},
			},
		},
	}

	vmssParameters := compute.VirtualMachineScaleSet{
		Location: to.StringPtr(cloud.Location),
		Sku: &compute.Sku{
			Name:     to.StringPtr("sku"),
			Capacity: to.Int64Ptr(2),
		},
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			UpgradePolicy: &compute.UpgradePolicy{
				Mode: compute.UpgradeModeManual,
			},
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				OsProfile:      osProfile,
				StorageProfile: storageProfile,
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: &[]compute.VirtualMachineScaleSetNetworkConfiguration{
						networkConfig,
					},
				},
			},
		},
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type: compute.ResourceIdentityTypeSystemAssigned,
		},
		Zones: &[]string{"zone1"},
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
	if a, e := *actual.Subnet.Name, subnetID.SubnetName; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.LoadBalancer.Name, loadBalancerID.LoadBalancerName; a != e {
		t.Errorf("unexpected Resource Group name: expected %s, but got %s", e, a)
	}
	// Check other major fields.
	if a, e := *actual.SKUName, *vmssParameters.Sku.Name; a != e {
		t.Errorf("unexpected SKU name: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Capacity, *vmssParameters.Sku.Capacity; a != e {
		t.Errorf("unexpected SKU Capacity: expected %d, but got %d", e, a)
	}
	if !*actual.RequirePublicIP {
		t.Errorf("unexpected require public IP")
	}
	if a, e := actual.Zones, *vmssParameters.Zones; !reflect.DeepEqual(a, e) {
		t.Errorf("unexpected Zone: expected %s, but got %s", e, a)
	}
}

func TestVMScaleSetRun(t *testing.T) {
	cloud := NewMockAzureCloud("eastus")
	ctx := &fi.Context{
		Cloud:  cloud,
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
		azure.TagClusterName: fi.String(testClusterName),
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
			e:       &VMScaleSet{Name: to.StringPtr("name")},
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
			a:       &VMScaleSet{Name: to.StringPtr("name")},
			changes: &VMScaleSet{Name: nil},
			success: true,
		},
		{
			a:       &VMScaleSet{Name: to.StringPtr("name")},
			changes: &VMScaleSet{Name: to.StringPtr("newName")},
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
