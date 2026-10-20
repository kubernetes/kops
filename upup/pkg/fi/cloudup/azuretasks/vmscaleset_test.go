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
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
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
		Name:      to.Ptr("vmss." + testClusterName),
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
		Tags: map[string]*string{
			azure.TagClusterName:       new(testClusterName),
			azure.InstanceGroupNameTag: new("vmss"),
		},
		Zones: []*string{to.Ptr("zone1")},
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

	actual := cloud.VMScaleSetsClient.VMSSes["vmss"]
	if a, e := *actual.Location, cloud.Region(); a != e {
		t.Fatalf("unexpected location: expected %s, but got %s", e, a)
	}
	if a, e := *actual.Name, "vmss"; a != e {
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
		Name:          to.Ptr("vmss." + testClusterName),
		ResourceGroup: rg,
		Tags: map[string]*string{
			azure.TagClusterName:       new(testClusterName),
			azure.InstanceGroupNameTag: new("vmss"),
		},
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
		Tags:  vmss.Tags,
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
	delete(vmss.Tags, azure.TagClusterName)
	err := vmss.Normalize(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = vmss.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expectedTags := map[string]*string{
		azure.TagClusterName:       new(testClusterName),
		azure.InstanceGroupNameTag: new("vmss"),
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

func TestVMScaleSetNaming(t *testing.T) {
	const clusterName = "pr18879-kops-azure-upgrade-dns-none"
	for _, tc := range []struct {
		name         string
		igName       string
		legacyName   string
		existingName string
		shared       bool
		failed       bool
	}{
		{name: "new nodes", igName: "nodes", legacyName: "nodes." + clusterName},
		{name: "new control plane", igName: "control-plane-eastus-1", legacyName: "control-plane-eastus-1.masters." + clusterName},
		{name: "existing nodes", igName: "nodes", legacyName: "nodes." + clusterName, existingName: "nodes." + clusterName},
		{name: "failed legacy nodes", igName: "nodes", legacyName: "nodes." + clusterName, existingName: "nodes." + clusterName, failed: true},
		{name: "existing control plane", igName: "control-plane-eastus-1", legacyName: "control-plane-eastus-1.masters." + clusterName, existingName: "control-plane-eastus-1.masters." + clusterName},
		{name: "existing 69-character control plane", igName: "control-plane-uksouth-2-1", legacyName: "control-plane-uksouth-2-1.masters." + clusterName, existingName: "control-plane-uksouth-2-1.masters." + clusterName},
		{name: "existing short name", igName: "nodes", legacyName: "nodes." + clusterName, existingName: "nodes"},
		{name: "dotted instance group", igName: "nodes.special", legacyName: "nodes.special." + clusterName},
		{name: "new shared resource group", igName: "nodes", legacyName: "nodes." + clusterName, shared: true},
		{name: "existing shared resource group", igName: "nodes", legacyName: "nodes." + clusterName, existingName: "nodes." + clusterName, shared: true},
		{name: "shared resource group with short name", igName: "nodes", legacyName: "nodes." + clusterName, existingName: "nodes", shared: true},
		{name: "shared resource group control plane", igName: "control-plane-eastus-1", legacyName: "control-plane-eastus-1.masters." + clusterName, shared: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cloud := NewMockAzureCloud("eastus")
			apiTarget := azure.NewAzureAPITarget(cloud)
			newTask := func() *VMScaleSet {
				s := newTestVMScaleSet()
				s.Name = new(tc.legacyName)
				s.ResourceGroup.Shared = new(tc.shared)
				s.Subnet.ID = new((&azure.SubnetID{
					SubscriptionID:     cloud.SubscriptionID(),
					ResourceGroupName:  *s.ResourceGroup.Name,
					VirtualNetworkName: *s.VirtualNetwork.Name,
					SubnetName:         *s.Subnet.Name,
				}).String())
				s.Tags = map[string]*string{
					azure.TagClusterName:       new(clusterName),
					azure.InstanceGroupNameTag: new(tc.igName),
				}
				s.StorageProfile.VirtualMachineScaleSetStorageProfile = &compute.VirtualMachineScaleSetStorageProfile{
					OSDisk: &compute.VirtualMachineScaleSetOSDisk{},
				}
				return s
			}
			wantName := tc.igName
			if tc.shared {
				wantName = tc.legacyName
			}
			if tc.existingName != "" {
				wantName = tc.existingName
				existing := newTask()
				// Seed the exact physical name using the shared-group naming convention.
				existing.Name = new(tc.existingName)
				existing.ResourceGroup.Shared = new(true)
				if err := existing.RenderAzure(apiTarget, nil, existing, nil); err != nil {
					t.Fatal(err)
				}
				if tc.failed {
					cloud.VMScaleSetsClient.VMSSes[tc.existingName].Properties.ProvisioningState = new("Failed")
				}
			}

			// Terraform names depend only on whether the resource group is shared,
			// even when a legacy VMSS exists. Resource addresses stay unchanged.
			tfTask := newTask()
			// A nil client ensures Terraform cannot query Azure for the VMSS name.
			tfCloud := &vmScaleSetLookupCloud{AzureCloud: cloud}
			tfTarget := terraform.NewTerraformTarget(tfCloud, "", t.TempDir(), nil)
			if err := tfTask.RenderTerraform(tfTarget, nil, tfTask, nil); err != nil {
				t.Fatal(err)
			}
			resources, err := tfTarget.GetResourcesByType()
			if err != nil {
				t.Fatal(err)
			}
			tf := resources["azurerm_linux_virtual_machine_scale_set"][strings.ReplaceAll(tc.legacyName, ".", "-")].(*terraformAzureVMScaleSet)
			ni := tf.NetworkInterface[0]
			ip := ni.IPConfiguration[0]
			wantTerraformName := tc.igName
			if tc.shared {
				wantTerraformName = tc.legacyName
			}
			for _, name := range []*string{tf.Name, ni.Name, ip.Name, ip.PublicIPAddress.Name} {
				if fi.ValueOf(name) != wantTerraformName {
					t.Errorf("Terraform name = %q, want %q", fi.ValueOf(name), wantTerraformName)
				}
			}

			// The API target must create or update the same VMSS, without adding a duplicate.
			desired := newTask()
			desired.Capacity = new(int64(11))
			ctx := &fi.CloudupContext{T: fi.CloudupSubContext{Cloud: cloud}, Target: apiTarget}
			if err := desired.Run(ctx); err != nil {
				t.Fatal(err)
			}
			sets := cloud.VMScaleSetsClient.VMSSes
			if len(sets) != 1 || sets[wantName] == nil {
				t.Fatalf("expected only VMSS %q, got %v", wantName, sets)
			}
			actual := sets[wantName]
			if fi.ValueOf(actual.SKU.Capacity) != 11 {
				t.Errorf("capacity = %d, want 11", fi.ValueOf(actual.SKU.Capacity))
			}
			niName := actual.Properties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations[0].Name
			if fi.ValueOf(niName) != wantName {
				t.Errorf("NIC name = %q, want %q", fi.ValueOf(niName), wantName)
			}

			// A repeated apply must leave the existing resource untouched.
			repeated := newTask()
			repeated.Capacity = new(int64(11))
			if err := repeated.Run(ctx); err != nil {
				t.Fatal(err)
			}
			if sets[wantName] != actual {
				t.Error("repeated apply updated the VM Scale Set")
			}
		})
	}
}

func TestVMScaleSetFindAndRenderNaming(t *testing.T) {
	clusterName := "cluster.example.com"
	legacyName := "nodes." + clusterName
	for _, tc := range []struct {
		name              string
		sets              []*compute.VirtualMachineScaleSet
		legacyName        string
		shared            bool
		getErrors         map[string]error
		want              string
		wantFound         bool
		wantErr           string
		wantGets          []string
		deleteBeforeRetry bool
	}{
		{name: "new scale set", want: "nodes"},
		{name: "long cluster name", legacyName: "nodes." + strings.Repeat("long-cluster-", 8) + "example.com", want: "nodes"},
		{name: "shared resource group with another cluster", shared: true, sets: []*compute.VirtualMachineScaleSet{
			{Name: new("nodes.other.example.com"), Tags: map[string]*string{azure.TagClusterName: new("other.example.com"), azure.InstanceGroupNameTag: new("nodes")}},
		}, want: legacyName},
		{name: "shared resource group with foreign short name", shared: true, sets: []*compute.VirtualMachineScaleSet{
			{Name: new("nodes"), Tags: map[string]*string{azure.TagClusterName: new("other.example.com"), azure.InstanceGroupNameTag: new("nodes")}},
		}, want: legacyName},
		{name: "legacy without IG tag", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName)}},
		}, wantFound: true, want: legacyName},
		{name: "shared resource group legacy without IG tag", shared: true, sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName)}},
		}, wantFound: true, want: legacyName},
		{name: "legacy removed before retry", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("nodes")}},
		}, deleteBeforeRetry: true, want: "nodes", wantGets: []string{"nodes", legacyName, "nodes", legacyName}},
		{name: "unrelated scale set", sets: []*compute.VirtualMachineScaleSet{
			{Name: new("other"), Tags: map[string]*string{azure.TagClusterName: new("other-cluster")}},
		}, want: "nodes"},
		{name: "short name with different IG tag", sets: []*compute.VirtualMachineScaleSet{
			{Name: new("nodes"), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("other")}},
		}, wantFound: true, want: "nodes", wantGets: []string{"nodes"}},
		{name: "short name without IG tag", sets: []*compute.VirtualMachineScaleSet{
			{Name: new("nodes"), Tags: map[string]*string{azure.TagClusterName: new(clusterName)}},
		}, wantFound: true, want: "nodes", wantGets: []string{"nodes"}},
		{name: "legacy without cluster tag", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.InstanceGroupNameTag: new("nodes")}},
		}, want: "nodes"},
		{name: "legacy with different IG tag", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("other")}},
		}, wantFound: true, want: legacyName},
		{name: "legacy name with different casing", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(strings.ToUpper(legacyName)), Tags: map[string]*string{azure.TagClusterName: new(clusterName)}},
		}, wantFound: true, want: strings.ToUpper(legacyName)},
		{name: "legacy name with unrelated short name", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("nodes")}},
			{Name: new("nodes"), Tags: map[string]*string{azure.TagClusterName: new("other-cluster")}},
		}, wantFound: true, want: legacyName},
		{name: "short name takes precedence", sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("nodes")}},
			{Name: new("nodes"), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("nodes")}},
		}, wantFound: true, want: "nodes", wantGets: []string{"nodes"}},
		{name: "missing scale set", getErrors: map[string]error{
			"nodes": fmt.Errorf("getting VMSS: %w", &azcore.ResponseError{StatusCode: http.StatusNotFound, ErrorCode: "ResourceNotFound"}),
		}, sets: []*compute.VirtualMachineScaleSet{
			{Name: new(legacyName), Tags: map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("nodes")}},
		}, wantFound: true, want: legacyName},
		{name: "missing resource group", getErrors: map[string]error{
			"nodes":    &azcore.ResponseError{StatusCode: http.StatusNotFound, ErrorCode: "ResourceGroupNotFound"},
			legacyName: &azcore.ResponseError{StatusCode: http.StatusNotFound, ErrorCode: "ResourceGroupNotFound"},
		}, want: "nodes"},
		{name: "authorization failure", getErrors: map[string]error{
			"nodes": &azcore.ResponseError{StatusCode: http.StatusForbidden, ErrorCode: "AuthorizationFailed"},
		}, wantErr: "AuthorizationFailed", wantGets: []string{"nodes"}},
		{name: "legacy lookup failure", getErrors: map[string]error{
			legacyName: &azcore.ResponseError{StatusCode: http.StatusInternalServerError, ErrorCode: "InternalServerError"},
		}, wantErr: "InternalServerError"},
		{name: "bad request", getErrors: map[string]error{
			"nodes": &azcore.ResponseError{StatusCode: http.StatusBadRequest, ErrorCode: "InvalidParameter"},
		}, wantErr: "InvalidParameter", wantGets: []string{"nodes"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cloud := NewMockAzureCloud("eastus")
			for _, set := range tc.sets {
				existing := newTestVMScaleSet()
				existing.Name = set.Name
				existing.Tags = set.Tags
				existing.ResourceGroup.Shared = new(true)
				if err := existing.RenderAzure(azure.NewAzureAPITarget(cloud), nil, existing, nil); err != nil {
					t.Fatal(err)
				}
			}
			task := newTestVMScaleSet()
			task.Name = new(legacyName)
			if tc.legacyName != "" {
				task.Name = new(tc.legacyName)
			}
			task.ResourceGroup.Shared = new(tc.shared)
			task.Tags = map[string]*string{azure.TagClusterName: new(clusterName), azure.InstanceGroupNameTag: new("nodes")}
			client := &vmScaleSetLookupClient{MockVMScaleSetsClient: cloud.VMScaleSetsClient, getErrors: tc.getErrors}
			lookupCloud := &vmScaleSetLookupCloud{AzureCloud: cloud, client: client}
			ctx := &fi.CloudupContext{T: fi.CloudupSubContext{Cloud: lookupCloud}}
			actual, err := task.Find(ctx)
			if tc.deleteBeforeRetry {
				if err != nil || actual == nil {
					t.Fatalf("expected existing VMSS before retry, got %v, %v", actual, err)
				}
				delete(cloud.VMScaleSetsClient.VMSSes, legacyName)
				actual, err = task.Find(ctx)
				if actual != nil {
					t.Fatalf("expected no VMSS after deletion, got %v", actual)
				}
			}
			wantGets := tc.wantGets
			if wantGets == nil {
				wantGets = []string{"nodes", *task.Name}
			}
			if !reflect.DeepEqual(client.gets, wantGets) {
				t.Errorf("VMSS lookups = %v, want %v", client.gets, wantGets)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if (actual != nil) != tc.wantFound {
				t.Fatalf("Find returned %v, want found=%t", actual, tc.wantFound)
			}
			// Rendering must use the discovery result or the configured creation name without another GET.
			client.denyGets = true
			task.Capacity = new(int64(11))
			if err := task.RenderAzure(azure.NewAzureAPITarget(lookupCloud), actual, task, nil); err != nil {
				t.Fatal(err)
			}
			set := cloud.VMScaleSetsClient.VMSSes[tc.want]
			if set == nil || fi.ValueOf(set.SKU.Capacity) != 11 {
				t.Errorf("VMSS %q was not created or updated to the desired capacity", tc.want)
			}
			if !reflect.DeepEqual(client.gets, wantGets) {
				t.Errorf("rendering added VMSS lookups: got %v, want %v", client.gets, wantGets)
			}
		})
	}
}

type vmScaleSetLookupCloud struct {
	azure.AzureCloud
	client azure.VMScaleSetsClient
}

func (c *vmScaleSetLookupCloud) VMScaleSet() azure.VMScaleSetsClient {
	return c.client
}

type vmScaleSetLookupClient struct {
	*MockVMScaleSetsClient
	getErrors map[string]error
	gets      []string
	denyGets  bool
}

func (c *vmScaleSetLookupClient) Get(ctx context.Context, resourceGroupName, name string) (*compute.VirtualMachineScaleSet, error) {
	c.gets = append(c.gets, name)
	if c.denyGets {
		return nil, fmt.Errorf("unexpected VMSS GET during rendering: %s", name)
	}
	if err := c.getErrors[name]; err != nil {
		return nil, err
	}
	return c.MockVMScaleSetsClient.Get(ctx, resourceGroupName, name)
}
