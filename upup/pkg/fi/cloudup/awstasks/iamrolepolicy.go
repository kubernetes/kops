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
	"hash/fnv"
	"net/url"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type IAMRolePolicy struct {
	ID        *string
	Lifecycle fi.Lifecycle

	Name *string
	Role *IAMRole

	// The PolicyDocument to create as an inline policy.
	// If the PolicyDocument is empty, the policy will be removed.
	PolicyDocument fi.Resource
	// External (non-kops managed) AWS policies to attach to the role
	ExternalPolicies *[]string
	// Managed tracks the use of ExternalPolicies
	Managed bool
}

var _ awsup.AWSTask[IAMRolePolicy] = &IAMRolePolicy{}

func (e *IAMRolePolicy) Find(c *fi.CloudupContext) (*IAMRolePolicy, error) {
	var actual IAMRolePolicy
	cloud := c.T.Cloud.(awsup.AWSCloud)

	// Handle policy overrides
	if e.ExternalPolicies != nil {
		request := &iam.ListAttachedRolePoliciesInput{
			RoleName: e.Role.Name,
		}

		response, err := cloud.IAM().ListAttachedRolePolicies(request)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
				return nil, nil
			}

			return nil, fmt.Errorf("error getting policies for role: %v", err)
		}

		var policies []string
		if response != nil && len(response.AttachedPolicies) > 0 {
			for _, policy := range response.AttachedPolicies {
				policies = append(policies, aws.StringValue(policy.PolicyArn))
			}
		}
		sort.Strings(policies)

		actual.ID = e.ID
		actual.Name = e.Name
		actual.Lifecycle = e.Lifecycle
		actual.Role = e.Role
		actual.Managed = true
		actual.ExternalPolicies = &policies

		return &actual, nil
	}

	request := &iam.GetRolePolicyInput{
		RoleName:   e.Role.Name,
		PolicyName: e.Name,
	}

	response, err := cloud.IAM().GetRolePolicy(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return nil, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error getting role: %v", err)
	}

	p := response
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

		// Reformat the PolicyDocument by unmarshaling and re-marshaling to JSON.
		// This will make it possible to compare it when using CloudFormation.
		var jsonData interface{}
		err = json.Unmarshal([]byte(policy), &jsonData)
		if err != nil {
			return nil, fmt.Errorf("error parsing cloudformation policy document from JSON: %v", err)
		}
		jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error converting cloudformation policy document to JSON: %v", err)
		}
		actual.PolicyDocument = fi.NewStringResource(string(jsonBytes))
	}

	actual.Name = p.PolicyName

	e.ID = actual.ID

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return &actual, nil
}

func (e *IAMRolePolicy) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
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

	if a == nil && ePolicy == "" && e.ExternalPolicies == nil {
		return false, nil
	}

	return true, nil
}

func (_ *IAMRolePolicy) RenderAWS(ctx *fi.CloudupContext, t *awsup.AWSAPITarget, a, e, changes *IAMRolePolicy) error {
	policy, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	// Handles the full lifecycle of Policy Overrides
	if e.Managed {
		// Attach policies that are not already attached
	AttachPolicies:
		for _, policy := range *e.ExternalPolicies {
			for _, cloudPolicy := range *a.ExternalPolicies {
				if cloudPolicy == policy {
					continue AttachPolicies
				}
			}

			request := &iam.AttachRolePolicyInput{
				RoleName:  e.Role.Name,
				PolicyArn: s(policy),
			}

			_, err = t.Cloud.IAM().AttachRolePolicy(request)
			if err != nil {
				return fmt.Errorf("error attaching IAMRolePolicy: %v", err)
			}
		}

		// Clean up unused cloud policies
	CheckPolicies:
		for _, cloudPolicy := range *a.ExternalPolicies {
			for _, policy := range *e.ExternalPolicies {
				if policy == cloudPolicy {
					continue CheckPolicies
				}
			}

			klog.V(2).Infof("Detaching unused IAMRolePolicy %s/%s", aws.StringValue(e.Role.Name), cloudPolicy)

			// Detach policy
			request := &iam.DetachRolePolicyInput{
				RoleName:  e.Role.Name,
				PolicyArn: s(cloudPolicy),
			}

			_, err := t.Cloud.IAM().DetachRolePolicy(request)
			if err != nil {
				klog.V(2).Infof("Unable to detach IAMRolePolicy %s/%s", aws.StringValue(e.Role.Name), cloudPolicy)
				return err
			}
		}

		return nil
	}

	if policy == "" {
		// A deletion

		request := &iam.DeleteRolePolicyInput{}
		request.RoleName = e.Role.Name
		request.PolicyName = e.Name

		klog.V(2).Infof("Deleting role policy %s/%s", aws.StringValue(e.Role.Name), aws.StringValue(e.Name))
		_, err = t.Cloud.IAM().DeleteRolePolicy(request)
		if err != nil {
			if awsup.AWSErrorCode(err) == iam.ErrCodeNoSuchEntityException {
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
			klog.V(2).Infof("PutRolePolicy RoleName=%s PolicyName=%s: %s", aws.StringValue(e.Role.Name), aws.StringValue(e.Name), policy)
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

	policy, err := fi.ResourceAsString(e.PolicyDocument)
	if err != nil {
		return "", err
	}
	policySize := len(strings.Join(strings.Fields(policy), ""))
	if policySize > 10240 {
		return "", fmt.Errorf("policy size was %d. Policy cannot exceed 10240 bytes", policySize)
	}
	return policy, err
}

type terraformIAMRolePolicy struct {
	Name           *string                  `cty:"name"`
	Role           *terraformWriter.Literal `cty:"role"`
	PolicyDocument *terraformWriter.Literal `cty:"policy"`
	PolicyArn      *string                  `cty:"policy_arn"`
}

func (_ *IAMRolePolicy) RenderTerraform(ctx *fi.CloudupContext, t *terraform.TerraformTarget, a, e, changes *IAMRolePolicy) error {
	if e.ExternalPolicies != nil && len(*e.ExternalPolicies) > 0 {
		for _, policy := range *e.ExternalPolicies {
			// create a hash of the arn
			h := fnv.New32a()
			h.Write([]byte(policy))

			name := fmt.Sprintf("%s-%d", *e.Name, h.Sum32())

			tf := &terraformIAMRolePolicy{
				Role:      e.Role.TerraformLink(),
				PolicyArn: s(policy),
			}

			err := t.RenderResource("aws_iam_role_policy_attachment", name, tf)
			if err != nil {
				return fmt.Errorf("error rendering RolePolicyAttachment: %v", err)
			}
		}
	}

	policyString, err := e.policyDocumentString()
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	if policyString == "" {
		// A deletion; we simply don't render; terraform will observe the removal
		return nil
	}

	policy, err := t.AddFileResource("aws_iam_role_policy", *e.Name, "policy", e.PolicyDocument, false)
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

func (e *IAMRolePolicy) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralSelfLink("aws_iam_role_policy", *e.Name)
}
