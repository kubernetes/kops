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
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// VMScaleSet is an Azure VM Scale Set.
// +kops:fitask
type VMScaleSet struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ResourceGroup             *ResourceGroup
	VirtualNetwork            *VirtualNetwork
	Subnet                    *Subnet
	ApplicationSecurityGroups []*ApplicationSecurityGroup
	StorageProfile            *VMScaleSetStorageProfile
	// RequirePublicIP is set to true when VMs require public IPs.
	RequirePublicIP *bool
	// LoadBalancer is the Load Balancer object the VMs will use.
	LoadBalancer *LoadBalancer
	// SKUName specifies the SKU of the VM Scale Set
	SKUName *string
	// Capacity specifies the number of virtual machines the VM Scale Set.
	Capacity *int64
	// ComputerNamePrefix is the prefix of each VM name of the form <prefix><base-36-instance-id>.
	// See https://docs.microsoft.com/en-us/azure/virtual-machine-scale-sets/virtual-machine-scale-sets-instance-ids.
	ComputerNamePrefix *string
	// AdmnUser specifies the name of the administrative account.
	AdminUser    *string
	SSHPublicKey *string
	// UserData is the user data configuration
	UserData    fi.Resource
	Tags        map[string]*string
	Zones       []*string
	PrincipalID *string
}

var _ fi.CloudupTaskNormalize = &VMScaleSet{}

// VMScaleSetStorageProfile wraps *compute.VirtualMachineScaleSetStorageProfile
// and implements fi.HasDependencies.
//
// If we don't implement the interface and directly use
// compute.VirtualMachineScaleSetStorageProfile in VMScaleSet, the
// topological sort on VMScaleSet will fail as StorageProfile doesn't
// implement a proper interface.
type VMScaleSetStorageProfile struct {
	*compute.VirtualMachineScaleSetStorageProfile
}

var _ fi.CloudupHasDependencies = &VMScaleSetStorageProfile{}

// GetDependencies returns a slice of tasks on which the tasks depends on.
func (p *VMScaleSetStorageProfile) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

var (
	_ fi.CloudupTask   = &VMScaleSet{}
	_ fi.CompareWithID = &VMScaleSet{}
)

// CompareWithID returns the Name of the VM Scale Set.
func (s *VMScaleSet) CompareWithID() *string {
	return s.Name
}

