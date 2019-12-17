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

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=InternetGateway
type InternetGateway struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID  *string
	VPC *VPC
	// Shared is set if this is a shared InternetGateway
	Shared *bool

	// Tags is a map of aws tags that are added to the InternetGateway
	Tags map[string]string
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
	cloud := c.Cloud.(awsup.AWSCloud)

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
		Tags: intersectTags(igw.Tags, e.Tags),
	}

	klog.V(2).Infof("found matching InternetGateway %q", *actual.ID)

	for _, attachment := range igw.Attachments {
		actual.VPC = &VPC{ID: attachment.VpcId}
	}

	// Prevent spurious comparison failures
	actual.Shared = e.Shared
	actual.Lifecycle = e.Lifecycle
	if shared {
		actual.Name = e.Name
	}
	if e.ID == nil {
		e.ID = actual.ID
	}

	// We don't set the tags for a shared IGW
	if fi.BoolValue(e.Shared) {
		actual.Tags = e.Tags
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

func (_ *InternetGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InternetGateway) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the InternetGateway was found and matches our required settings
		if a == nil {
			return fmt.Errorf("InternetGateway for shared VPC was not found")
		}

		return nil
	}

	if a == nil {
		klog.V(2).Infof("Creating InternetGateway")

		request := &ec2.CreateInternetGatewayInput{}

		response, err := t.Cloud.EC2().CreateInternetGateway(request)
		if err != nil {
			return fmt.Errorf("error creating InternetGateway: %v", err)
		}

		e.ID = response.InternetGateway.InternetGatewayId
	}

	if a == nil || (changes != nil && changes.VPC != nil) {
		klog.V(2).Infof("Creating InternetGatewayAttachment")

		attachRequest := &ec2.AttachInternetGatewayInput{
			VpcId:             e.VPC.ID,
			InternetGatewayId: e.ID,
		}

		_, err := t.Cloud.EC2().AttachInternetGateway(attachRequest)
		if err != nil {
			return fmt.Errorf("error attaching InternetGateway to VPC: %v", err)
		}
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

type terraformInternetGateway struct {
	VPCID *terraform.Literal `json:"vpc_id"`
	Tags  map[string]string  `json:"tags,omitempty"`
}

func (_ *InternetGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InternetGateway) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed

		// But ... attempt to discover the ID so TerraformLink works
		if e.ID == nil {
			request := &ec2.DescribeInternetGatewaysInput{}
			vpcID := fi.StringValue(e.VPC.ID)
			if vpcID == "" {
				return fmt.Errorf("VPC ID is required when InternetGateway is shared")
			}
			request.Filters = []*ec2.Filter{awsup.NewEC2Filter("attachment.vpc-id", vpcID)}
			igw, err := findInternetGateway(t.Cloud.(awsup.AWSCloud), request)
			if err != nil {
				return err
			}
			if igw == nil {
				klog.Warningf("Cannot find internet gateway for VPC %q", vpcID)
			} else {
				e.ID = igw.InternetGatewayId
			}
		}

		return nil
	}

	tf := &terraformInternetGateway{
		VPCID: e.VPC.TerraformLink(),
		Tags:  e.Tags,
	}

	return t.RenderResource("aws_internet_gateway", *e.Name, tf)
}

func (e *InternetGateway) TerraformLink() *terraform.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if InternetGateway is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing InternetGateway with id %q", *e.ID)
		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralProperty("aws_internet_gateway", *e.Name, "id")
}

type cloudformationInternetGateway struct {
	Tags []cloudformationTag `json:"Tags,omitempty"`
}

type cloudformationVpcGatewayAttachment struct {
	VpcId             *cloudformation.Literal `json:"VpcId,omitempty"`
	InternetGatewayId *cloudformation.Literal `json:"InternetGatewayId,omitempty"`
}

func (_ *InternetGateway) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *InternetGateway) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not cloudformation owned / managed

		// But ... attempt to discover the ID so CloudformationLink works
		if e.ID == nil {
			request := &ec2.DescribeInternetGatewaysInput{}
			vpcID := fi.StringValue(e.VPC.ID)
			if vpcID == "" {
				return fmt.Errorf("VPC ID is required when InternetGateway is shared")
			}
			request.Filters = []*ec2.Filter{awsup.NewEC2Filter("attachment.vpc-id", vpcID)}
			igw, err := findInternetGateway(t.Cloud.(awsup.AWSCloud), request)
			if err != nil {
				return err
			}
			if igw == nil {
				klog.Warningf("Cannot find internet gateway for VPC %q", vpcID)
			} else {
				e.ID = igw.InternetGatewayId
			}
		}

		return nil
	}

	{
		cf := &cloudformationInternetGateway{
			Tags: buildCloudformationTags(e.Tags),
		}

		err := t.RenderResource("AWS::EC2::InternetGateway", *e.Name, cf)
		if err != nil {
			return err
		}
	}

	{
		cf := &cloudformationVpcGatewayAttachment{
			VpcId:             e.VPC.CloudformationLink(),
			InternetGatewayId: e.CloudformationLink(),
		}

		err := t.RenderResource("AWS::EC2::VPCGatewayAttachment", *e.Name, cf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *InternetGateway) CloudformationLink() *cloudformation.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if InternetGateway is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing InternetGateway with id %q", *e.ID)
		return cloudformation.LiteralString(*e.ID)
	}

	return cloudformation.Ref("AWS::EC2::InternetGateway", *e.Name)
}
