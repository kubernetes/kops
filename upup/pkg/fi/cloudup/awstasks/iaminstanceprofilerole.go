package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

type IAMInstanceProfileRole struct {
	InstanceProfile *IAMInstanceProfile
	Role            *IAMRole
}

func (e *IAMInstanceProfileRole) String() string {
	return fi.TaskAsString(e)
}

func (e *IAMInstanceProfileRole) Find(c *fi.Context) (*IAMInstanceProfileRole, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	if e.Role == nil || e.Role.ID == nil {
		glog.V(2).Infof("Role/RoleID not set")
		return nil, nil
	}
	roleID := *e.Role.ID

	request := &iam.GetInstanceProfileInput{InstanceProfileName: e.InstanceProfile.Name}

	response, err := cloud.IAM.GetInstanceProfile(request)
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

		_, err := t.Cloud.IAM.AddRoleToInstanceProfile(request)
		if err != nil {
			return fmt.Errorf("error creating IAMInstanceProfileRole: %v", err)
		}
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}

type terraformIAMInstanceProfile struct {
	Name  *string              `json:"name"`
	Roles []*terraform.Literal `json:"roles"`
}

func (_ *IAMInstanceProfileRole) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMInstanceProfileRole) error {
	tf := &terraformIAMInstanceProfile{
		Name:  e.InstanceProfile.Name,
		Roles: []*terraform.Literal{e.Role.TerraformLink()},
	}

	return t.RenderResource("aws_iam_instance_profile", *e.InstanceProfile.Name, tf)
}