// Find discovers the VMScaleSet in the cloud provider.
func (s *VMScaleSet) Find(c *fi.CloudupContext) (*VMScaleSet, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	found, err := cloud.VMScaleSet().Get(context.TODO(), *s.ResourceGroup.Name, *s.Name)
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return nil, err
	}
	if found == nil {
		return nil, nil
	}

	if found.ID == nil {
		return nil, fmt.Errorf("found VMSS without ID")
	}
	if found.Properties == nil {
		return nil, fmt.Errorf("found VMSS without properties")
	}
	if found.Properties.VirtualMachineProfile == nil {
		return nil, fmt.Errorf("found VMSS without VM profile")
	}
	if found.Properties.VirtualMachineProfile.NetworkProfile == nil {
		return nil, fmt.Errorf("found VMSS without network profile")
	}
	if found.Properties.VirtualMachineProfile.OSProfile == nil {
		return nil, fmt.Errorf("found VMSS without OS profile")
	}

	profile := found.Properties.VirtualMachineProfile

	nwConfigs := profile.NetworkProfile.NetworkInterfaceConfigurations
	if len(nwConfigs) != 1 {
		return nil, fmt.Errorf("expecting exactly 1 network interface config for %q, found %d: %+v", *s.Name, len(nwConfigs), nwConfigs)
	}
	nwConfig := nwConfigs[0]
	if nwConfig.Properties == nil {
		return nil, fmt.Errorf("found VMSS without network interface config properties")
	}
	ipConfigs := nwConfig.Properties.IPConfigurations
	if len(ipConfigs) != 1 {
		return nil, fmt.Errorf("expecting exactly 1 network interface IP config for %q, found %d: %+v", *s.Name, len(ipConfigs), ipConfigs)
	}
	ipConfig := ipConfigs[0]
	if ipConfig.Properties == nil {
		return nil, fmt.Errorf("found VMSS without IP config properties")
	}
	if ipConfig.Properties.Subnet == nil {
		return nil, fmt.Errorf("found VMSS without IP config subnet")
	}
	if ipConfig.Properties.Subnet.ID == nil {
		return nil, fmt.Errorf("found VMSS without IP config subnet ID")
	}
	subnetID, err := azure.ParseSubnetID(*ipConfig.Properties.Subnet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subnet ID %s", *ipConfig.Properties.Subnet.ID)
	}

	var loadBalancerID *azure.LoadBalancerID
	if ipConfig.Properties.LoadBalancerBackendAddressPools != nil {
		for _, i := range ipConfig.Properties.LoadBalancerBackendAddressPools {
			if !strings.Contains(*i.ID, "api") {
				continue
			}
			loadBalancerID, err = azure.ParseLoadBalancerID(*i.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse loadbalancer ID %s", *i.ID)
			}
		}
	}

	osProfile := profile.OSProfile
	if osProfile.LinuxConfiguration == nil {
		return nil, fmt.Errorf("found VMSS without Linux config")
	}
	if osProfile.LinuxConfiguration.SSH == nil {
		return nil, fmt.Errorf("found VMSS without SSH config")
	}
	if osProfile.LinuxConfiguration.SSH.PublicKeys == nil {
		return nil, fmt.Errorf("found VMSS without SSH public keys")
	}
	sshKeys := osProfile.LinuxConfiguration.SSH.PublicKeys
	if len(sshKeys) != 1 {
		return nil, fmt.Errorf("expecting exactly 1 SSH key for %q, found %d: %+v", *s.Name, len(sshKeys), sshKeys)
	}

	userData, err := base64.StdEncoding.DecodeString(*profile.UserData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user data: %w", err)
	}

	vmss := &VMScaleSet{
		Name:      s.Name,
		Lifecycle: s.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: s.ResourceGroup.Name,
		},
		VirtualNetwork: &VirtualNetwork{
			Name: to.Ptr(subnetID.VirtualNetworkName),
		},
		Subnet: &Subnet{
			ID: ipConfig.Properties.Subnet.ID,
		},
		StorageProfile: &VMScaleSetStorageProfile{
			VirtualMachineScaleSetStorageProfile: profile.StorageProfile,
		},
		RequirePublicIP:    to.Ptr(ipConfig.Properties.PublicIPAddressConfiguration != nil),
		SKUName:            found.SKU.Name,
		Capacity:           found.SKU.Capacity,
		ComputerNamePrefix: osProfile.ComputerNamePrefix,
		AdminUser:          osProfile.AdminUsername,
		SSHPublicKey:       sshKeys[0].KeyData,
		UserData:           fi.NewBytesResource(userData),
		Tags:               found.Tags,
		PrincipalID:        found.Identity.PrincipalID,
	}
	if ipConfig.Properties != nil && ipConfig.Properties.ApplicationSecurityGroups != nil {
		for _, asg := range ipConfig.Properties.ApplicationSecurityGroups {
			vmss.ApplicationSecurityGroups = append(vmss.ApplicationSecurityGroups, &ApplicationSecurityGroup{
				ID: asg.ID,
			})
		}
	}
	if loadBalancerID != nil {
		vmss.LoadBalancer = &LoadBalancer{
			Name: to.Ptr(loadBalancerID.LoadBalancerName),
		}
	}
	if found.Zones != nil {
		vmss.Zones = found.Zones
	}
	s.PrincipalID = found.Identity.PrincipalID
	return vmss, nil
}

func (s *VMScaleSet) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(s.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (s *VMScaleSet) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(s, c)
}

