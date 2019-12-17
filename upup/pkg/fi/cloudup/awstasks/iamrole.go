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
	"reflect"

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

//go:generate fitask -type=IAMRole
type IAMRole struct {
	ID        *string
	Lifecycle *fi.Lifecycle

	Name               *string
	RolePolicyDocument *fi.ResourceHolder // "inline" IAM policy

	// ExportWithId will expose the name & ARN for reuse as part of a larger system.  Only supported by terraform currently.
	ExportWithID *string
}

var _ fi.CompareWithID = &IAMRole{}

func (e *IAMRole) CompareWithID() *string {
	return e.ID
}

func (e *IAMRole) Find(c *fi.Context) (*IAMRole, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &iam.GetRoleInput{RoleName: e.Name}

	response, err := cloud.IAM().GetRole(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "NoSuchEntity" {
			return nil, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error getting role: %v", err)
	}

	r := response.Role
	actual := &IAMRole{}
	actual.ID = r.RoleId
	actual.Name = r.RoleName
	if r.AssumeRolePolicyDocument != nil {
		// The AssumeRolePolicyDocument is URI encoded (?)
		actualPolicy := *r.AssumeRolePolicyDocument
		actualPolicy, err = url.QueryUnescape(actualPolicy)
		if err != nil {
			return nil, fmt.Errorf("error parsing AssumeRolePolicyDocument for IAMRole %s: %v", *e.Name, err)
		}

		// The RolePolicyDocument is reformatted by AWS
		// We parse both as JSON; if the json forms are equal we pretend the actual value is the expected value
		if e.RolePolicyDocument != nil {
			expectedPolicy, err := e.RolePolicyDocument.AsString()
			if err != nil {
				return nil, fmt.Errorf("error reading expected RolePolicyDocument for IAMRole %q: %v", aws.StringValue(e.Name), err)
			}
			expectedJson := make(map[string]interface{})
			err = json.Unmarshal([]byte(expectedPolicy), &expectedJson)
			if err != nil {
				return nil, fmt.Errorf("error parsing expected RolePolicyDocument for IAMRole %q: %v", aws.StringValue(e.Name), err)
			}
			actualJson := make(map[string]interface{})
			err = json.Unmarshal([]byte(actualPolicy), &actualJson)
			if err != nil {
				return nil, fmt.Errorf("error parsing actual RolePolicyDocument for IAMRole %q: %v", aws.StringValue(e.Name), err)
			}

			if reflect.DeepEqual(actualJson, expectedJson) {
				klog.V(2).Infof("actual RolePolicyDocument was json-equal to expected; returning expected value")
				actualPolicy = expectedPolicy
			}
		}

		actual.RolePolicyDocument = fi.WrapResource(fi.NewStringResource(actualPolicy))
	}

	klog.V(2).Infof("found matching IAMRole %q", aws.StringValue(actual.ID))
	e.ID = actual.ID

	// Avoid spurious changes
	actual.ExportWithID = e.ExportWithID
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *IAMRole) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *IAMRole) CheckChanges(a, e, changes *IAMRole) error {
	if a != nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.Name == nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *IAMRole) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMRole) error {
	policy, err := e.RolePolicyDocument.AsString()
	if err != nil {
		return fmt.Errorf("error rendering RolePolicyDocument: %v", err)
	}

	if a == nil {
		klog.V(2).Infof("Creating IAMRole with Name:%q", *e.Name)

		request := &iam.CreateRoleInput{}
		request.AssumeRolePolicyDocument = aws.String(policy)
		request.RoleName = e.Name

		response, err := t.Cloud.IAM().CreateRole(request)
		if err != nil {
			return fmt.Errorf("error creating IAMRole: %v", err)
		}

		e.ID = response.Role.RoleId
	} else {
		if changes.RolePolicyDocument != nil {
			klog.V(2).Infof("Updating IAMRole AssumeRolePolicy %q", *e.Name)

			var err error

			actualPolicy := ""
			if a.RolePolicyDocument != nil {
				actualPolicy, err = a.RolePolicyDocument.AsString()
				if err != nil {
					return fmt.Errorf("error reading actual policy document: %v", err)
				}
			}

			if actualPolicy == policy {
				klog.Warning("Policies were actually the same")
			} else {
				d := diff.FormatDiff(actualPolicy, policy)
				klog.V(2).Infof("diff: %s", d)
			}

			request := &iam.UpdateAssumeRolePolicyInput{}
			request.PolicyDocument = aws.String(policy)
			request.RoleName = e.Name

			_, err = t.Cloud.IAM().UpdateAssumeRolePolicy(request)
			if err != nil {
				return fmt.Errorf("error updating IAMRole: %v", err)
			}
		}
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}

type terraformIAMRole struct {
	Name             *string            `json:"name"`
	AssumeRolePolicy *terraform.Literal `json:"assume_role_policy"`
}

func (_ *IAMRole) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMRole) error {
	policy, err := t.AddFile("aws_iam_role", *e.Name, "policy", e.RolePolicyDocument)
	if err != nil {
		return fmt.Errorf("error rendering RolePolicyDocument: %v", err)
	}

	tf := &terraformIAMRole{
		Name:             e.Name,
		AssumeRolePolicy: policy,
	}

	if fi.StringValue(e.ExportWithID) != "" {
		t.AddOutputVariable(*e.ExportWithID+"_role_arn", terraform.LiteralProperty("aws_iam_role", *e.Name, "arn"))
		t.AddOutputVariable(*e.ExportWithID+"_role_name", e.TerraformLink())
	}

	return t.RenderResource("aws_iam_role", *e.Name, tf)
}

func (e *IAMRole) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_iam_role", *e.Name, "name")
}

type cloudformationIAMRole struct {
	RoleName                 *string `json:"RoleName"`
	AssumeRolePolicyDocument map[string]interface{}
}

func (_ *IAMRole) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *IAMRole) error {
	jsonString, err := e.RolePolicyDocument.AsBytes()
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(jsonString, &data)
	if err != nil {
		return fmt.Errorf("error parsing RolePolicyDocument: %v", err)
	}

	cf := &cloudformationIAMRole{
		RoleName:                 e.Name,
		AssumeRolePolicyDocument: data,
	}

	return t.RenderResource("AWS::IAM::Role", *e.Name, cf)
}

func (e *IAMRole) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::IAM::Role", *e.Name)
}
