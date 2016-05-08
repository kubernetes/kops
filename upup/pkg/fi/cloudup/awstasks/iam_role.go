package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"net/url"
)

type IAMRole struct {
	ID                 *string
	Name               *string
	RolePolicyDocument fi.Resource // "inline" IAM policy
}

var _ fi.CompareWithID = &InternetGateway{}

func (e *IAMRole) CompareWithID() *string {
	return e.Name
}

func (e *IAMRole) String() string {
	return fi.TaskAsString(e)
}

func (e *IAMRole) Find(c *fi.Context) (*IAMRole, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &iam.GetRoleInput{RoleName: e.Name}

	response, err := cloud.IAM.GetRole(request)
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
		policy := *r.AssumeRolePolicyDocument
		policy, err = url.QueryUnescape(policy)
		if err != nil {
			return nil, fmt.Errorf("error parsing AssumeRolePolicyDocument for IAMRole %q: %v", e.Name, err)
		}
		actual.RolePolicyDocument = fi.NewStringResource(policy)
	}

	glog.V(2).Infof("found matching IAMRole %q", *actual.ID)
	e.ID = actual.ID

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
	policy, err := fi.ResourceAsString(e.RolePolicyDocument)
	if err != nil {
		return fmt.Errorf("error rendering PolicyDocument: %v", err)
	}

	if a == nil {
		glog.V(2).Infof("Creating IAMRole with Name:%q", *e.Name)

		request := &iam.CreateRoleInput{}
		request.AssumeRolePolicyDocument = aws.String(policy)
		request.RoleName = e.Name

		response, err := t.Cloud.IAM.CreateRole(request)
		if err != nil {
			return fmt.Errorf("error creating IAMRole: %v", err)
		}

		e.ID = response.Role.RoleId
	} else {
		if changes.RolePolicyDocument != nil {
			glog.V(2).Infof("Updating IAMRole AssumeRolePolicy  %q", *e.Name)

			request := &iam.UpdateAssumeRolePolicyInput{}
			request.PolicyDocument = aws.String(policy)
			request.RoleName = e.Name

			_, err := t.Cloud.IAM.UpdateAssumeRolePolicy(request)
			if err != nil {
				return fmt.Errorf("error updating IAMRole: %v", err)
			}
		}
	}

	return nil //return output.AddAWSTags(cloud.Tags(), v, "vpc")
}
