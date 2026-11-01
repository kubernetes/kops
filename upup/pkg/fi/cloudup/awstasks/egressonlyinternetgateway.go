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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

var _ fi.CompareWithID = (*EgressOnlyInternetGateway)(nil)

func (e *EgressOnlyInternetGateway) CompareWithID() *string {
	return e.ID
}

func findEgressOnlyInternetGateway(ctx context.Context, cloud awsup.AWSCloud, vpcID string) (*ec2types.EgressOnlyInternetGateway, error) {
	request := &ec2.DescribeEgressOnlyInternetGatewaysInput{}
	paginator := ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(cloud.EC2(), request)
	for paginator.HasMorePages() {
		response, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing EgressOnlyInternetGateways: %v", err)
		}
		if response == nil {
			return nil, nil
		}

		for _, igw := range response.EgressOnlyInternetGateways {
			if *igw.Attachments[0].VpcId == vpcID { // NOTE: Since EgressOnlyInternetGateway is always associated with a single VPC, length of `Attachments[]` is always 1.
				return &igw, nil
			}
		}
	}

	return nil, nil
}

func (e *EgressOnlyInternetGateway) Find(c *fi.CloudupContext) (*EgressOnlyInternetGateway, error) {
	ctx := c.Context()
	cloud := awsup.GetCloud(c)

	vpcID := fi.ValueOf(e.VPC.ID)
	shared := fi.ValueOf(e.Shared)
	if vpcID == "" {
		if shared {
			return nil, fmt.Errorf("VPC ID is required when EgressOnlyInternetGateway is shared")
		}
		return nil, nil
	}

	eigw, err := findEgressOnlyInternetGateway(ctx, cloud, vpcID)
	if eigw == nil || err != nil {
		return nil, err
	}
	actual := &EgressOnlyInternetGateway{
		ID:   eigw.EgressOnlyInternetGatewayId,
		Name: findNameTag(eigw.Tags),
		Tags: intersectTags(eigw.Tags, e.Tags),
		VPC:  &VPC{ID: eigw.Attachments[0].VpcId},
	}

	klog.V(2).Infof("found matching EgressOnlyInternetGateway %q", *actual.ID)

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

func (e *EgressOnlyInternetGateway) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (s *EgressOnlyInternetGateway) CheckChanges(a, e, changes *EgressOnlyInternetGateway) error {
	if a != nil {
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *EgressOnlyInternetGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *EgressOnlyInternetGateway) error {
	ctx := context.TODO()
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
			TagSpecifications: awsup.EC2TagSpecification(ec2types.ResourceTypeEgressOnlyInternetGateway, e.Tags),
		}

		response, err := t.Cloud.EC2().CreateEgressOnlyInternetGateway(ctx, request)
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

func (_ *EgressOnlyInternetGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *EgressOnlyInternetGateway) error {
	ctx := context.TODO()
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed

		// But ... attempt to discover the ID so TerraformLink works
		if e.ID == nil {
			vpcID := fi.ValueOf(e.VPC.ID)
			if vpcID == "" {
				return fmt.Errorf("VPC ID is required when EgressOnlyInternetGateway is shared")
			}

			igw, err := findEgressOnlyInternetGateway(ctx, t.Cloud.(awsup.AWSCloud), vpcID)
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
