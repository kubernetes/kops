/*
Copyright 2023 The Kubernetes Authors.

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

// NatGateway is an Azure Nat Gateway
// +kops:fitask
type NatGateway struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID                *string
	PublicIPAddresses []*PublicIPAddress
	ResourceGroup     *ResourceGroup

	Tags map[string]*string
}

var (
	_ fi.CloudupTask          = &NatGateway{}
	_ fi.CompareWithID        = &NatGateway{}
	_ fi.CloudupTaskNormalize = &NatGateway{}
)

// CompareWithID returns the Name of the Nat Gateway
func (ngw *NatGateway) CompareWithID() *string {
	return ngw.ID
}

// Find discovers the Nat Gateway in the cloud provider
func (ngw *NatGateway) Find(c *fi.CloudupContext) (*NatGateway, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	l, err := cloud.NatGateway().List(context.TODO(), *ngw.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.NatGateway
	for _, v := range l {
		if *v.Name == *ngw.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	ngw.ID = found.ID

	var pips []*PublicIPAddress
	if found.PublicIPAddresses != nil {
		for _, pip := range *found.PublicIPAddresses {
			pips = append(pips, &PublicIPAddress{ID: pip.ID})
		}
	}

	return &NatGateway{
		Name:              ngw.Name,
		Lifecycle:         ngw.Lifecycle,
		ResourceGroup:     &ResourceGroup{Name: ngw.ResourceGroup.Name},
		ID:                found.ID,
		PublicIPAddresses: pips,
		Tags:              found.Tags,
	}, nil
}

func (ngw *NatGateway) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(ngw.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (ngw *NatGateway) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(ngw, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*NatGateway) CheckChanges(a, e, changes *NatGateway) error {
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

// RenderAzure creates or updates a Nat Gateway.
func (*NatGateway) RenderAzure(t *azure.AzureAPITarget, a, e, changes *NatGateway) error {
	if a == nil {
		klog.Infof("Creating a new Nat Gateway with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Nat Gateway with name: %s", fi.ValueOf(e.Name))
	}

	p := network.NatGateway{
		Location:                   to.StringPtr(t.Cloud.Region()),
		Name:                       to.StringPtr(*e.Name),
		NatGatewayPropertiesFormat: &network.NatGatewayPropertiesFormat{},
		Sku: &network.NatGatewaySku{
			Name: network.NatGatewaySkuNameStandard,
		},
		Tags: e.Tags,
	}

	if len(e.PublicIPAddresses) > 0 {
		var pips []network.SubResource
		for _, pip := range e.PublicIPAddresses {
			pips = append(pips, network.SubResource{ID: pip.ID})
		}
		p.PublicIPAddresses = &pips
	}

	ngw, err := t.Cloud.NatGateway().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		p)
	if err != nil {
		return err
	}

	e.ID = ngw.ID

	return nil
}
