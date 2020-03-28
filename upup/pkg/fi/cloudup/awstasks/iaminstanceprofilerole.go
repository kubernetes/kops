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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=IAMInstanceProfileRole
type IAMInstanceProfileRole struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	InstanceProfile *IAMInstanceProfile
	Role            *IAMRole
}

func (e *IAMInstanceProfileRole) Find(c *fi.Context) (*IAMInstanceProfileRole, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.Role == nil || e.Role.ID == nil {
		klog.V(2).Infof("Role/RoleID not set")
		return nil, nil
	}
	roleID := *e.Role.ID

	request := &iam.GetInstanceProfileInput{InstanceProfileName: e.InstanceProfile.Name}

	response, err := cloud.IAM().GetInstanceProfile(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "NoSuchEntity" {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error getting IAMInstanceProfile: %v", err)
	}

	ip := response.InstanceProfile
	for _, role := range ip.Roles {
		if aws.StringValue(role.RoleId) != roleID {
			continue
		}
		actual := &IAMInstanceProfileRole{}
		actual.InstanceProfile = &IAMInstanceProfile{ID: ip.InstanceProfileId, Name: ip.InstanceProfileName}
		actual.Role = &IAMRole{ID: role.RoleId, Name: role.RoleName}

		// Prevent spurious changes
		actual.Name = e.Name
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}
	return nil, nil
}

func (e *IAMInstanceProfileRole) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *IAMInstanceProfileRole) CheckChanges(a, e, changes *IAMInstanceProfileRole) error {
	if a != nil {
		if e.Role == nil {
			return fi.RequiredField("Role")
		}
		if e.InstanceProfile == nil {
			return fi.RequiredField("InstanceProfile")
		}
	}
	return nil
}

func (_ *IAMInstanceProfileRole) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMInstanceProfileRole) error {
	if a == nil {
		request := &iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: e.InstanceProfile.Name,
			RoleName:            e.Role.Name,
		}

		_, err := t.Cloud.IAM().AddRoleToInstanceProfile(request)
		if err != nil {
			return fmt.Errorf("error creating IAMInstanceProfileRole: %v", err)
		}
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}

type terraformIAMInstanceProfile struct {
	Name *string            `json:"name" cty:"name"`
	Role *terraform.Literal `json:"role" cty:"role"`
}

func (_ *IAMInstanceProfileRole) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMInstanceProfileRole) error {
	tf := &terraformIAMInstanceProfile{
		Name: e.InstanceProfile.Name,
		Role: e.Role.TerraformLink(),
	}

	return t.RenderResource("aws_iam_instance_profile", *e.InstanceProfile.Name, tf)
}

type cloudformationIAMInstanceProfile struct {
	//Path  *string              `json:"name"`
	Roles []*cloudformation.Literal `json:"Roles"`
}

func (_ *IAMInstanceProfileRole) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *IAMInstanceProfileRole) error {
	cf := &cloudformationIAMInstanceProfile{
		//Path:  e.InstanceProfile.Name,
		Roles: []*cloudformation.Literal{e.Role.CloudformationLink()},
	}

	return t.RenderResource("AWS::IAM::InstanceProfile", *e.InstanceProfile.Name, cf)
}
