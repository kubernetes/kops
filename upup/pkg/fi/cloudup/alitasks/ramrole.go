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
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ram"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=RAMRole

type RAMRole struct {
	ID                       *string
	Lifecycle                *fi.Lifecycle
	Name                     *string
	AssumeRolePolicyDocument *string
}

var _ fi.CompareWithID = &RAMRole{}

func (r *RAMRole) CompareWithID() *string {
	return r.ID
}

func compactPolicy(s string) string {
	removedStrings := []string{"\n", "<br>", " ", "\r\n"}
	for _, each := range removedStrings {
		s = strings.Replace(s, each, "", -1)
	}
	return s
}

func (r *RAMRole) Find(c *fi.Context) (*RAMRole, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	request := ram.RoleQueryRequest{
		RoleName: fi.StringValue(r.Name),
	}

	roleResp, err := cloud.RamClient().GetRole(request)
	if err != nil {
		if e, ok := err.(*common.Error); ok && e.StatusCode == 404 {
			klog.V(2).Infof("no RamRole with name: %q", *r.Name)
			return nil, nil
		}
		return nil, fmt.Errorf("error get RamRole %q: %v", *r.Name, err)
	}

	role := roleResp.Role
	if role.RoleId == "" {
		klog.V(2).Infof("no RamRole with name: %q", *r.Name)
		return nil, nil
	}

	klog.V(2).Infof("found matching RamRole with name: %q", *r.Name)
	actual := &RAMRole{
		Name:                     fi.String(role.RoleName),
		ID:                       fi.String(role.RoleId),
		AssumeRolePolicyDocument: fi.String(compactPolicy(role.AssumeRolePolicyDocument)),
	}

	// Ignore "system" fields
	actual.Lifecycle = r.Lifecycle
	r.ID = actual.ID

	return actual, nil
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

		e.ID = fi.String(roleResponse.Role.RoleId)
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
