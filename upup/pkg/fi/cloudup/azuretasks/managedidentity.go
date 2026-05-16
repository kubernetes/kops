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
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// ManagedIdentity is an Azure user-assigned managed identity.
// +kops:fitask
type ManagedIdentity struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ResourceGroup *ResourceGroup
	Tags          map[string]*string

	// PrincipalID is populated after creation and used by RoleAssignment.
	PrincipalID *string
	// ResourceID is populated after creation and used by VMScaleSet.
	ResourceID *string
}

var (
	_ fi.CloudupTask   = &ManagedIdentity{}
	_ fi.CompareWithID = &ManagedIdentity{}
)

// CompareWithID returns the Name of the ManagedIdentity.
func (m *ManagedIdentity) CompareWithID() *string {
	return m.Name
}

// Find discovers the ManagedIdentity in the cloud provider.
func (m *ManagedIdentity) Find(c *fi.CloudupContext) (*ManagedIdentity, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	found, err := cloud.ManagedIdentity().Get(context.TODO(), *m.ResourceGroup.Name, *m.Name)
	if err != nil {
		var azErr *azcore.ResponseError
		if errors.As(err, &azErr) {
			if azErr.ErrorCode == "ResourceNotFound" || azErr.ErrorCode == "ResourceGroupNotFound" {
				return nil, nil
			}
		}
		return nil, err
	}

	// Populate PrincipalID and ResourceID on the source object so that
	// dependent tasks (RoleAssignment, VMScaleSet) can reference them.
	m.PrincipalID = found.Properties.PrincipalID
	m.ResourceID = found.ID

	return &ManagedIdentity{
		Name:      found.Name,
		Lifecycle: m.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: m.ResourceGroup.Name,
		},
		Tags:        found.Tags,
		PrincipalID: found.Properties.PrincipalID,
		ResourceID:  found.ID,
	}, nil
}

// Run implements fi.Task.Run.
func (m *ManagedIdentity) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(m, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*ManagedIdentity) CheckChanges(a, e, changes *ManagedIdentity) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a user-assigned managed identity.
func (*ManagedIdentity) RenderAzure(t *azure.AzureAPITarget, a, e, changes *ManagedIdentity) error {
	if a == nil {
		klog.Infof("Creating a new Managed Identity with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Managed Identity with name: %s", fi.ValueOf(e.Name))
	}

	result, err := t.Cloud.ManagedIdentity().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		armmsi.Identity{
			Location: to.Ptr(t.Cloud.Region()),
			Tags:     e.Tags,
		},
	)
	if err != nil {
		return err
	}

	e.PrincipalID = result.Properties.PrincipalID
	e.ResourceID = result.ID
	return nil
}
