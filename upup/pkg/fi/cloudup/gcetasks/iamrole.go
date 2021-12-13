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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// IAMRole represents an IAM custom role on a project
// +kops:fitask
type IAMRole struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Project     *string
	RoleID      *string
	Description *string
	Title       *string
	Permissions []string
}

var _ fi.CompareWithID = &IAMRole{}

func (e *IAMRole) CompareWithID() *string {
	return e.Name
}

func (e *IAMRole) Find(c *fi.Context) (*IAMRole, error) {
	ctx := context.TODO()

	cloud := c.Cloud.(gce.GCECloud)

	projectID := fi.StringValue(e.Project)
	roleID := fi.StringValue(e.RoleID)

	fqn := "projects/" + projectID + "/roles/" + roleID
	role, err := cloud.IAM().Projects.Roles.Get(fqn).Context(ctx).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting custom IAM role %q: %w", fqn, err)
	}

	if role.Deleted {
		return nil, nil
	}

	actual := &IAMRole{}
	actual.Project = e.Project
	actual.RoleID = e.RoleID
	actual.Description = &role.Description
	actual.Title = &role.Title
	actual.Permissions = role.IncludedPermissions

	// Ignore "system" fields
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *IAMRole) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *IAMRole) CheckChanges(a, e, changes *IAMRole) error {
	if fi.StringValue(e.Project) == "" {
		return fi.RequiredField("Project")
	}
	if fi.StringValue(e.RoleID) == "" {
		return fi.RequiredField("RoleID")
	}
	return nil
}

func (_ *IAMRole) RenderGCE(t *gce.GCEAPITarget, a, e, changes *IAMRole) error {
	ctx := context.TODO()

	cloud := t.Cloud.(gce.GCECloud)

	projectID := fi.StringValue(e.Project)
	roleID := fi.StringValue(e.RoleID)

	fqn := "projects/" + projectID + "/roles/" + roleID

	role := &iam.Role{
		Description:         fi.StringValue(e.Description),
		Title:               fi.StringValue(e.Title),
		IncludedPermissions: e.Permissions,
	}

	if a == nil {
		request := &iam.CreateRoleRequest{
			RoleId: roleID,
			Role:   role,
		}
		parent := "projects/" + projectID
		_, err := cloud.IAM().Projects.Roles.Create(parent, request).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("error creating role %q: %w", fqn, err)
		}
	} else {
		if changes.Permissions != nil || changes.Description != nil || changes.Title != nil {
			_, err := cloud.IAM().Projects.Roles.Patch(fqn, role).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("error updating role %q: %w", fqn, err)
			}

			changes.Permissions = nil
			changes.Description = nil
			changes.Title = nil
		}

		empty := &IAMRole{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to IAMRole: %v", changes)
		}
	}

	return nil
}

// terraformIAMRole is the model for a terraform google_project_iam_custom_role rule
type terraformIAMRole struct {
	Project     *string  `json:"project,omitempty" cty:"project"`
	RoleID      *string  `json:"role_id,omitempty" cty:"role_id"`
	Description *string  `json:"description,omitempty" cty:"description"`
	Title       *string  `json:"title,omitempty" cty:"title"`
	Permissions []string `json:"permissions,omitempty" cty:"permissions"`
}

func (_ *IAMRole) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMRole) error {
	tf := &terraformIAMRole{
		Project:     e.Project,
		RoleID:      e.RoleID,
		Description: e.Description,
		Title:       e.Title,
		Permissions: e.Permissions,
	}

	return t.RenderResource("google_project_iam_custom_role", *e.Name, tf)
}
