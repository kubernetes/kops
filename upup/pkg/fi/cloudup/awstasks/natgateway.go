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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=NatGateway
type NatGateway struct {
	Name      *string
	ElasticIp *ElasticIP
	Subnet    *Subnet
	ID        *string
}

var _ fi.CompareWithID = &NatGateway{} // Validate the IDs

func (e *NatGateway) CompareWithID() *string {
	return e.ID
}

func (e *NatGateway) Find(c *fi.Context) (*NatGateway, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	id := e.ID

	// Find via tag on foreign resource
	if id == nil && e.Subnet != nil {
		var filters []*ec2.Filter
		filters = append(filters, awsup.NewEC2Filter("key", "AssociatedNatgateway"))
		if e.Subnet.ID == nil {
			glog.V(2).Infof("Unable to find subnet, bypassing Find() for NGW")
			return nil, nil
		}
		filters = append(filters, awsup.NewEC2Filter("resource-id", *e.Subnet.ID))

		request := &ec2.DescribeTagsInput{
			Filters: filters,
		}

		response, err := cloud.EC2().DescribeTags(request)
		if err != nil {
			return nil, fmt.Errorf("error listing tags: %v", err)
		}

		if response == nil || len(response.Tags) == 0 {
			return nil, nil
		}

		if len(response.Tags) != 1 {
			return nil, fmt.Errorf("found multiple tags for: %v", e)
		}
		t := response.Tags[0]
		id = t.Value
		glog.V(2).Infof("Found nat gateway via tag: %v", *id)
	}

	if id != nil {
		request := &ec2.DescribeNatGatewaysInput{}
		request.NatGatewayIds = []*string{id}
		response, err := cloud.EC2().DescribeNatGateways(request)
		if err != nil {
			return nil, fmt.Errorf("error listing NAT Gateways: %v", err)
		}

		if response == nil || len(response.NatGateways) == 0 {
			glog.V(2).Infof("Unable to find Nat Gateways")
			return nil, nil
		}
		if len(response.NatGateways) != 1 {
			return nil, fmt.Errorf("found multiple NAT Gateways for: %v", e)
		}
		a := response.NatGateways[0]
		actual := &NatGateway{
			ID: a.NatGatewayId,
		}
		actual.Subnet = e.Subnet
		if len(a.NatGatewayAddresses) == 0 {
			// Not sure if this ever happens
			actual.ElasticIp = nil
		} else if len(a.NatGatewayAddresses) == 1 {
			actual.ElasticIp = &ElasticIP{ID: a.NatGatewayAddresses[0].AllocationId}
		} else {
			return nil, fmt.Errorf("found multiple elastic IPs attached to NatGateway %q", aws.StringValue(a.NatGatewayId))
		}

		// NATGateways don't have a Name (no tags), so we set the name to avoid spurious changes
		actual.Name = e.Name

		e.ID = actual.ID
		return actual, nil
	}
	return nil, nil
}

func (s *NatGateway) CheckChanges(a, e, changes *NatGateway) error {

	// New
	if a == nil {
		if e.ElasticIp == nil {
			return fi.RequiredField("ElasticIp")
		}
		if e.Subnet == nil {
			return fi.RequiredField("Subnet")
		}
	}

	// Delta
	if a != nil {
		if changes.ElasticIp != nil {
			return fi.CannotChangeField("ElasticIp")
		}
		if changes.Subnet != nil {
			return fi.CannotChangeField("Subnet")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
	}
	return nil
}

func (e *NatGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *NatGateway) waitAvailable(cloud awsup.AWSCloud) error {
	// It takes 'forever' (up to 5 min...) for a NatGateway to become available after it has been created
	// We have to wait until it is actually up

	// TODO: Cache availability status

	id := aws.StringValue(e.ID)
	if id == "" {
		return fmt.Errorf("NAT Gateway %q did not have ID", e.Name)
	}

	glog.Infof("Waiting for NAT Gateway %q to be available", id)
	params := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{e.ID},
	}
	err := cloud.EC2().WaitUntilNatGatewayAvailable(params)
	if err != nil {
		return fmt.Errorf("error waiting for NAT Gateway %q to be available: %v", id, err)
	}

	return nil
}

func (_ *NatGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NatGateway) error {

	// New NGW
	var id *string
	if a == nil {
		glog.V(2).Infof("Creating Nat Gateway")

		request := &ec2.CreateNatGatewayInput{}
		request.AllocationId = e.ElasticIp.ID
		request.SubnetId = e.Subnet.ID
		response, err := t.Cloud.EC2().CreateNatGateway(request)
		if err != nil {
			return fmt.Errorf("Error creating Nat Gateway: %v", err)
		}
		e.ID = response.NatGateway.NatGatewayId
		id = e.ID
	} else {
		id = a.ID
	}

	// Tag the associated subnet
	if e.Subnet == nil {
		return fmt.Errorf("Subnet not set")
	} else if e.Subnet.ID == nil {
		return fmt.Errorf("Subnet ID not set")
	}

	tags := make(map[string]string)
	tags["AssociatedNatgateway"] = *id
	err := t.AddAWSTags(*e.Subnet.ID, tags)
	if err != nil {
		return fmt.Errorf("Unable to tag subnet %v", err)
	}
	return nil
}

type terraformNATGateway struct {
	AllocationID *terraform.Literal `json:"allocation_id,omitempty"`
	SubnetID     *terraform.Literal `json:"subnet_id,omitempty"`
}

func (_ *NatGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NatGateway) error {
	tf := &terraformNATGateway{
		AllocationID: e.ElasticIp.TerraformLink(),
		SubnetID:     e.Subnet.TerraformLink(),
	}

	return t.RenderResource("aws_nat_gateway", *e.Name, tf)
}

func (e *NatGateway) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_nat_gateway", *e.Name, "id")
}
