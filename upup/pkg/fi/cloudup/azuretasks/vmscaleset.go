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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// SubnetID contains the resource ID/names required to construct a subnet ID.
type SubnetID struct {
	SubscriptionID     string
	ResourceGroupName  string
	VirtualNetworkName string
	SubnetName         string
}

// String returns the subnet ID in the path format.
func (s *SubnetID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/subnets/%s",
		s.SubscriptionID,
		s.ResourceGroupName,
		s.VirtualNetworkName,
		s.SubnetName)
}

// ParseSubnetID parses a given subnet ID string and returns a SubnetID.
func ParseSubnetID(s string) (*SubnetID, error) {
	l := strings.Split(s, "/")
	if len(l) != 11 {
		return nil, fmt.Errorf("malformed format of subnet ID: %s, %d", s, len(l))
	}
	return &SubnetID{
		SubscriptionID:     l[2],
		ResourceGroupName:  l[4],
		VirtualNetworkName: l[8],
		SubnetName:         l[10],
	}, nil
}

// loadBalancerID contains the resource ID/names required to construct a loadbalancer ID.
type loadBalancerID struct {
	SubscriptionID    string
	ResourceGroupName string
	LoadBalancerName  string
}

// String returns the loadbalancer ID in the path format.
func (lb *loadBalancerID) String() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadbalancers/%s/backendAddressPools/LoadBalancerBackEnd",
		lb.SubscriptionID,
		lb.ResourceGroupName,
		lb.LoadBalancerName,
	)
}

// parseLoadBalancerID parses a given loadbalancer ID string and returns a loadBalancerID.
func parseLoadBalancerID(lb string) (*loadBalancerID, error) {
	l := strings.Split(lb, "/")
	if len(l) != 11 {
		return nil, fmt.Errorf("malformed format of loadbalancer ID: %s, %d", lb, len(l))
	}
	return &loadBalancerID{
		SubscriptionID:    l[2],
		ResourceGroupName: l[4],
		LoadBalancerName:  l[8],
	}, nil
}

// VMScaleSet is an Azure VM Scale Set.
// +kops:fitask
type VMScaleSet struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ResourceGroup  *ResourceGroup
	VirtualNetwork *VirtualNetwork
	Subnet         *Subnet
	StorageProfile *VMScaleSetStorageProfile
	// RequirePublicIP is set to true when VMs require public IPs.
	RequirePublicIP *bool
	// LoadBalancer is the Load Balancer object the VMs will use.
	LoadBalancer *LoadBalancer
	// SKUName specifies the SKU of of the VM Scale Set
	SKUName *string
	// Capacity specifies the number of virtual machines the VM Scale Set.
	Capacity *int64
	// ComputerNamePrefix is the prefix of each VM name of the form <prefix><base-36-instance-id>.
	// See https://docs.microsoft.com/en-us/azure/virtual-machine-scale-sets/virtual-machine-scale-sets-instance-ids.
	ComputerNamePrefix *string
	// AdmnUser specifies the name of the administrative account.
	AdminUser    *string
	SSHPublicKey *string
	// CustomData is the user data configuration
	CustomData  fi.Resource
	Tags        map[string]*string
	Zones       []string
	PrincipalID *string
}

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

var _ fi.HasDependencies = &VMScaleSetStorageProfile{}

// GetDependencies returns a slice of tasks on which the tasks depends on.
func (p *VMScaleSetStorageProfile) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

var (
	_ fi.Task          = &VMScaleSet{}
	_ fi.CompareWithID = &VMScaleSet{}
)

// CompareWithID returns the Name of the VM Scale Set.
func (s *VMScaleSet) CompareWithID() *string {
	return s.Name
}

