/*
Copyright 2017 The Kubernetes Authors.

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
	"github.com/golang/glog"
	"google.golang.org/api/iam/v1"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"reflect"
	"strings"
)

// ServiceAccount models a GCE ServiceAccount
//go:generate fitask -type=ServiceAccount
type ServiceAccount struct {
	Name *string

	// AccountId is the GCE account id:
	// The account id that is used to generate the service account email
	// address and a stable unique id.
	// It is unique within a project, must be 6-30 characters long, and
	// match the regular expression
	// `[a-z]([-a-z0-9]*[a-z0-9])` to comply with RFC1035.
	AccountId *string

	// DisplayName is an optional description for the service account
	DisplayName *string

	Lifecycle *fi.Lifecycle

	// gceName is the Name of the ServiceAccount, if it is found in GCE
	gceName *string
	// etag is the Etag of the ServiceAccount, if it is found in GCE
	etag *string
}

var _ fi.CompareWithID = &ServiceAccount{}

func (e *ServiceAccount) CompareWithID() *string {
	return e.Name
}

func (e *ServiceAccount) projectName(c *fi.Context) string {
	cloud := c.Cloud.(gce.GCECloud)
	return "project/" + cloud.Project()
}

func (e *ServiceAccount) Find(c *fi.Context) (*ServiceAccount, error) {
	cloud := c.Cloud.(gce.GCECloud)

	projectName := e.projectName(c)

	accountId := fi.StringValue(e.AccountId)

	var matches []*iam.ServiceAccount

	ctx := context.Background()
	err := cloud.IAM().Projects.ServiceAccounts.List(projectName).Pages(ctx, func(page *iam.ListServiceAccountsResponse) error {
		for _, a := range page.Accounts {
			tokens := strings.SplitN(a.Email, "@", 2)
			if len(tokens) != 0 && tokens[0] == accountId {
				matches = append(matches, a)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting ServiceAccount %q: %v", accountId, err)
	}

	if len(matches) == 0 {
		return nil, nil
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("found multiple ServiceAccounts matching %q", accountId)
	}

	match := matches[0]
	actual := &ServiceAccount{}
	if match.DisplayName != "" {
		actual.DisplayName = &match.DisplayName
	}
	actual.AccountId = &accountId

	e.gceName = fi.String(match.Name)
	e.etag = fi.String(match.Etag)

	// Prevent spurious changes
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle
	actual.gceName = e.gceName
	actual.etag = e.etag

	return actual, nil
}

func (e *ServiceAccount) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *ServiceAccount) CheckChanges(a, e, changes *ServiceAccount) error {
	if fi.StringValue(e.AccountId) == "" {
		return fi.RequiredField("AccountId")
	}
	return nil
}

func (_ *ServiceAccount) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ServiceAccount) error {
	accountId := fi.StringValue(e.AccountId)

	if a == nil {
		glog.V(2).Infof("Creating ServiceAccount with AccountId: %q", accountId)

		request := &iam.CreateServiceAccountRequest{
			AccountId: accountId,
			ServiceAccount: &iam.ServiceAccount{
				DisplayName: *e.DisplayName,
			},
		}

		path := "project/" + t.Cloud.Project()
		_, err := t.Cloud.IAM().Projects.ServiceAccounts.Create(path, request).Do()
		if err != nil {
			return fmt.Errorf("error creating ServiceAccount: %v", err)
		}
	} else {
		if changes.DisplayName != nil {
			glog.V(2).Infof("Updating ServiceAccount DisplayName for %q", fi.StringValue(e.gceName))

			serviceAccount := &iam.ServiceAccount{
				DisplayName: fi.StringValue(e.DisplayName),
				Etag:        fi.StringValue(e.etag),
			}

			_, err := t.Cloud.IAM().Projects.ServiceAccounts.Update(fi.StringValue(e.gceName), serviceAccount).Do()
			if err != nil {
				return fmt.Errorf("error updating ServiceAccount: %v", err)
			}

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
	AccountId   *string `json:"account_id,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
}

func (_ *ServiceAccount) RenderServiceAccount(t *terraform.TerraformTarget, a, e, changes *ServiceAccount) error {
	tf := &terraformServiceAccount{
		AccountId:   e.AccountId,
		DisplayName: e.DisplayName,
	}
	return t.RenderResource("google_service_account", *e.Name, tf)
}
