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

package awstasks

import (
	"fmt"

	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=IAMRolePolicy
type IAMRolePolicy struct {
	ID        *string
	Lifecycle *fi.Lifecycle

	Name *string
	Role *IAMRole

	// The PolicyDocument to create as an inline policy.
	// If the PolicyDocument is empty, the policy will be removed.
	PolicyDocument fi.Resource
}

func (e *IAMRolePolicy) Find(c *fi.Context) (*IAMRolePolicy, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &iam.GetRolePolicyInput{
		RoleName:   e.Role.Name,
		PolicyName: e.Name,
	}

	response, err := cloud.IAM().GetRolePolicy(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "NoSuchEntity" {
			return nil, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error getting role: %v", err)
	}

	p := response
	actual := &IAMRolePolicy{}
	actual.Role = &IAMRole{Name: p.RoleName}
	if aws.StringValue(e.Role.Name) == aws.StringValue(p.RoleName) {
		actual.Role.ID = e.Role.ID
	}
	if p.PolicyDocument != nil {
		// The PolicyDocument is URI encoded (?)
		policy := *p.PolicyDocument
		policy, err = url.QueryUnescape(policy)
		if err != nil {
			return nil, fmt.Errorf("error parsing PolicyDocument for IAMRolePolicy %q: %v", aws.StringValue(e.Name), err)
		}
		actual.PolicyDocument = fi.WrapResource(fi.NewStringResource(policy))
	}

	actual.Name = p.PolicyName

	e.ID = actual.ID

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *IAMRolePolicy) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *IAMRolePolicy) CheckChanges(a, e, changes *IAMRolePolicy) error {
	if a != nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *IAMRolePolicy) ShouldCreate(a, e, changes *IAMRolePolicy) (bool, error) {
	ePolicy, err := e.policyDocumentString()
	if err != nil {
		return false, fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	if a == nil && ePolicy == "" {
		return false, nil
	}
	return true, nil
}

func (_ *IAMRolePolicy) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMRolePolicy) error {
	policy, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	if policy == "" {
		// A deletion

		request := &iam.DeleteRolePolicyInput{}
		request.RoleName = e.Role.Name
		request.PolicyName = e.Name

		klog.V(2).Infof("Deleting role policy %s/%s", aws.StringValue(e.Role.Name), aws.StringValue(e.Name))
		_, err = t.Cloud.IAM().DeleteRolePolicy(request)
		if err != nil {
			if awsup.AWSErrorCode(err) == "NoSuchEntity" {
				// Already deleted
				klog.V(2).Infof("Got NoSuchEntity deleting role policy %s/%s; assuming does not exist", aws.StringValue(e.Role.Name), aws.StringValue(e.Name))
				return nil
			}
			return fmt.Errorf("error deleting IAMRolePolicy: %v", err)
		}
		return nil
	}

	doPut := false

	if a == nil {
		klog.V(2).Infof("Creating IAMRolePolicy")
		doPut = true
	} else if changes != nil {
		if changes.PolicyDocument != nil {
			klog.V(2).Infof("Applying changed role policy to %q:", *e.Name)

			actualPolicy, err := a.policyDocumentString()
			if err != nil {
				return fmt.Errorf("error reading actual policy document: %v", err)
			}

			if actualPolicy == policy {
				klog.Warning("Policies were actually the same")
			} else {
				d := diff.FormatDiff(actualPolicy, policy)
				klog.V(2).Infof("diff: %s", d)
			}

			doPut = true
		}
	}

	if doPut {
		request := &iam.PutRolePolicyInput{}
		request.PolicyDocument = aws.String(policy)
		request.RoleName = e.Role.Name
		request.PolicyName = e.Name

		klog.V(8).Infof("PutRolePolicy RoleName=%s PolicyName=%s: %s", aws.StringValue(e.Role.Name), aws.StringValue(e.Name), policy)

		_, err = t.Cloud.IAM().PutRolePolicy(request)
		if err != nil {
			return fmt.Errorf("error creating/updating IAMRolePolicy: %v", err)
		}
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}

func (e *IAMRolePolicy) policyDocumentString() (string, error) {
	if e.PolicyDocument == nil {
		return "", nil
	}
	return fi.ResourceAsString(e.PolicyDocument)
}

type terraformIAMRolePolicy struct {
	Name           *string            `json:"name"`
	Role           *terraform.Literal `json:"role"`
	PolicyDocument *terraform.Literal `json:"policy"`
}

func (_ *IAMRolePolicy) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMRolePolicy) error {
	policyString, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	if policyString == "" {
		// A deletion; we simply don't render; terraform will observe the removal
		return nil
	}

	policy, err := t.AddFile("aws_iam_role_policy", *e.Name, "policy", e.PolicyDocument)
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	tf := &terraformIAMRolePolicy{
		Name:           e.Name,
		Role:           e.Role.TerraformLink(),
		PolicyDocument: policy,
	}

	return t.RenderResource("aws_iam_role_policy", *e.Name, tf)
}

func (e *IAMRolePolicy) TerraformLink() *terraform.Literal {
	return terraform.LiteralSelfLink("aws_iam_role_policy", *e.Name)
}

type cloudformationIAMRolePolicy struct {
	PolicyName     *string                   `json:"PolicyName"`
	Roles          []*cloudformation.Literal `json:"Roles"`
	PolicyDocument map[string]interface{}    `json:"PolicyDocument"`
}

func (_ *IAMRolePolicy) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *IAMRolePolicy) error {
	policyString, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}
	if policyString == "" {
		// A deletion; we simply don't render; cloudformation will observe the removal
		return nil
	}

	tf := &cloudformationIAMRolePolicy{
		PolicyName: e.Name,
		Roles:      []*cloudformation.Literal{e.Role.CloudformationLink()},
	}

	{
		data := make(map[string]interface{})
		err = json.Unmarshal([]byte(policyString), &data)
		if err != nil {
			return fmt.Errorf("error parsing PolicyDocument: %v", err)
		}

		tf.PolicyDocument = data
	}

	return t.RenderResource("AWS::IAM::Policy", *e.Name, tf)
}

func (e *IAMRolePolicy) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::IAM::Policy", *e.Name)
}
