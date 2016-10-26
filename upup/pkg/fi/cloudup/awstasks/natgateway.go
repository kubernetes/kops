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

package awstasks

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=natgateway
type NATGateway struct {
	Name         *string
	ID           *string
	AllocationID *string
	SubnetID     *string
	VPCID        *string
}

var _ fi.CompareWithID = &NATGateway{} // Validate the IDs

func (e *NATGateway) CompareWithID() *string {
	return e.ID
}

func (e *NATGateway) Find(c *fi.Context) (*NATGateway, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &ec2.DescribeNatGatewaysInput{}

	if fi.StringValue(e.ID) != "" {
		request.NatGatewayIds = []*string{e.ID}
	} else {
		request.Filter = cloud.BuildFilters(e.SubnetID)
	}

	response, err := cloud.EC2().DescribeNatGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing NAT Gateways: %v", err)
	}
	if response == nil || len(response.NatGateways) == 0 {
		return nil, nil
	}

	if len(response.NatGateways) != 1 {
		return nil, fmt.Errorf("found multiple NAT Gateways matching tags")
	}
	ngw := response.NatGateways[0]
	actual := &NATGateway{
		ID:   ngw.NatGatewayId,
		VPCID: ngw.VpcId,
		SubnetID: ngw.SubnetId,
	}

	glog.V(4).Infof("found matching NAT gateway %v", actual)

	// Allocation ID
	if actual.ID != nil {
		request := &ec2.DescribeAddressesInput{}
		request.Filters = cloud.BuildFilters(e.VPCID)
		response, err := cloud.EC2().DescribeAddresses(request)
		if err != nil || len(response.Addresses) != 1 {
			return nil, fmt.Errorf("error querying for elastic ip support: %v", err)
		}
		actual.AllocationID = response.Addresses[0].AllocationId
	}

	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (s *NATGateway) CheckChanges(a, e, changes *NATGateway) error {
	if a == nil {
		if e.AllocationID == nil {
			return fi.RequiredField("AllocationID")
		}
		if e.SubnetID == nil {
			return fi.RequiredField("SubnetID")
		}
	}
	if a != nil {
		if changes.AllocationID != nil {
			// TODO: Do we want to destroy & recreate the VPC?
			return fi.CannotChangeField("AllocationID")
		}
	}
	return nil
}

func (e *NATGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *NATGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NATGateway) error {
	if a == nil {
		glog.V(2).Infof("Creating NGW with Allocation ID: %q", *e.AllocationID)

		request := &ec2.CreateNatGatewayInput{
			AllocationId: e.AllocationID,
			SubnetId: e.SubnetID,
		}

		response, err := t.Cloud.EC2().CreateNatGateway(request)
		if err != nil {
			return fmt.Errorf("error creating Nat gateway: %v", err)
		}

		e.ID = response.NatGateway.NatGatewayId
	}

	return nil
}

type terraformNatGateway struct {
	AllocationId *string           `json:"AllocationID,omitempty"`
	SubnetID     *bool             `json:"SubnetID,omitempty"`
}

func (_ *NATGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NATGateway) error {
	//	cloud := t.Cloud.(awsup.AWSCloud)

	tf := &terraformNatGateway{
		AllocationId:  e.AllocationID,
		//SubnetID:      e.SubnetID,
	}

	return t.RenderResource("aws_natgateway", *e.AllocationID, tf)
}

func (e *NATGateway) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_natgateway", *e.AllocationID, "id")
}