/*
Copyright 2026 The Kubernetes Authors.

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
	"fmt"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type terraformAzureVMScaleSetSSHKey struct {
	Username  *string `cty:"username"`
	PublicKey *string `cty:"public_key"`
}

type terraformAzureVMScaleSetImageReference struct {
	Publisher *string `cty:"publisher"`
	Offer     *string `cty:"offer"`
	SKU       *string `cty:"sku"`
	Version   *string `cty:"version"`
}

type terraformAzureVMScaleSetOSDisk struct {
	Caching            *string `cty:"caching"`
	StorageAccountType *string `cty:"storage_account_type"`
	DiskSizeGB         *int32  `cty:"disk_size_gb"`
}

type terraformAzureVMScaleSetPublicIPAddress struct {
	Name    *string `cty:"name"`
	Version *string `cty:"version"`
}

type terraformAzureVMScaleSetIPConfiguration struct {
	Name                              *string                                  `cty:"name"`
	Primary                           *bool                                    `cty:"primary"`
	SubnetID                          *terraformWriter.Literal                 `cty:"subnet_id"`
	ApplicationSecurityGroupIDs       []*terraformWriter.Literal               `cty:"application_security_group_ids"`
	LoadBalancerBackendAddressPoolIDs []*terraformWriter.Literal               `cty:"load_balancer_backend_address_pool_ids"`
	PublicIPAddress                   *terraformAzureVMScaleSetPublicIPAddress `cty:"public_ip_address"`
}

type terraformAzureVMScaleSetNetworkInterface struct {
	Name               *string                                    `cty:"name"`
	Primary            *bool                                      `cty:"primary"`
	EnableIPForwarding *bool                                      `cty:"enable_ip_forwarding"`
	IPConfiguration    []*terraformAzureVMScaleSetIPConfiguration `cty:"ip_configuration"`
}

type terraformAzureVMScaleSetIdentity struct {
	Type *string `cty:"type"`
}

type terraformAzureVMScaleSet struct {
	Name                          *string                                     `cty:"name"`
	ResourceGroupName             *terraformWriter.Literal                    `cty:"resource_group_name"`
	Location                      *string                                     `cty:"location"`
	SKU                           *string                                     `cty:"sku"`
	Instances                     *int64                                      `cty:"instances"`
	Zones                         []string                                    `cty:"zones"`
	UpgradeMode                   *string                                     `cty:"upgrade_mode"`
	ComputerNamePrefix            *string                                     `cty:"computer_name_prefix"`
	AdminUsername                 *string                                     `cty:"admin_username"`
	DisablePasswordAuthentication *bool                                       `cty:"disable_password_authentication"`
	AdminSSHKey                   []*terraformAzureVMScaleSetSSHKey           `cty:"admin_ssh_key"`
	SourceImageReference          *terraformAzureVMScaleSetImageReference     `cty:"source_image_reference"`
	SourceImageID                 *string                                     `cty:"source_image_id"`
	OSDisk                        *terraformAzureVMScaleSetOSDisk             `cty:"os_disk"`
	NetworkInterface              []*terraformAzureVMScaleSetNetworkInterface `cty:"network_interface"`
	Identity                      *terraformAzureVMScaleSetIdentity           `cty:"identity"`
	UserData                      *terraformWriter.Literal                    `cty:"user_data"`
	Tags                          map[string]string                           `cty:"tags"`
}

func (*VMScaleSet) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VMScaleSet) error {
	if e.StorageProfile == nil || e.StorageProfile.VirtualMachineScaleSetStorageProfile == nil {
		return fmt.Errorf("storage profile is required for VMScaleSet %q", fi.ValueOf(e.Name))
	}

	storageProfile := e.StorageProfile.VirtualMachineScaleSetStorageProfile
	if storageProfile.OSDisk == nil {
		return fmt.Errorf("os disk is required for VMScaleSet %q", fi.ValueOf(e.Name))
	}

	upgradeMode := string(compute.UpgradeModeManual)
	disablePasswordAuthentication := true
	tf := &terraformAzureVMScaleSet{
		Name:                          e.Name,
		ResourceGroupName:             e.ResourceGroup.terraformName(),
		Location:                      new(t.Cloud.Region()),
		SKU:                           e.SKUName,
		Instances:                     e.Capacity,
		Zones:                         stringSlice(e.Zones),
		UpgradeMode:                   &upgradeMode,
		ComputerNamePrefix:            e.ComputerNamePrefix,
		AdminUsername:                 e.AdminUser,
		DisablePasswordAuthentication: &disablePasswordAuthentication,
		AdminSSHKey: []*terraformAzureVMScaleSetSSHKey{
			{
				Username:  e.AdminUser,
				PublicKey: e.SSHPublicKey,
			},
		},
		OSDisk: &terraformAzureVMScaleSetOSDisk{
			Caching:            stringPtr(storageProfile.OSDisk.Caching),
			StorageAccountType: storageAccountType(storageProfile.OSDisk.ManagedDisk),
			DiskSizeGB:         storageProfile.OSDisk.DiskSizeGB,
		},
		Identity: &terraformAzureVMScaleSetIdentity{
			Type: new("SystemAssigned"),
		},
		Tags: stringMap(e.Tags),
	}

	if e.UserData != nil {
		userData, err := t.AddFileResource("azurerm_linux_virtual_machine_scale_set", fi.ValueOf(e.Name), "user_data", e.UserData, true)
		if err != nil {
			return err
		}
		tf.UserData = userData
	}

	ipConfig, err := e.terraformIPConfiguration(t)
	if err != nil {
		return err
	}
	tf.NetworkInterface = []*terraformAzureVMScaleSetNetworkInterface{
		{
			Name:               e.Name,
			Primary:            new(true),
			EnableIPForwarding: new(true),
			IPConfiguration: []*terraformAzureVMScaleSetIPConfiguration{
				ipConfig,
			},
		},
	}

	if image := storageProfile.ImageReference; image != nil {
		if image.ID != nil {
			tf.SourceImageID = image.ID
		} else {
			tf.SourceImageReference = &terraformAzureVMScaleSetImageReference{
				Publisher: image.Publisher,
				Offer:     image.Offer,
				SKU:       image.SKU,
				Version:   image.Version,
			}
		}
	}

	return t.RenderResource("azurerm_linux_virtual_machine_scale_set", fi.ValueOf(e.Name), tf)
}

func (vmss *VMScaleSet) terraformIPConfiguration(t *terraform.TerraformTarget) (*terraformAzureVMScaleSetIPConfiguration, error) {
	subnetID, err := vmss.Subnet.terraformID(t)
	if err != nil {
		return nil, err
	}
	cfg := &terraformAzureVMScaleSetIPConfiguration{
		Name:                        vmss.Name,
		Primary:                     new(true),
		SubnetID:                    subnetID,
		ApplicationSecurityGroupIDs: applicationSecurityGroupIDs(vmss.ApplicationSecurityGroups),
	}
	if fi.ValueOf(vmss.RequirePublicIP) {
		version := string(network.IPVersionIPv4)
		cfg.PublicIPAddress = &terraformAzureVMScaleSetPublicIPAddress{
			Name:    vmss.Name,
			Version: &version,
		}
	}
	if vmss.LoadBalancer != nil {
		cfg.LoadBalancerBackendAddressPoolIDs = []*terraformWriter.Literal{vmss.LoadBalancer.terraformBackendAddressPoolID()}
	}
	return cfg, nil
}

func (vmss *VMScaleSet) terraformPrincipalID() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("azurerm_linux_virtual_machine_scale_set", fi.ValueOf(vmss.Name), "identity[0].principal_id")
}

func storageAccountType(disk *compute.VirtualMachineScaleSetManagedDiskParameters) *string {
	if disk == nil || disk.StorageAccountType == nil {
		return nil
	}
	return stringPtr(disk.StorageAccountType)
}
