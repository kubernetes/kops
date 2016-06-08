package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=IAMInstanceProfile
type IAMInstanceProfile struct {
	Name *string
	ID   *string
}

var _ fi.CompareWithID = &IAMInstanceProfile{}

func (e *IAMInstanceProfile) CompareWithID() *string {
	return e.Name
}

func (e *IAMInstanceProfile) Find(c *fi.Context) (*IAMInstanceProfile, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &iam.GetInstanceProfileInput{InstanceProfileName: e.Name}

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
	actual := &IAMInstanceProfile{
		ID:   ip.InstanceProfileId,
		Name: ip.InstanceProfileName,
	}

	e.ID = actual.ID
	e.Name = actual.Name

	return actual, nil
}

func (e *IAMInstanceProfile) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *IAMInstanceProfile) CheckChanges(a, e, changes *IAMInstanceProfile) error {
	if a != nil {
		if fi.StringValue(e.Name) == "" {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *IAMInstanceProfile) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMInstanceProfile) error {
	if a == nil {
		glog.V(2).Infof("Creating IAMInstanceProfile with Name:%q", *e.Name)

		request := &iam.CreateInstanceProfileInput{
			InstanceProfileName: e.Name,
		}

		response, err := t.Cloud.IAM.CreateInstanceProfile(request)
		if err != nil {
			return fmt.Errorf("error creating IAMInstanceProfile: %v", err)
		}

		e.ID = response.InstanceProfile.InstanceProfileId
		e.Name = response.InstanceProfile.InstanceProfileName
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}

func (_ *IAMInstanceProfile) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMInstanceProfile) error {
	// Done on IAMInstanceProfileRole
	return nil
}

func (e *IAMInstanceProfile) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_iam_instance_profile", *e.Name, "id")
}
