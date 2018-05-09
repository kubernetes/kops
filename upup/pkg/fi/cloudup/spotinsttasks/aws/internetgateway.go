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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=InternetGateway
type InternetGateway struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID     *string
	VPC    *VPC
	Shared *bool
}

var _ fi.CompareWithID = &InternetGateway{}

func (e *InternetGateway) CompareWithID() *string {
	return e.ID
}

func findInternetGateway(cloud awsup.AWSCloud, request *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	response, err := cloud.EC2().DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateways: %v", err)
	}
	if response == nil || len(response.InternetGateways) == 0 {
		return nil, nil
	}

	if len(response.InternetGateways) != 1 {
		return nil, fmt.Errorf("found multiple InternetGateways matching tags")
	}
	igw := response.InternetGateways[0]
	return igw, nil
}

func (e *InternetGateway) Find(c *fi.Context) (*InternetGateway, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	request := &ec2.DescribeInternetGatewaysInput{}

	shared := fi.BoolValue(e.Shared)
	if shared {
		if fi.StringValue(e.VPC.ID) == "" {
			return nil, fmt.Errorf("VPC ID is required when InternetGateway is shared")
		}

		request.Filters = []*ec2.Filter{awsup.NewEC2Filter("attachment.vpc-id", *e.VPC.ID)}
	} else {
		if e.ID != nil {
			request.InternetGatewayIds = []*string{e.ID}
		} else {
			request.Filters = cloud.BuildFilters(e.Name)
		}
	}

	igw, err := findInternetGateway(cloud, request)
	if err != nil {
		return nil, err
	}
	if igw == nil {
		return nil, nil
	}
	actual := &InternetGateway{
		ID:   igw.InternetGatewayId,
		Name: findNameTag(igw.Tags),
	}

	glog.V(2).Infof("found matching InternetGateway %q", *actual.ID)

	for _, attachment := range igw.Attachments {
		actual.VPC = &VPC{ID: attachment.VpcId}
	}

	// Prevent spurious comparison failures
	actual.Shared = e.Shared
	actual.Lifecycle = e.Lifecycle
	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (e *InternetGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *InternetGateway) CheckChanges(a, e, changes *InternetGateway) error {
	if a != nil {
		// TODO: I think we can change it; we just detach & attach
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *InternetGateway) Render(t *spotinst.Target, a, e, changes *InternetGateway) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the InternetGateway was found and matches our required settings
		if a == nil {
			return fmt.Errorf("InternetGateway for shared VPC was not found")
		}

		return nil
	}

	if a == nil {
		glog.V(2).Infof("Creating InternetGateway")

		request := &ec2.CreateInternetGatewayInput{}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateInternetGateway(request)
		if err != nil {
			return fmt.Errorf("error creating InternetGateway: %v", err)
		}

		e.ID = response.InternetGateway.InternetGatewayId
	}

	if a == nil || (changes != nil && changes.VPC != nil) {
		glog.V(2).Infof("Creating InternetGatewayAttachment")

		attachRequest := &ec2.AttachInternetGatewayInput{
			VpcId:             e.VPC.ID,
			InternetGatewayId: e.ID,
		}

		_, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().AttachInternetGateway(attachRequest)
		if err != nil {
			return fmt.Errorf("error attaching InternetGateway to VPC: %v", err)
		}
	}

	tags := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).BuildTags(e.Name)
	if shared {
		// Don't tag shared resources
		tags = nil
	}
	return t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.ID, tags)
}
