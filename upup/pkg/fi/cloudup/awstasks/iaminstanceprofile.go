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
	"time"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog"
)

//go:generate fitask -type=IAMInstanceProfile
type IAMInstanceProfile struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID     *string
	Shared *bool
}

var _ fi.CompareWithID = &IAMInstanceProfile{}

func (e *IAMInstanceProfile) CompareWithID() *string {
	return e.Name
}

// findIAMInstanceProfile retrieves the InstanceProfile with specified name
// It returns nil,nil if not found
func findIAMInstanceProfile(cloud awsup.AWSCloud, name string) (*iam.InstanceProfile, error) {
	request := &iam.GetInstanceProfileInput{InstanceProfileName: aws.String(name)}

	response, err := cloud.IAM().GetInstanceProfile(request)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "NoSuchEntity" {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error getting IAMInstanceProfile: %v", err)
	}

	return response.InstanceProfile, nil
}

func (e *IAMInstanceProfile) Find(c *fi.Context) (*IAMInstanceProfile, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	p, err := findIAMInstanceProfile(cloud, *e.Name)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	actual := &IAMInstanceProfile{
		ID:   p.InstanceProfileId,
		Name: p.InstanceProfileName,
	}

	e.ID = actual.ID
	e.Name = actual.Name

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	return actual, nil
}

func (e *IAMInstanceProfile) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *IAMInstanceProfile) CheckChanges(a, e, changes *IAMInstanceProfile) error {
	if a != nil {
		if fi.StringValue(e.Name) == "" && !fi.BoolValue(e.Shared) {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *IAMInstanceProfile) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMInstanceProfile) error {
	if fi.BoolValue(e.Shared) {
		if a == nil {
			return fmt.Errorf("instance role profile with id %q not found", fi.StringValue(e.ID))
		}
	} else if a == nil {
		klog.V(2).Infof("Creating IAMInstanceProfile with Name:%q", *e.Name)

		request := &iam.CreateInstanceProfileInput{
			InstanceProfileName: e.Name,
		}

		response, err := t.Cloud.IAM().CreateInstanceProfile(request)
		if err != nil {
			return fmt.Errorf("error creating IAMInstanceProfile: %v", err)
		}

		e.ID = response.InstanceProfile.InstanceProfileId
		e.Name = response.InstanceProfile.InstanceProfileName

		// IAM instance profile seems to be highly asynchronous
		// and if we don't wait creating dependent resources fail
		attempt := 0
		for {
			if attempt > 10 {
				klog.Warningf("unable to retrieve newly-created IAM instance profile %q; timed out", *e.Name)
				break
			}

			ip, err := findIAMInstanceProfile(t.Cloud, *e.Name)
			if err != nil {
				klog.Warningf("ignoring error while retrieving newly-created IAM instance profile %q: %v", *e.Name, err)
			}

			if ip != nil {
				// Found
				klog.V(4).Infof("Found IAM instance profile %q", *e.Name)
				break
			}

			// TODO: Use a real backoff algorithm
			time.Sleep(3 * time.Second)
			attempt++
		}
	}

	// TODO: Should we use path as our tag?
	return nil // No tags in IAM
}

func (_ *IAMInstanceProfile) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMInstanceProfile) error {
	// Done on IAMInstanceProfileRole
	return nil
}

func (e *IAMInstanceProfile) TerraformLink() *terraform.Literal {
	if fi.BoolValue(e.Shared) {
		return terraform.LiteralFromStringValue(fi.StringValue(e.Name))
	}
	return terraform.LiteralProperty("aws_iam_instance_profile", *e.Name, "id")
}

func (_ *IAMInstanceProfile) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *IAMInstanceProfile) error {
	// Done on IAMInstanceProfileRole
	return nil
}

func (e *IAMInstanceProfile) CloudformationLink() *cloudformation.Literal {
	if fi.BoolValue(e.Shared) {
		return cloudformation.LiteralString(fi.StringValue(e.Name))
	}
	return cloudformation.Ref("AWS::IAM::InstanceProfile", fi.StringValue(e.Name))
}
