/*
Copyright 2021 The Kubernetes Authors.

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

package gcetasks

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/api/iam/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type ServiceAccount struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Email       *string
	Description *string
	DisplayName *string

	Shared *bool
}

var _ fi.CompareWithID = &ServiceAccount{}

func (e *ServiceAccount) CompareWithID() *string {
	return e.Email
}

func (e *ServiceAccount) Find(c *fi.Context) (*ServiceAccount, error) {
	cloud := c.Cloud.(gce.GCECloud)

	ctx := context.TODO()

	email := fi.StringValue(e.Email)

	if email == "default" {
		// Special case - the default serviceaccount always exists
		return e, nil
	}

	_, projectID, err := gce.SplitServiceAccountEmail(email)
	if err != nil {
		return nil, err
	}
	fqn := "projects/" + projectID + "/serviceAccounts/" + email
	sa, err := cloud.IAM().ServiceAccounts().Get(ctx, fqn)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing ServiceAccount %q: %w", fqn, err)
	}

	// Check the email actually matches what we expect
	if email != sa.Email {
		return nil, fmt.Errorf("found ServiceAccount but email did not match expected; got %q; want %q", sa.Email, email)
	}

	actual := &ServiceAccount{}
	actual.DisplayName = &sa.DisplayName
	actual.Description = &sa.Description
	actual.Email = &sa.Email

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Name = e.Name
	actual.Shared = e.Shared

	return actual, nil
}

func (e *ServiceAccount) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *ServiceAccount) CheckChanges(a, e, changes *ServiceAccount) error {
	return nil
}

func (_ *ServiceAccount) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ServiceAccount) error {
	ctx := context.TODO()

	cloud := t.Cloud

	email := fi.StringValue(e.Email)

	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the service account was found
		if a == nil {
			return fmt.Errorf("ServiceAccount with email %q not found", email)
		}
	}

	accountID, projectID, err := gce.SplitServiceAccountEmail(email)
	if err != nil {
		return err
	}

	fqn := "projects/" + projectID + "/serviceAccounts/" + email

	if a == nil {
		klog.V(2).Infof("Creating ServiceAccount %q", fqn)

		sa := &iam.CreateServiceAccountRequest{
			AccountId: accountID,
			ServiceAccount: &iam.ServiceAccount{
				Description: fi.StringValue(e.Description),
				DisplayName: fi.StringValue(e.DisplayName),
			},
		}

		created, err := cloud.IAM().ServiceAccounts().Create(ctx, "projects/"+projectID, sa)
		if err != nil {
			return fmt.Errorf("error creating ServiceAccount %q: %w", fqn, err)
		}
		if created.Email != email {
			return fmt.Errorf("created ServiceAccount did not have expected email; got %q; want %q", created.Email, email)
		}
	} else {
		if changes.Description != nil || changes.DisplayName != nil {
			sa := &iam.ServiceAccount{
				Email:       email,
				Description: fi.StringValue(e.Description),
				DisplayName: fi.StringValue(e.DisplayName),
			}

			_, err := cloud.IAM().ServiceAccounts().Update(ctx, fqn, sa)
			if err != nil {
				return fmt.Errorf("error creating ServiceAccount %q: %w", fqn, err)
			}

			changes.Description = nil
			changes.DisplayName = nil
		}

		empty := &ServiceAccount{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to ServiceAccount: %v", changes)
		}
	}

	return nil
}

type terraformServiceAccount struct {
	AccountID   *string `cty:"account_id"`
	ProjectID   *string `cty:"project"`
	Description *string `cty:"description"`
	DisplayName *string `cty:"display_name"`
}

func (_ *ServiceAccount) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ServiceAccount) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed
		return nil
	}

	email := fi.StringValue(e.Email)
	accountID, projectID, err := gce.SplitServiceAccountEmail(email)
	if err != nil {
		return err
	}

	tf := &terraformServiceAccount{
		AccountID:   &accountID,
		ProjectID:   &projectID,
		Description: e.Description,
		DisplayName: e.DisplayName,
	}

	return t.RenderResource("google_service_account", *e.Name, tf)
}

func (e *ServiceAccount) TerraformLink() *terraformWriter.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		email := fi.StringValue(e.Email)
		if email == "" {
			klog.Fatalf("Email must be set, if ServiceAccount is shared: %#v", e)
		}

		klog.V(4).Infof("reusing existing ServiceAccount %q", email)
		return terraformWriter.LiteralFromStringValue(email)
	}

	return terraformWriter.LiteralProperty("google_service_account", *e.Name, "email")
}
