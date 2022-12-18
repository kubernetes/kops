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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type IAMInstanceProfileRole struct {
	Name      *string
	Lifecycle fi.Lifecycle

	InstanceProfile *IAMInstanceProfile
	Role            *IAMRole
}

var _ awsup.AWSTask[IAMInstanceProfileRole] = &IAMInstanceProfileRole{}

func (e *IAMInstanceProfileRole) Find(c *fi.CloudupContext) (*IAMInstanceProfileRole, error) {
	cloud := c.T.Cloud.(awsup.AWSCloud)

	if e.Role == nil || e.Role.ID == nil {
		klog.V(2).Infof("Role/RoleID not set")
		return nil, nil
	}
	roleID := *e.Role.ID

	request := &iam.GetInstanceProfileInput{InstanceProfileName: e.InstanceProfile.Name}

	response, err := cloud.IAM().GetInstanceProfile(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
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

func (e *IAMInstanceProfileRole) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
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

func (_ *IAMInstanceProfileRole) RenderAWS(ctx *fi.CloudupContext, t *awsup.AWSAPITarget, a, e, changes *IAMInstanceProfileRole) error {
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
	return nil
}

type terraformIAMInstanceProfile struct {
	Name *string                  `cty:"name"`
	Role *terraformWriter.Literal `cty:"role"`
	Tags map[string]string        `cty:"tags"`
}

func (_ *IAMInstanceProfileRole) RenderTerraform(ctx *fi.CloudupContext, t *terraform.TerraformTarget, a, e, changes *IAMInstanceProfileRole) error {
	tf := &terraformIAMInstanceProfile{
		Name: e.InstanceProfile.Name,
		Role: e.Role.TerraformLink(),
		Tags: e.InstanceProfile.Tags,
	}

	return t.RenderResource("aws_iam_instance_profile", *e.InstanceProfile.Name, tf)
}
