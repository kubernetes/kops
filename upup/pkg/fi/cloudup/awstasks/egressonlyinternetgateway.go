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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type EgressOnlyInternetGateway struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID  *string
	VPC *VPC
	// Shared is set if this is a shared EgressOnlyInternetGateway
	Shared *bool

	// Tags is a map of aws tags that are added to the EgressOnlyInternetGateway
	Tags map[string]string
}

var _ fi.CompareWithID = &EgressOnlyInternetGateway{}

func (e *EgressOnlyInternetGateway) CompareWithID() *string {
	return e.ID
}

func findEgressOnlyInternetGateway(cloud awsup.AWSCloud, request *ec2.DescribeEgressOnlyInternetGatewaysInput) (*ec2.EgressOnlyInternetGateway, error) {
	response, err := cloud.EC2().DescribeEgressOnlyInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing EgressOnlyInternetGateways: %v", err)
	}
	if response == nil || len(response.EgressOnlyInternetGateways) == 0 {
		return nil, nil
	}

	if len(response.EgressOnlyInternetGateways) != 1 {
		return nil, fmt.Errorf("found multiple EgressOnlyInternetGateways matching tags")
	}
	igw := response.EgressOnlyInternetGateways[0]
	return igw, nil
}

func (e *EgressOnlyInternetGateway) Find(c *fi.Context) (*EgressOnlyInternetGateway, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &ec2.DescribeEgressOnlyInternetGatewaysInput{}

	shared := fi.ValueOf(e.Shared)
	if shared {
		if fi.ValueOf(e.VPC.ID) == "" {
			return nil, fmt.Errorf("VPC ID is required when EgressOnlyInternetGateway is shared")
		}

		request.Filters = []*ec2.Filter{awsup.NewEC2Filter("attachment.vpc-id", *e.VPC.ID)}
	} else {
		if e.ID != nil {
			request.EgressOnlyInternetGatewayIds = []*string{e.ID}
		} else {
			request.Filters = cloud.BuildFilters(e.Name)
		}
	}

	eigw, err := findEgressOnlyInternetGateway(cloud, request)
	if err != nil {
		return nil, err
	}
	if eigw == nil {
		return nil, nil
	}
	actual := &EgressOnlyInternetGateway{
		ID:   eigw.EgressOnlyInternetGatewayId,
		Name: findNameTag(eigw.Tags),
		Tags: intersectTags(eigw.Tags, e.Tags),
	}

	klog.V(2).Infof("found matching EgressOnlyInternetGateway %q", *actual.ID)

	for _, attachment := range eigw.Attachments {
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

	// We don't set the tags for a shared EIGW
	if fi.ValueOf(e.Shared) {
		actual.Tags = e.Tags
	}

	return actual, nil
}

func (e *EgressOnlyInternetGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *EgressOnlyInternetGateway) CheckChanges(a, e, changes *EgressOnlyInternetGateway) error {
	if a != nil {
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *EgressOnlyInternetGateway) RenderAWS(ctx *fi.Context, t *awsup.AWSAPITarget, a, e, changes *EgressOnlyInternetGateway) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Verify the EgressOnlyInternetGateway was found and matches our required settings
		if a == nil {
			return fmt.Errorf("EgressOnlyInternetGateway for shared VPC was not found")
		}

		return nil
	}

	if a == nil {
		klog.V(2).Infof("Creating EgressOnlyInternetGateway")

		request := &ec2.CreateEgressOnlyInternetGatewayInput{
			VpcId:             e.VPC.ID,
			TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeEgressOnlyInternetGateway, e.Tags),
		}

		response, err := t.Cloud.EC2().CreateEgressOnlyInternetGateway(request)
		if err != nil {
			return fmt.Errorf("error creating EgressOnlyInternetGateway: %v", err)
		}

		e.ID = response.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId
		return nil
	}

	return t.UpdateTags(*e.ID, e.Tags)
}

type terraformEgressOnlyInternetGateway struct {
	VPCID *terraformWriter.Literal `cty:"vpc_id"`
	Tags  map[string]string        `cty:"tags"`
}

func (_ *EgressOnlyInternetGateway) RenderTerraform(ctx *fi.Context, t *terraform.TerraformTarget, a, e, changes *EgressOnlyInternetGateway) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed

		// But ... attempt to discover the ID so TerraformLink works
		if e.ID == nil {
			request := &ec2.DescribeEgressOnlyInternetGatewaysInput{}
			vpcID := fi.ValueOf(e.VPC.ID)
			if vpcID == "" {
				return fmt.Errorf("VPC ID is required when EgressOnlyInternetGateway is shared")
			}
			request.Filters = []*ec2.Filter{awsup.NewEC2Filter("attachment.vpc-id", vpcID)}
			igw, err := findEgressOnlyInternetGateway(t.Cloud.(awsup.AWSCloud), request)
			if err != nil {
				return err
			}
			if igw == nil {
				klog.Warningf("Cannot find egress-only internet gateway for VPC %q", vpcID)
			} else {
				e.ID = igw.EgressOnlyInternetGatewayId
			}
		}

		return nil
	}

	tf := &terraformEgressOnlyInternetGateway{
		VPCID: e.VPC.TerraformLink(),
		Tags:  e.Tags,
	}

	return t.RenderResource("aws_egress_only_internet_gateway", *e.Name, tf)
}

func (e *EgressOnlyInternetGateway) TerraformLink() *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if EgressOnlyInternetGateway is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing EgressOnlyInternetGateway with id %q", *e.ID)
		return terraformWriter.LiteralFromStringValue(*e.ID)
	}

	return terraformWriter.LiteralProperty("aws_egress_only_internet_gateway", *e.Name, "id")
}
