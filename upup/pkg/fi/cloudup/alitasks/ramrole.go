/*
Copyright 2019 The Kubernetes Authors.

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

package alitasks

import (
	"fmt"

	"github.com/denverdino/aliyungo/ram"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=RAMRole

type RAMRole struct {
	Lifecycle                *fi.Lifecycle
	Name                     *string
	AssumeRolePolicyDocument *string
	RAMRoleId                *string
}

var _ fi.CompareWithID = &RAMRole{}

func (r *RAMRole) CompareWithID() *string {
	return r.Name
}

func (r *RAMRole) Find(c *fi.Context) (*RAMRole, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	roleList, err := cloud.RamClient().ListRoles()
	if err != nil {
		return nil, fmt.Errorf("error listing RamRoles: %v", err)
	}

	// Don't exist RAMrole with specified User.
	if len(roleList.Roles.Role) == 0 {
		return nil, nil
	}

	// The same user's RAM resource name can not be repeated
	for _, role := range roleList.Roles.Role {
		if role.RoleName == fi.StringValue(r.Name) {

			klog.V(2).Infof("found matching RamRole with name: %q", *r.Name)
			actual := &RAMRole{}
			actual.Name = fi.String(role.RoleName)
			actual.RAMRoleId = fi.String(role.RoleId)
			actual.AssumeRolePolicyDocument = fi.String(role.AssumeRolePolicyDocument)

			// Ignore "system" fields
			actual.Lifecycle = r.Lifecycle
			r.RAMRoleId = actual.RAMRoleId
			return actual, nil
		}
	}

	return nil, nil
}

func (r *RAMRole) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(r, c)
}

func (_ *RAMRole) CheckChanges(a, e, changes *RAMRole) error {
	if a == nil {
		if e.AssumeRolePolicyDocument == nil {
			return fi.RequiredField("AssumeRolePolicyDocument")
		}
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *RAMRole) RenderALI(t *aliup.ALIAPITarget, a, e, changes *RAMRole) error {
	if a == nil {
		klog.V(2).Infof("Creating RAMRole with Name:%q", fi.StringValue(e.Name))

		roleRequest := ram.RoleRequest{
			RoleName:                 fi.StringValue(e.Name),
			AssumeRolePolicyDocument: fi.StringValue(e.AssumeRolePolicyDocument),
		}

		roleResponse, err := t.Cloud.RamClient().CreateRole(roleRequest)
		if err != nil {
			return fmt.Errorf("error creating RAMRole: %v", err)
		}

		e.RAMRoleId = fi.String(roleResponse.Role.RoleId)
	}

	return nil
}

type terraformRAMRole struct {
	Name     *string `json:"name,omitempty"`
	Document *string `json:"document,omitempty"`
}

func (_ *RAMRole) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RAMRole) error {
	tf := &terraformRAMRole{
		Name:     e.Name,
		Document: e.AssumeRolePolicyDocument,
	}
	return t.RenderResource("alicloud_ram_role", *e.Name, tf)
}

func (s *RAMRole) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_ram_role", *s.Name, "name")
}
