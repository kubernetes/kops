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
	"net/url"

	"github.com/denverdino/aliyungo/common"
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
	RamRole        *RAMRole
	PolicyType     *string
	PolicyDocument fi.Resource
}

var _ fi.CompareWithID = &RAMPolicy{}

func (r *RAMPolicy) CompareWithID() *string {
	return r.Name
}

func (r *RAMPolicy) Find(c *fi.Context) (*RAMPolicy, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	policyRequest := ram.PolicyRequest{
		PolicyName: fi.StringValue(r.Name),
		PolicyType: ram.Type(fi.StringValue(r.PolicyType)),
	}
	policyResp, err := cloud.RamClient().GetPolicy(policyRequest)
	if err != nil {
		if e, ok := err.(*common.Error); ok && e.StatusCode == 404 {
			klog.V(2).Infof("no RamPolicy with name: %q", *r.Name)
			return nil, nil
		}
		return nil, fmt.Errorf("error get RamPolicy %s: %v", *r.Name, err)
	}

	klog.V(2).Infof("found matching RamPolicy with name: %q", *r.Name)
	policy := policyResp.Policy

	defaultPolicy, err := url.QueryUnescape(policyResp.DefaultPolicyVersion.PolicyDocument)
	if err != nil {
		return nil, fmt.Errorf("error parsing PolicyDocument for RAMPolicy %q: %v", fi.StringValue(r.Name), err)
	}

	actual := &RAMPolicy{
		Name:           fi.String(policy.PolicyName),
		PolicyType:     fi.String(string(policy.PolicyType)),
		PolicyDocument: fi.WrapResource(fi.NewStringResource(defaultPolicy)),
	}

	// Avoid spurious changes
	actual.RamRole = r.RamRole
	actual.Lifecycle = r.Lifecycle

	return actual, nil
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
	policy, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	policyRequest := ram.PolicyRequest{}

	if a == nil {
		klog.V(2).Infof("Creating RAMPolicy with Name:%q", fi.StringValue(e.Name))

		policyRequest = ram.PolicyRequest{
			PolicyName:     fi.StringValue(e.Name),
			PolicyDocument: policy,
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

func (r *RAMPolicy) policyDocumentString() (string, error) {
	if r.PolicyDocument == nil {
		return "", nil
	}
	return fi.ResourceAsString(r.PolicyDocument)
}

type terraformRAMPolicy struct {
	Name     *string `json:"name,omitempty" cty:"name"`
	Document *string `json:"document,omitempty" cty:"document"`
}

type terraformRAMPolicyAttach struct {
	PolicyName *terraform.Literal `json:"policy_name,omitempty" cty:"policy_name"`
	PolicyType *string            `json:"policy_type,omitempty" cty:"policy_type"`
	RoleName   *terraform.Literal `json:"role_name,omitempty" cty:"role_name"`
}

func (_ *RAMPolicy) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RAMPolicy) error {
	policyString, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	tf := &terraformRAMPolicy{
		Name:     e.Name,
		Document: fi.String(policyString),
	}
	err = t.RenderResource("alicloud_ram_policy", *e.Name, tf)
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
