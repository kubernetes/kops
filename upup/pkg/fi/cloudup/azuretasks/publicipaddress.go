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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// PublicIPAddress is an Azure Cloud Public IP Address
// +kops:fitask
type PublicIPAddress struct {
	Name          *string
	Lifecycle     fi.Lifecycle
	ResourceGroup *ResourceGroup

	Tags map[string]*string
}

var (
	_ fi.Task          = &PublicIPAddress{}
	_ fi.CompareWithID = &PublicIPAddress{}
	_ fi.TaskNormalize = &PublicIPAddress{}
)

// CompareWithID returns the Name of the Public IP Address
func (p *PublicIPAddress) CompareWithID() *string {
	return p.Name
}

// Find discovers the Public IP Address in the cloud provider
func (p *PublicIPAddress) Find(c *fi.Context) (*PublicIPAddress, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	l, err := cloud.PublicIPAddress().List(context.TODO(), *p.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.PublicIPAddress
	for _, v := range l {
		if *v.Name == *p.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	return &PublicIPAddress{
		Name:      p.Name,
		Lifecycle: p.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: p.ResourceGroup.Name,
		},

		Tags: found.Tags,
	}, nil
}

func (p *PublicIPAddress) Normalize(c *fi.Context) error {
	c.Cloud.(azure.AzureCloud).AddClusterTags(p.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (p *PublicIPAddress) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(p, c)
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
		Location: to.StringPtr(t.Cloud.Region()),
		Name:     to.StringPtr(*e.Name),
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAddressVersion:   network.IPv4,
			PublicIPAllocationMethod: network.Dynamic,
		},
		Tags: e.Tags,
	}

	return t.Cloud.PublicIPAddress().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		p)
}