// Find discovers the VMScaleSet in the cloud provider.
func (s *VMScaleSet) Find(c *fi.Context) (*VMScaleSet, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	l, err := cloud.VMScaleSet().List(context.TODO(), *s.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *compute.VirtualMachineScaleSet
	for _, v := range l {
		if *v.Name == *s.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	profile := found.VirtualMachineProfile

	nwConfigs := *profile.NetworkProfile.NetworkInterfaceConfigurations
	if len(nwConfigs) != 1 {
		return nil, fmt.Errorf("unexpected number of network configs found for VM ScaleSet %s: %d", *s.Name, len(nwConfigs))
	}
	nwConfig := nwConfigs[0]
	ipConfigs := *nwConfig.VirtualMachineScaleSetNetworkConfigurationProperties.IPConfigurations
	if len(ipConfigs) != 1 {
		return nil, fmt.Errorf("unexpected number of IP configs found for VM ScaleSet %s: %d", *s.Name, len(ipConfigs))
	}
	ipConfig := ipConfigs[0]
	subnetID, err := ParseSubnetID(*ipConfig.Subnet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subnet ID %s", *ipConfig.Subnet.ID)
	}

	var loadBalancerID *loadBalancerID
	if ipConfig.LoadBalancerBackendAddressPools != nil {
		for _, i := range *ipConfig.LoadBalancerBackendAddressPools {
			if !strings.Contains(*i.ID, "api") {
				continue
			}
			loadBalancerID, err = parseLoadBalancerID(*i.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse loadbalancer ID %s", *ipConfig.Subnet.ID)
			}
		}
	}

	osProfile := profile.OsProfile
	sshKeys := *osProfile.LinuxConfiguration.SSH.PublicKeys
	if len(sshKeys) != 1 {
		return nil, fmt.Errorf("unexpected number of SSH keys found for VM ScaleSet %s: %d", *s.Name, len(sshKeys))
	}

	// TODO(kenji): Do not check custom data as Azure doesn't
	// populate (https://github.com/Azure/azure-cli/issues/5866).
	// Find a way to work around this.
	vmss := &VMScaleSet{
		Name:      s.Name,
		Lifecycle: s.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: s.ResourceGroup.Name,
		},
		VirtualNetwork: &VirtualNetwork{
			Name: to.StringPtr(subnetID.VirtualNetworkName),
		},
		Subnet: &Subnet{
			Name: to.StringPtr(subnetID.SubnetName),
		},
		StorageProfile: &VMScaleSetStorageProfile{
			VirtualMachineScaleSetStorageProfile: profile.StorageProfile,
		},
		RequirePublicIP:    to.BoolPtr(ipConfig.PublicIPAddressConfiguration != nil),
		SKUName:            found.Sku.Name,
		Capacity:           found.Sku.Capacity,
		ComputerNamePrefix: osProfile.ComputerNamePrefix,
		AdminUser:          osProfile.AdminUsername,
		SSHPublicKey:       sshKeys[0].KeyData,
		Tags:               found.Tags,
		PrincipalID:        found.Identity.PrincipalID,
	}
	if loadBalancerID != nil {
		vmss.LoadBalancer = &LoadBalancer{
			Name: to.StringPtr(loadBalancerID.LoadBalancerName),
		}
	}
	if found.Zones != nil {
		vmss.Zones = *found.Zones
	}
	return vmss, nil
}

// Run implements fi.Task.Run.
func (s *VMScaleSet) Run(c *fi.Context) error {
	c.Cloud.(azure.AzureCloud).AddClusterTags(s.Tags)
	return fi.DefaultDeltaRunMethod(s, c)
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
		klog.Infof("Creating a new VM Scale Set with name: %s", fi.StringValue(e.Name))
	} else {
		klog.Infof("Updating a VM Scale Set with name: %s", fi.StringValue(e.Name))
	}

	name := *e.Name

	var customData *string
	if e.CustomData != nil {
		d, err := fi.ResourceAsBytes(e.CustomData)
		if err != nil {
			return fmt.Errorf("error rendering CustomData: %s", err)
		}
		customData = to.StringPtr(base64.StdEncoding.EncodeToString(d))
	}

	osProfile := &compute.VirtualMachineScaleSetOSProfile{
		ComputerNamePrefix: e.ComputerNamePrefix,
		AdminUsername:      e.AdminUser,
		CustomData:         customData,
		LinuxConfiguration: &compute.LinuxConfiguration{
			SSH: &compute.SSHConfiguration{
				PublicKeys: &[]compute.SSHPublicKey{
					{
						Path:    to.StringPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", *e.AdminUser)),
						KeyData: to.StringPtr(*e.SSHPublicKey),
					},
				},
			},
			DisablePasswordAuthentication: to.BoolPtr(true),
		},
	}

	subnetID := SubnetID{
		SubscriptionID:     t.Cloud.SubscriptionID(),
		ResourceGroupName:  *e.ResourceGroup.Name,
		VirtualNetworkName: *e.VirtualNetwork.Name,
		SubnetName:         *e.Subnet.Name,
	}
	ipConfigProperties := &compute.VirtualMachineScaleSetIPConfigurationProperties{
		Subnet: &compute.APIEntityReference{
			ID: to.StringPtr(subnetID.String()),
		},
		Primary:                 to.BoolPtr(true),
		PrivateIPAddressVersion: compute.IPv4,
	}
	if *e.RequirePublicIP {
		ipConfigProperties.PublicIPAddressConfiguration = &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
			Name: to.StringPtr(name + "-publicipconfig"),
			VirtualMachineScaleSetPublicIPAddressConfigurationProperties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
				PublicIPAddressVersion: compute.IPv4,
			},
		}
	}
	if e.LoadBalancer != nil {
		loadBalancerID := loadBalancerID{
			SubscriptionID:    t.Cloud.SubscriptionID(),
			ResourceGroupName: *e.ResourceGroup.Name,
			LoadBalancerName:  *e.LoadBalancer.Name,
		}
		ipConfigProperties.LoadBalancerBackendAddressPools = &[]compute.SubResource{
			{
				ID: to.StringPtr(loadBalancerID.String()),
			},
		}
	}

	networkConfig := compute.VirtualMachineScaleSetNetworkConfiguration{
		Name: to.StringPtr(name + "-netconfig"),
		VirtualMachineScaleSetNetworkConfigurationProperties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
			Primary:            to.BoolPtr(true),
			EnableIPForwarding: to.BoolPtr(true),
			IPConfigurations: &[]compute.VirtualMachineScaleSetIPConfiguration{
				{
					Name: to.StringPtr(name + "-ipconfig"),
					VirtualMachineScaleSetIPConfigurationProperties: ipConfigProperties,
				},
			},
		},
	}

	vmss := compute.VirtualMachineScaleSet{
		Location: to.StringPtr(t.Cloud.Region()),
		Sku: &compute.Sku{
			Name:     e.SKUName,
			Capacity: e.Capacity,
		},
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			UpgradePolicy: &compute.UpgradePolicy{
				Mode: compute.UpgradeModeManual,
			},
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				OsProfile:      osProfile,
				StorageProfile: e.StorageProfile.VirtualMachineScaleSetStorageProfile,
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: &[]compute.VirtualMachineScaleSetNetworkConfiguration{
						networkConfig,
					},
				},
			},
		},
		// Assign a system-assigned managed identity so that
		// Azure creates an identity for VMs and provision
		// its credentials on the VMs.
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type: compute.ResourceIdentityTypeSystemAssigned,
		},
		Tags:  e.Tags,
		Zones: &e.Zones,
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
