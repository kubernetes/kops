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
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// CloudTagServiceAccountName is a tag that defines the service account's name
const CloudTagServiceAccountName = "service-account.kops.k8s.io/name"

// CloudTagServiceAccountNamespace is a tag that defines the service account's namespace
const CloudTagServiceAccountNamespace = "service-account.kops.k8s.io/namespace"

// +kops:fitask
type IAMRole struct {
	ID        *string
	Lifecycle fi.Lifecycle

	Name                *string
	RolePolicyDocument  fi.Resource // "inline" IAM policy
	PermissionsBoundary *string

	Tags map[string]string

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
		if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
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
	if r.PermissionsBoundary != nil {
		actual.PermissionsBoundary = r.PermissionsBoundary.PermissionsBoundaryArn
	}
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
			expectedPolicy, err := fi.ResourceAsString(e.RolePolicyDocument)
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

		actual.RolePolicyDocument = fi.NewStringResource(actualPolicy)
	}
	actual.Tags = mapIAMTagsToMap(r.Tags)

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

func (s *IAMRole) ShouldCreate(a, e, changes *IAMRole) (bool, error) {
	if len(*e.Name) > 64 {
		return false, fmt.Errorf("role name length must be equal to 64 or less: %q", *e.Name)
	}
	return true, nil
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
	if e.RolePolicyDocument == nil {
		klog.V(2).Infof("Deleting IAM role %q", a.Name)

		var attachedPolicies []*iam.AttachedPolicy
		var policyNames []string

		// List Inline policies
		{
			request := &iam.ListRolePoliciesInput{
				RoleName: a.Name,
			}
			err := t.Cloud.IAM().ListRolePoliciesPages(request, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
				for _, policy := range page.PolicyNames {
					policyNames = append(policyNames, aws.StringValue(policy))
				}
				return true
			})
			if err != nil {
				if awsup.AWSErrorCode(err) == iam.ErrCodeNoSuchEntityException {
					klog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy; will treat as already-deleted")
					return nil
				}

				return fmt.Errorf("error listing IAM role policies: %v", err)
			}
		}

		// List Attached Policies
		{
			request := &iam.ListAttachedRolePoliciesInput{
				RoleName: a.Name,
			}
			err := t.Cloud.IAM().ListAttachedRolePoliciesPages(request, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
				attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)
				return true
			})
			if err != nil {
				if awsup.AWSErrorCode(err) == iam.ErrCodeNoSuchEntityException {
					klog.V(2).Infof("Got NoSuchEntity describing IAM RolePolicy; will treat as already-detached")
					return nil
				}

				return fmt.Errorf("error listing IAM role policies for %v", err)
			}
		}

		// Delete inline policies
		for _, policyName := range policyNames {
			klog.V(2).Infof("Deleting IAM role policy %q", policyName)
			request := &iam.DeleteRolePolicyInput{
				RoleName:   a.Name,
				PolicyName: aws.String(policyName),
			}
			_, err := t.Cloud.IAM().DeleteRolePolicy(request)
			if err != nil {
				return fmt.Errorf("error deleting IAM role policy %q: %v", policyName, err)
			}
		}

		// Detach Managed Policies
		for _, policy := range attachedPolicies {
			klog.V(2).Infof("Detaching IAM role policy %q", policy)
			request := &iam.DetachRolePolicyInput{
				RoleName:  a.Name,
				PolicyArn: policy.PolicyArn,
			}
			_, err := t.Cloud.IAM().DetachRolePolicy(request)
			if err != nil {
				return fmt.Errorf("error detaching IAM role policy %q: %v", *policy.PolicyArn, err)
			}
		}

		request := &iam.DeleteRoleInput{
			RoleName: a.Name,
		}
		if _, err := t.Cloud.IAM().DeleteRole(request); err != nil {
			return fmt.Errorf("error deleting IAM role: %v", err)
		}
		return nil
	}

	policy, err := fi.ResourceAsString(e.RolePolicyDocument)
	if err != nil {
		return fmt.Errorf("error rendering RolePolicyDocument: %v", err)
	}

	if a == nil {
		klog.V(2).Infof("Creating IAMRole with Name:%q", *e.Name)

		request := &iam.CreateRoleInput{}
		request.AssumeRolePolicyDocument = aws.String(policy)
		request.RoleName = e.Name
		request.Tags = mapToIAMTags(e.Tags)

		if e.PermissionsBoundary != nil {
			request.PermissionsBoundary = e.PermissionsBoundary
		}

		response, err := t.Cloud.IAM().CreateRole(request)
		if err != nil {
			klog.V(2).Infof("IAMRole policy: %s", policy)
			return fmt.Errorf("error creating IAMRole: %v", err)
		}

		e.ID = response.Role.RoleId
	} else {
		if changes.RolePolicyDocument != nil {
			klog.V(2).Infof("Updating IAMRole AssumeRolePolicy %q", *e.Name)

			var err error

			actualPolicy := ""
			if a.RolePolicyDocument != nil {
				actualPolicy, err = fi.ResourceAsString(a.RolePolicyDocument)
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
		if changes.PermissionsBoundary != nil {
			klog.V(2).Infof("Updating IAMRole PermissionsBoundary %q", *e.Name)

			request := &iam.PutRolePermissionsBoundaryInput{}
			request.RoleName = e.Name
			request.PermissionsBoundary = e.PermissionsBoundary

			if _, err := t.Cloud.IAM().PutRolePermissionsBoundary(request); err != nil {
				return fmt.Errorf("error updating IAMRole: %v", err)
			}
		} else if a.PermissionsBoundary != nil && e.PermissionsBoundary == nil {
			request := &iam.DeleteRolePermissionsBoundaryInput{}
			request.RoleName = e.Name

			if _, err := t.Cloud.IAM().DeleteRolePermissionsBoundary(request); err != nil {
				return fmt.Errorf("error updating IAMRole: %v", err)
			}
		}
		if changes.Tags != nil {
			if len(a.Tags) > 0 {
				existingTagKeys := make([]*string, 0)
				for k := range a.Tags {
					existingTagKeys = append(existingTagKeys, &k)
				}
				untagRequest := &iam.UntagRoleInput{
					RoleName: e.Name,
					TagKeys:  existingTagKeys,
				}
				_, err = t.Cloud.IAM().UntagRole(untagRequest)
				if err != nil {
					return fmt.Errorf("error untagging IAMRole: %v", err)
				}
			}
			if len(e.Tags) > 0 {
				tagRequest := &iam.TagRoleInput{
					RoleName: e.Name,
					Tags:     mapToIAMTags(e.Tags),
				}
				_, err = t.Cloud.IAM().TagRole(tagRequest)
				if err != nil {
					return fmt.Errorf("error tagging IAMRole: %v", err)
				}
			}
		}
	}
	return nil
}

type terraformIAMRole struct {
	Name                *string                  `cty:"name"`
	AssumeRolePolicy    *terraformWriter.Literal `cty:"assume_role_policy"`
	PermissionsBoundary *string                  `cty:"permissions_boundary"`
	Tags                map[string]string        `cty:"tags"`
}

func (_ *IAMRole) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMRole) error {
	policy, err := t.AddFileResource("aws_iam_role", *e.Name, "policy", e.RolePolicyDocument, false)
	if err != nil {
		return fmt.Errorf("error rendering RolePolicyDocument: %v", err)
	}

	tf := &terraformIAMRole{
		Name:             e.Name,
		AssumeRolePolicy: policy,
		Tags:             e.Tags,
	}

	if e.PermissionsBoundary != nil {
		tf.PermissionsBoundary = e.PermissionsBoundary
	}

	if fi.StringValue(e.ExportWithID) != "" {
		t.AddOutputVariable(*e.ExportWithID+"_role_arn", terraformWriter.LiteralProperty("aws_iam_role", *e.Name, "arn"))
		t.AddOutputVariable(*e.ExportWithID+"_role_name", e.TerraformLink())
	}

	return t.RenderResource("aws_iam_role", *e.Name, tf)
}

func (e *IAMRole) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("aws_iam_role", *e.Name, "name")
}

type cloudformationIAMRole struct {
	RoleName                 *string `json:"RoleName"`
	AssumeRolePolicyDocument map[string]interface{}
	PermissionsBoundary      *string             `json:"PermissionsBoundary,omitempty"`
	Tags                     []cloudformationTag `json:"Tags,omitempty"`
}

func (_ *IAMRole) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *IAMRole) error {
	jsonString, err := fi.ResourceAsBytes(e.RolePolicyDocument)
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
		Tags:                     buildCloudformationTags(e.Tags),
	}

	if e.PermissionsBoundary != nil {
		cf.PermissionsBoundary = e.PermissionsBoundary
	}

	return t.RenderResource("AWS::IAM::Role", *e.Name, cf)
}

func (e *IAMRole) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::IAM::Role", *e.Name)
}
