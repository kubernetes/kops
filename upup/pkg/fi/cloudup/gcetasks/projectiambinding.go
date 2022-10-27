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

	"google.golang.org/api/cloudresourcemanager/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// ProjectIAMBinding represents an IAM rule on a project
// +kops:fitask
type ProjectIAMBinding struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Project *string
	Member  *string
	Role    *string
}

var _ fi.CompareWithID = &ProjectIAMBinding{}

func (e *ProjectIAMBinding) CompareWithID() *string {
	return e.Name
}

func (e *ProjectIAMBinding) Find(c *fi.CloudContext) (*ProjectIAMBinding, error) {
	ctx := context.TODO()

	cloud := c.Cloud.(gce.GCECloud)

	projectID := fi.StringValue(e.Project)
	member := fi.StringValue(e.Member)
	role := fi.StringValue(e.Role)

	klog.V(2).Infof("Checking IAM for project %q", projectID)
	options := &cloudresourcemanager.GetIamPolicyRequest{Options: &cloudresourcemanager.GetPolicyOptions{RequestedPolicyVersion: 3}}
	policy, err := cloud.CloudResourceManager().Projects.GetIamPolicy(projectID, options).Context(ctx).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking IAM for project %s: %w", projectID, err)
	}

	changed := patchCRMPolicy(policy, member, role)
	if changed {
		return nil, nil
	}

	actual := &ProjectIAMBinding{}
	actual.Project = e.Project
	actual.Member = e.Member
	actual.Role = e.Role

	// Ignore "system" fields
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *ProjectIAMBinding) Run(c *fi.CloudContext) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *ProjectIAMBinding) CheckChanges(a, e, changes *ProjectIAMBinding) error {
	if fi.StringValue(e.Project) == "" {
		return fi.RequiredField("Project")
	}
	if fi.StringValue(e.Member) == "" {
		return fi.RequiredField("Member")
	}
	if fi.StringValue(e.Role) == "" {
		return fi.RequiredField("Role")
	}
	return nil
}

func (_ *ProjectIAMBinding) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ProjectIAMBinding) error {
	ctx := context.TODO()

	projectID := fi.StringValue(e.Project)
	member := fi.StringValue(e.Member)
	role := fi.StringValue(e.Role)

	request := &cloudresourcemanager.GetIamPolicyRequest{}
	policy, err := t.Cloud.CloudResourceManager().Projects.GetIamPolicy(projectID, request).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error getting IAM policy for project %s: %w", projectID, err)
	}

	changed := patchCRMPolicy(policy, member, role)

	if !changed {
		klog.Warningf("did not need to change policy (concurrent change?)")
		return nil
	}

	klog.V(2).Infof("updating IAM for project %s", projectID)
	if _, err := t.Cloud.CloudResourceManager().Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("error updating IAM for project %s: %w", projectID, err)
	}

	return nil
}

// terraformProjectIAMBinding is the model for a terraform google_project_iam_binding rule
type terraformProjectIAMBinding struct {
	Project string   `cty:"project"`
	Role    string   `cty:"role"`
	Members []string `cty:"members"`
}

func (_ *ProjectIAMBinding) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ProjectIAMBinding) error {
	tf := &terraformProjectIAMBinding{
		Project: fi.StringValue(e.Project),
		Role:    fi.StringValue(e.Role),
		Members: []string{fi.StringValue(e.Member)},
	}

	return t.RenderResource("google_project_iam_binding", *e.Name, tf)
}

func patchCRMPolicy(policy *cloudresourcemanager.Policy, wantMember string, wantRole string) bool {
	for _, binding := range policy.Bindings {
		if binding.Condition != nil {
			continue
		}
		if binding.Role != wantRole {
			continue
		}
		exists := false
		for _, member := range binding.Members {
			if member == wantMember {
				exists = true
			}
		}
		if exists {
			return false
		}

		if !exists {
			binding.Members = append(binding.Members, wantMember)
			return true
		}
	}

	policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
		Members: []string{wantMember},
		Role:    wantRole,
	})
	return true
}
