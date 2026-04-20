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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// FederatedIdentityCredential links a Kubernetes service account to an Azure User-Assigned Managed Identity.
// +kops:fitask
type FederatedIdentityCredential struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ManagedIdentity *ManagedIdentity
	ResourceGroup   *ResourceGroup
	Issuer          *string
	Subject         *string
	Audiences       []*string
}

var (
	_ fi.CloudupTask   = &FederatedIdentityCredential{}
	_ fi.CompareWithID = &FederatedIdentityCredential{}
)

// CompareWithID returns the Name.
func (f *FederatedIdentityCredential) CompareWithID() *string {
	return f.Name
}

// Find discovers the FederatedIdentityCredential in the cloud provider.
func (f *FederatedIdentityCredential) Find(c *fi.CloudupContext) (*FederatedIdentityCredential, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	found, err := cloud.FederatedIdentityCredential().Get(
		context.TODO(),
		fi.ValueOf(f.ResourceGroup.Name),
		fi.ValueOf(f.ManagedIdentity.Name),
		fi.ValueOf(f.Name),
	)
	if err != nil {
		var respErr *azcore.ResponseError
		if ok := errors.As(err, &respErr); ok && respErr.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("getting federated identity credential %q: %w", fi.ValueOf(f.Name), err)
	}

	result := &FederatedIdentityCredential{
		Name:            f.Name,
		Lifecycle:       f.Lifecycle,
		ManagedIdentity: f.ManagedIdentity,
		ResourceGroup:   f.ResourceGroup,
	}
	if found.Properties != nil {
		result.Issuer = found.Properties.Issuer
		result.Subject = found.Properties.Subject
		result.Audiences = found.Properties.Audiences
	}
	return result, nil
}

// Run implements fi.Task.Run.
func (f *FederatedIdentityCredential) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(f, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*FederatedIdentityCredential) CheckChanges(a, e, changes *FederatedIdentityCredential) error {
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

// RenderAzure creates or updates a federated identity credential.
func (*FederatedIdentityCredential) RenderAzure(t *azure.AzureAPITarget, a, e, changes *FederatedIdentityCredential) error {
	if a == nil {
		klog.Infof("Creating a new Federated Identity Credential with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Federated Identity Credential with name: %s", fi.ValueOf(e.Name))
	}

	_, err := t.Cloud.FederatedIdentityCredential().CreateOrUpdate(
		context.TODO(),
		fi.ValueOf(e.ResourceGroup.Name),
		fi.ValueOf(e.ManagedIdentity.Name),
		fi.ValueOf(e.Name),
		armmsi.FederatedIdentityCredential{
			Properties: &armmsi.FederatedIdentityCredentialProperties{
				Issuer:    e.Issuer,
				Subject:   e.Subject,
				Audiences: e.Audiences,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("creating/updating federated identity credential: %w", err)
	}

	return nil
}
