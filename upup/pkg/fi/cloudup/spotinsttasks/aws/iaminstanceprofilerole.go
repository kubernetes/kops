/*
Copyright 2016 The Kubernetes Authors.

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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=IAMInstanceProfileRole
type IAMInstanceProfileRole struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	InstanceProfile *IAMInstanceProfile
	Role            *IAMRole
}

func (e *IAMInstanceProfileRole) Find(c *fi.Context) (*IAMInstanceProfileRole, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	if e.Role == nil || e.Role.ID == nil {
		glog.V(2).Infof("Role/RoleID not set")
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

func (_ *IAMInstanceProfileRole) Render(t *spotinst.Target, a, e, changes *IAMInstanceProfileRole) error {
	if a == nil {
		request := &iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: e.InstanceProfile.Name,
			RoleName:            e.Role.Name,
		}

		_, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).IAM().AddRoleToInstanceProfile(request)
		if err != nil {
			return fmt.Errorf("error creating IAMInstanceProfileRole: %v", err)
		}
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}
