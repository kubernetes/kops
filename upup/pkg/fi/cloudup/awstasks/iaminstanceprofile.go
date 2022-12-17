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

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog/v2"
)

// +kops:fitask
type IAMInstanceProfile struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Tags map[string]string

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
		if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error getting IAMInstanceProfile: %v", err)
	}

	return response.InstanceProfile, nil
}

func (e *IAMInstanceProfile) Find(c *fi.CloudupContext) (*IAMInstanceProfile, error) {
	cloud := c.T.Cloud.(awsup.AWSCloud)

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
		Tags: mapIAMTagsToMap(p.Tags),
	}

	e.ID = actual.ID
	e.Name = actual.Name

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	return actual, nil
}

func (e *IAMInstanceProfile) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (s *IAMInstanceProfile) CheckChanges(a, e, changes *IAMInstanceProfile) error {
	if a != nil {
		if fi.ValueOf(e.Name) == "" && !fi.ValueOf(e.Shared) {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *IAMInstanceProfile) RenderAWS(ctx *fi.CloudupContext, t *awsup.AWSAPITarget, a, e, changes *IAMInstanceProfile) error {
	if fi.ValueOf(e.Shared) {
		if a == nil {
			return fmt.Errorf("instance role profile with id %q not found", fi.ValueOf(e.ID))
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

		tagRequest := &iam.TagInstanceProfileInput{
			InstanceProfileName: e.Name,
			Tags:                mapToIAMTags(e.Tags),
		}
		_, err = t.Cloud.IAM().TagInstanceProfile(tagRequest)
		if err != nil {
			if awsup.AWSErrorCode(err) == awsup.AWSErrCodeInvalidAction {
				klog.Warningf("Ignoring unsupported IAMInstanceProfile tagging %v", *a.Name)
			} else {
				return fmt.Errorf("error tagging IAMInstanceProfile: %v", err)
			}
		}

		e.ID = response.InstanceProfile.InstanceProfileId
		e.Name = response.InstanceProfile.InstanceProfileName
	} else {
		if changes.Tags != nil {
			if len(a.Tags) > 0 {
				existingTagKeys := make([]*string, 0)
				for k := range a.Tags {
					existingTagKeys = append(existingTagKeys, &k)
				}
				untagRequest := &iam.UntagInstanceProfileInput{
					InstanceProfileName: a.Name,
					TagKeys:             existingTagKeys,
				}
				_, err := t.Cloud.IAM().UntagInstanceProfile(untagRequest)
				if err != nil {
					return fmt.Errorf("error untagging IAMInstanceProfile: %v", err)
				}
			}
			if len(e.Tags) > 0 {
				tagRequest := &iam.TagInstanceProfileInput{
					InstanceProfileName: a.Name,
					Tags:                mapToIAMTags(e.Tags),
				}
				_, err := t.Cloud.IAM().TagInstanceProfile(tagRequest)
				if err != nil {
					if awsup.AWSErrorCode(err) == awsup.AWSErrCodeInvalidAction {
						klog.Warningf("Ignoring unsupported IAMInstanceProfile tagging %v", *a.Name)
					} else {
						return fmt.Errorf("error tagging IAMInstanceProfile: %v", err)
					}
				}
			}
		}
	}

	return nil
}

func (_ *IAMInstanceProfile) RenderTerraform(ctx *fi.CloudupContext, t *terraform.TerraformTarget, a, e, changes *IAMInstanceProfile) error {
	// Done on IAMInstanceProfileRole
	return nil
}

func (e *IAMInstanceProfile) TerraformLink() *terraformWriter.Literal {
	if fi.ValueOf(e.Shared) {
		return terraformWriter.LiteralFromStringValue(fi.ValueOf(e.Name))
	}
	return terraformWriter.LiteralProperty("aws_iam_instance_profile", *e.Name, "id")
}
