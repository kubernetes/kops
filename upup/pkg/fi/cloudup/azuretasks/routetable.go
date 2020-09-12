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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

//go:generate fitask -type=RouteTable

// RouteTable is an Azure Route Table.
type RouteTable struct {
	Name          *string
	Lifecycle     *fi.Lifecycle
	ResourceGroup *ResourceGroup
	Tags          map[string]*string
}

var _ fi.Task = &RouteTable{}
var _ fi.CompareWithID = &RouteTable{}

// CompareWithID returns the Name of the VM Scale Set.
func (r *RouteTable) CompareWithID() *string {
	return r.Name
}

// Find discovers the RouteTable in the cloud provider.
func (r *RouteTable) Find(c *fi.Context) (*RouteTable, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	l, err := cloud.RouteTable().List(context.TODO(), *r.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.RouteTable
	for _, v := range l {
		if *v.Name == *r.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}
	return &RouteTable{
		Name:      r.Name,
		Lifecycle: r.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: r.ResourceGroup.Name,
		},
		Tags: found.Tags,
	}, nil
}

// Run implements fi.Task.Run.
func (r *RouteTable) Run(c *fi.Context) error {
	c.Cloud.(azure.AzureCloud).AddClusterTags(r.Tags)
	return fi.DefaultDeltaRunMethod(r, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*RouteTable) CheckChanges(a, e, changes *RouteTable) error {
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

// RenderAzure creates or updates a Route Table.
func (*RouteTable) RenderAzure(t *azure.AzureAPITarget, a, e, changes *RouteTable) error {
	if a == nil {
		klog.Infof("Creating a new Route Table with name: %s", fi.StringValue(e.Name))
	} else {
		klog.Infof("Updating a Route Table with name: %s", fi.StringValue(e.Name))
	}

	rt := network.RouteTable{
		Location: to.StringPtr(t.Cloud.Region()),
		Tags:     e.Tags,
	}
	return t.Cloud.RouteTable().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		rt)
}
