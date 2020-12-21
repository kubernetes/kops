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

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

//go:generate fitask -type=ResourceGroup

// ResourceGroup is an Azure resource group.
type ResourceGroup struct {
	Name      *string
	Lifecycle *fi.Lifecycle
	Tags      map[string]*string
	// Shared is set to true if the resource group is not solely
	// owned by kops. The resource group will not be deleted when
	// a cluster is destroyed.
	Shared *bool
}

var _ fi.Task = &ResourceGroup{}
var _ fi.CompareWithID = &ResourceGroup{}

// CompareWithID returns the Name of the VM Scale Set.
func (r *ResourceGroup) CompareWithID() *string {
	return r.Name
}

// Find discovers the ResourceGroup in the cloud provider.
func (r *ResourceGroup) Find(c *fi.Context) (*ResourceGroup, error) {
	cloud := c.Cloud.(azure.AzureCloud)
	l, err := cloud.ResourceGroup().List(context.TODO(), "" /* filter*/)
	if err != nil {
		return nil, err
	}
	var found *resources.Group
	for _, rg := range l {
		if *rg.Name == *r.Name {
			found = &rg
			break
		}
	}
	if found == nil {
		return nil, nil
	}
	return &ResourceGroup{
		Name:      to.StringPtr(*found.Name),
		Lifecycle: r.Lifecycle,
		Tags:      found.Tags,
		// To prevent spurious comparison failures. Follow awstask.VPC.
		Shared: r.Shared,
	}, nil
}

// Run implements fi.Task.Run.
func (r *ResourceGroup) Run(c *fi.Context) error {
	c.Cloud.(azure.AzureCloud).AddClusterTags(r.Tags)
	return fi.DefaultDeltaRunMethod(r, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*ResourceGroup) CheckChanges(a, e, changes *ResourceGroup) error {
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

// RenderAzure creates or updates a resource group.
func (*ResourceGroup) RenderAzure(t *azure.AzureAPITarget, a, e, changes *ResourceGroup) error {
	if a == nil {
		klog.Infof("Creating a new Resource Group with name: %s", fi.StringValue(e.Name))
	} else {
		klog.Infof("Updating a Resource Group with name: %s", fi.StringValue(e.Name))
	}
	return t.Cloud.ResourceGroup().CreateOrUpdate(
		context.TODO(),
		*e.Name,
		resources.Group{
			Location: to.StringPtr(t.Cloud.Region()),
			Tags:     e.Tags,
		})
}
