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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=IAMInstanceProfile
type IAMInstanceProfile struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID *string
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
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

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

func (_ *IAMInstanceProfile) Render(t *spotinst.Target, a, e, changes *IAMInstanceProfile) error {
	if a == nil {
		glog.V(2).Infof("Creating IAMInstanceProfile with Name:%q", *e.Name)

		request := &iam.CreateInstanceProfileInput{
			InstanceProfileName: e.Name,
		}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).IAM().CreateInstanceProfile(request)
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
				glog.Warningf("unable to retrieve newly-created IAM instance profile %q; timed out", *e.Name)
				break
			}

			ip, err := findIAMInstanceProfile(t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud), *e.Name)
			if err != nil {
				glog.Warningf("ignoring error while retrieving newly-created IAM instance profile %q: %v", *e.Name, err)
			}

			if ip != nil {
				// Found
				glog.V(4).Infof("Found IAM instance profile %q", *e.Name)
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
