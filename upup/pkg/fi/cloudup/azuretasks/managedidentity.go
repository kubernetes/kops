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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// ManagedIdentity is an Azure User-Assigned Managed Identity.
// +kops:fitask
type ManagedIdentity struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ResourceGroup *ResourceGroup
	Tags          map[string]*string

	// ClientID is populated after creation — this is the UAMI's client ID
	// needed for workload identity azure.json configuration.
	ClientID *string
	// PrincipalID is populated after creation — this is the UAMI's principal ID
	// needed for role assignments.
	PrincipalID *string
}

var (
	_ fi.CloudupTask   = &ManagedIdentity{}
	_ fi.CompareWithID = &ManagedIdentity{}
)

// CompareWithID returns the Name of the managed identity.
func (m *ManagedIdentity) CompareWithID() *string {
	return m.Name
}

// Find discovers the ManagedIdentity in the cloud provider.
func (m *ManagedIdentity) Find(c *fi.CloudupContext) (*ManagedIdentity, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	found, err := cloud.ManagedIdentity().Get(context.TODO(), fi.ValueOf(m.ResourceGroup.Name), fi.ValueOf(m.Name))
	if err != nil {
		var respErr *azcore.ResponseError
		if ok := errors.As(err, &respErr); ok && respErr.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("getting managed identity %q: %w", fi.ValueOf(m.Name), err)
	}

	result := &ManagedIdentity{
		Name:          m.Name,
		Lifecycle:     m.Lifecycle,
		ResourceGroup: m.ResourceGroup,
		Tags:          found.Tags,
	}
	if found.Properties != nil {
		result.ClientID = found.Properties.ClientID
		result.PrincipalID = found.Properties.PrincipalID
		// Also populate the expected task so dependent tasks (e.g. RoleAssignment)
		// can read these server-generated IDs when RenderAzure is skipped because
		// no changes are needed. Matches the VMScaleSet.Find convention.
		m.ClientID = found.Properties.ClientID
		m.PrincipalID = found.Properties.PrincipalID
	}
	return result, nil
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

// RenderAzure creates or updates a managed identity.
func (*ManagedIdentity) RenderAzure(t *azure.AzureAPITarget, a, e, changes *ManagedIdentity) error {
	if a == nil {
		klog.Infof("Creating a new Managed Identity with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Managed Identity with name: %s", fi.ValueOf(e.Name))
	}

	result, err := t.Cloud.ManagedIdentity().CreateOrUpdate(
		context.TODO(),
		fi.ValueOf(e.ResourceGroup.Name),
		fi.ValueOf(e.Name),
		armmsi.Identity{
			Location: to.Ptr(t.Cloud.Region()),
			Tags:     e.Tags,
		},
	)
	if err != nil {
		return fmt.Errorf("creating/updating managed identity: %w", err)
	}

	if result.Properties != nil {
		e.ClientID = result.Properties.ClientID
		e.PrincipalID = result.Properties.PrincipalID
	}

	return nil
}