// CheckChanges returns an error if a change is not allowed.
func (s *VMScaleSet) CheckChanges(a, e, changes *VMScaleSet) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchangeable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a VM Scale Set.
func (s *VMScaleSet) RenderAzure(t *azure.AzureAPITarget, a, e, changes *VMScaleSet) error {
	if a == nil {
		klog.Infof("Creating a new VM Scale Set with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a VM Scale Set with name: %s", fi.ValueOf(e.Name))
	}

	name := *e.Name

	var customData *string
	if e.UserData != nil {
		d, err := fi.ResourceAsBytes(e.UserData)
		if err != nil {
			return fmt.Errorf("error rendering UserData: %s", err)
		}
		customData = to.Ptr(base64.StdEncoding.EncodeToString(d))
	}

	osProfile := &compute.VirtualMachineScaleSetOSProfile{
		ComputerNamePrefix: e.ComputerNamePrefix,
		AdminUsername:      e.AdminUser,
		LinuxConfiguration: &compute.LinuxConfiguration{
			SSH: &compute.SSHConfiguration{
				PublicKeys: []*compute.SSHPublicKey{
					{
						Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", *e.AdminUser)),
						KeyData: to.Ptr(*e.SSHPublicKey),
					},
				},
			},
			DisablePasswordAuthentication: to.Ptr(true),
		},
	}

	subnetID := azure.SubnetID{
		SubscriptionID:     t.Cloud.SubscriptionID(),
		ResourceGroupName:  *e.ResourceGroup.Name,
		VirtualNetworkName: *e.VirtualNetwork.Name,
		SubnetName:         *e.Subnet.Name,
	}
	var asgs []*compute.SubResource
	for _, asg := range e.ApplicationSecurityGroups {
		asgs = append(asgs, &compute.SubResource{
			ID: asg.ID,
		})
	}
	ipConfigProperties := &compute.VirtualMachineScaleSetIPConfigurationProperties{
		Subnet: &compute.APIEntityReference{
			ID: to.Ptr(subnetID.String()),
		},
		Primary:                   to.Ptr(true),
		PrivateIPAddressVersion:   to.Ptr(compute.IPVersionIPv4),
		ApplicationSecurityGroups: asgs,
	}
	if *e.RequirePublicIP {
		ipConfigProperties.PublicIPAddressConfiguration = &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
			Name: to.Ptr(name + "-publicipconfig"),
			Properties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
				PublicIPAddressVersion: to.Ptr(compute.IPVersionIPv4),
			},
		}
	}
	if e.LoadBalancer != nil {
		loadBalancerID := azure.LoadBalancerID{
			SubscriptionID:    t.Cloud.SubscriptionID(),
			ResourceGroupName: *e.ResourceGroup.Name,
			LoadBalancerName:  *e.LoadBalancer.Name,
		}
		ipConfigProperties.LoadBalancerBackendAddressPools = []*compute.SubResource{
			{
				ID: to.Ptr(loadBalancerID.String()),
			},
		}
	}

	networkConfig := &compute.VirtualMachineScaleSetNetworkConfiguration{
		Name: to.Ptr(name),
		Properties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
			Primary:            to.Ptr(true),
			EnableIPForwarding: to.Ptr(true),
			IPConfigurations: []*compute.VirtualMachineScaleSetIPConfiguration{
				{
					Name:       to.Ptr(name),
					Properties: ipConfigProperties,
				},
			},
		},
	}

	vmss := compute.VirtualMachineScaleSet{
		Location: to.Ptr(t.Cloud.Region()),
		SKU: &compute.SKU{
			Name:     e.SKUName,
			Capacity: e.Capacity,
		},
		Properties: &compute.VirtualMachineScaleSetProperties{
			UpgradePolicy: &compute.UpgradePolicy{
				Mode: to.Ptr(compute.UpgradeModeManual),
			},
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				OSProfile:      osProfile,
				StorageProfile: e.StorageProfile.VirtualMachineScaleSetStorageProfile,
				UserData:       customData,
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: []*compute.VirtualMachineScaleSetNetworkConfiguration{
						networkConfig,
					},
				},
			},
		},
		// Assign a system-assigned managed identity so that
		// Azure creates an identity for VMs and provision
		// its credentials on the VMs.
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type: to.Ptr(compute.ResourceIdentityTypeSystemAssigned),
		},
		Tags:  e.Tags,
		Zones: e.Zones,
	}

	result, err := t.Cloud.VMScaleSet().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		name,
		vmss)
	if err != nil {
		return err
	}
	e.PrincipalID = result.Identity.PrincipalID
	return nil
}
