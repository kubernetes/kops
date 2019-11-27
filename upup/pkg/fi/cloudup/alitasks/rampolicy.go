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

//go:generate fitask -type=RAMPolicy

type RAMPolicy struct {
	Lifecycle      *fi.Lifecycle
	Name           *string
	PolicyDocument *string
	RamRole        *RAMRole
	PolicyType     *string
}

var _ fi.CompareWithID = &RAMPolicy{}

func (r *RAMPolicy) CompareWithID() *string {
	return r.Name
}

func (r *RAMPolicy) Find(c *fi.Context) (*RAMPolicy, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	policyQueryRequest := ram.PolicyQueryRequest{
		PolicyType: ram.Type(fi.StringValue(r.PolicyType)),
	}
	policyList, err := cloud.RamClient().ListPolicies(policyQueryRequest)

	if err != nil {
		return nil, fmt.Errorf("error listing RamPolicy: %v", err)
	}

	if len(policyList.Policies.Policy) == 0 {
		return nil, nil
	}

	for _, policy := range policyList.Policies.Policy {
		if policy.PolicyName == fi.StringValue(r.Name) {

			klog.V(2).Infof("found matching RamPolicy with name: %q", *r.Name)
			actual := &RAMPolicy{}
			actual.Name = fi.String(policy.PolicyName)
			actual.PolicyType = fi.String(string(policy.PolicyType))

			// Ignore "system" fields
			actual.Lifecycle = r.Lifecycle
			return actual, nil
		}
	}

	return nil, nil
}

func (r *RAMPolicy) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(r, c)
}

func (_ *RAMPolicy) CheckChanges(a, e, changes *RAMPolicy) error {

	if e.PolicyDocument == nil {
		return fi.RequiredField("PolicyDocument")
	}
	if e.Name == nil {
		return fi.RequiredField("Name")
	}

	return nil
}

func (_ *RAMPolicy) RenderALI(t *aliup.ALIAPITarget, a, e, changes *RAMPolicy) error {

	policyRequest := ram.PolicyRequest{}

	if a == nil {
		klog.V(2).Infof("Creating RAMPolicy with Name:%q", fi.StringValue(e.Name))

		policyRequest = ram.PolicyRequest{
			PolicyName:     fi.StringValue(e.Name),
			PolicyDocument: fi.StringValue(e.PolicyDocument),
			PolicyType:     ram.Type(fi.StringValue(e.PolicyType)),
		}

		_, err := t.Cloud.RamClient().CreatePolicy(policyRequest)
		if err != nil {
			return fmt.Errorf("error creating RAMPolicy: %v", err)
		}

		attachPolicyRequest := ram.AttachPolicyToRoleRequest{
			PolicyRequest: policyRequest,
			RoleName:      fi.StringValue(e.RamRole.Name),
		}

		_, err = t.Cloud.RamClient().AttachPolicyToRole(attachPolicyRequest)
		if err != nil {
			return fmt.Errorf("error attaching RAMPolicy to RAMRole: %v", err)
		}
		return nil
	}

	return nil

}

type terraformRAMPolicy struct {
	Name     *string `json:"name,omitempty"`
	Document *string `json:"document,omitempty"`
}

type terraformRAMPolicyAttach struct {
	PolicyName *terraform.Literal `json:"policy_name,omitempty"`
	PolicyType *string            `json:"policy_type,omitempty"`
	RoleName   *terraform.Literal `json:"role_name,omitempty"`
}

func (_ *RAMPolicy) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RAMPolicy) error {
	tf := &terraformRAMPolicy{
		Name:     e.Name,
		Document: e.PolicyDocument,
	}
	err := t.RenderResource("alicloud_ram_policy", *e.Name, tf)
	if err != nil {
		return err
	}

	policyType := "Custom"
	tfAttach := &terraformRAMPolicyAttach{
		PolicyName: e.TerraformLink(),
		RoleName:   e.RamRole.TerraformLink(),
		PolicyType: &policyType,
	}
	err = t.RenderResource("alicloud_ram_role_policy_attachment", *e.Name, tfAttach)
	return err
}

func (s *RAMPolicy) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_ram_policy", *s.Name, "id")
}
