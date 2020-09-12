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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

//go:generate fitask -type=VirtualNetwork

// VirtualNetwork is an Azure Virtual Network.
type VirtualNetwork struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ResourceGroup *ResourceGroup
	CIDR          *string
	Tags          map[string]*string
}

var _ fi.Task = &VirtualNetwork{}
var _ fi.CompareWithID = &VirtualNetwork{}

// CompareWithID returns the Name of the VM Scale Set.
func (n *VirtualNetwork) CompareWithID() *string {
	return n.Name
}

// Find discovers the VirtualNetwork in the cloud provider.
func (n *VirtualNetwork) Find(c *fi.Context) (*VirtualNetwork, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	l, err := cloud.VirtualNetwork().List(context.TODO(), *n.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.VirtualNetwork
	for _, v := range l {
		if *v.Name == *n.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	addrPrefixes := *found.AddressSpace.AddressPrefixes
	if len(addrPrefixes) != 1 {
		return nil, fmt.Errorf("expected exactly one address prefix, but got %+v", addrPrefixes)
	}
	return &VirtualNetwork{
		Name:      n.Name,
		Lifecycle: n.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: n.ResourceGroup.Name,
		},
		CIDR: to.StringPtr(addrPrefixes[0]),
		Tags: found.Tags,
	}, nil
}

// Run implements fi.Task.Run.
func (n *VirtualNetwork) Run(c *fi.Context) error {
	c.Cloud.(azure.AzureCloud).AddClusterTags(n.Tags)
	return fi.DefaultDeltaRunMethod(n, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*VirtualNetwork) CheckChanges(a, e, changes *VirtualNetwork) error {
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
	if changes.CIDR != nil {
		return fi.CannotChangeField("CIDR")
	}
	return nil
}

// RenderAzure creates or updates a Virtual Network.
func (*VirtualNetwork) RenderAzure(t *azure.AzureAPITarget, a, e, changes *VirtualNetwork) error {
	if a == nil {
		klog.Infof("Creating a new Virtual Network with name: %s", fi.StringValue(e.Name))
	} else {
		// Only allow tags to be updated.
		if changes.Tags == nil {
			return nil
		}
		// TODO(kenji): Fix this. Update is not supported yet as updating the tag will recreate a virtual network,
		// which causes an InUseSubnetCannotBeDeleted error.
		klog.Infof("Skip updating a Virtual Network with name: %s", fi.StringValue(e.Name))
		return nil
	}

	vnet := network.VirtualNetwork{
		Location: to.StringPtr(t.Cloud.Region()),
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{*e.CIDR},
			},
		},
		Tags: e.Tags,
	}
	return t.Cloud.VirtualNetwork().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		vnet)
}
