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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// PublicIPAddress is an Azure Cloud Public IP Address
// +kops:fitask
type PublicIPAddress struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID            *string
	IPAddress     *string
	ResourceGroup *ResourceGroup

	// IPVersion is the IP version, e.g. network.IPVersionIPv4.
	IPVersion network.IPVersion
	// AllocationMethod is the allocation method, e.g. network.IPAllocationMethodStatic.
	AllocationMethod network.IPAllocationMethod
	// SKU is the public IP SKU, e.g. network.PublicIPAddressSKUNameStandard.
	SKU network.PublicIPAddressSKUName

	Tags map[string]*string
}

var (
	_ fi.CloudupTask          = &PublicIPAddress{}
	_ fi.CompareWithID        = &PublicIPAddress{}
	_ fi.CloudupTaskNormalize = &PublicIPAddress{}
)

// CompareWithID returns the Name of the Public IP Address
func (p *PublicIPAddress) CompareWithID() *string {
	return p.ID
}

// Find discovers the Public IP Address in the cloud provider
func (p *PublicIPAddress) Find(c *fi.CloudupContext) (*PublicIPAddress, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	l, err := cloud.PublicIPAddress().List(context.TODO(), *p.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.PublicIPAddress
	for _, v := range l {
		if *v.Name == *p.Name {
			found = v
			break
		}
	}
	if found == nil {
		return nil, nil
	}
	if found.Properties != nil && found.Properties.ProvisioningState != nil && *found.Properties.ProvisioningState == network.ProvisioningStateFailed {
		klog.Warningf("found public IP address %q in failed provisioning state", *p.Name)
		return nil, nil
	}

	p.ID = found.ID
	if found.Properties != nil {
		p.IPAddress = found.Properties.IPAddress
	}

	actual := &PublicIPAddress{
		Name:      p.Name,
		Lifecycle: p.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: p.ResourceGroup.Name,
		},
		ID:   found.ID,
		Tags: found.Tags,
	}
	if found.Properties != nil {
		actual.IPVersion = fi.ValueOf(found.Properties.PublicIPAddressVersion)
		actual.AllocationMethod = fi.ValueOf(found.Properties.PublicIPAllocationMethod)
		actual.IPAddress = found.Properties.IPAddress
	}
	if found.SKU != nil {
		actual.SKU = fi.ValueOf(found.SKU.Name)
	}
	return actual, nil
}

func (p *PublicIPAddress) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(p.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (p *PublicIPAddress) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(p, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*PublicIPAddress) CheckChanges(a, e, changes *PublicIPAddress) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchanegable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a Public IP Address.
func (*PublicIPAddress) RenderAzure(t *azure.AzureAPITarget, a, e, changes *PublicIPAddress) error {
	if a == nil {
		klog.Infof("Creating a new Public IP Address with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Public IP Address with name: %s", fi.ValueOf(e.Name))
	}

	p := network.PublicIPAddress{
		Location: to.Ptr(t.Cloud.Region()),
		Name:     to.Ptr(*e.Name),
		Properties: &network.PublicIPAddressPropertiesFormat{
			PublicIPAddressVersion:   to.Ptr(e.IPVersion),
			PublicIPAllocationMethod: to.Ptr(e.AllocationMethod),
		},
		SKU: &network.PublicIPAddressSKU{
			Name: to.Ptr(e.SKU),
		},
		Tags: e.Tags,
	}

	pip, err := t.Cloud.PublicIPAddress().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		p)
	if err != nil {
		return err
	}

	e.ID = pip.ID
	if pip.Properties != nil {
		e.IPAddress = pip.Properties.IPAddress
	}

	return nil
}
