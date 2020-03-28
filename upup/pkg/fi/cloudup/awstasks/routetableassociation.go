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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=RouteTableAssociation
type RouteTableAssociation struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID         *string
	RouteTable *RouteTable
	Subnet     *Subnet
}

func (s *RouteTableAssociation) CompareWithID() *string {
	return s.ID
}

func (e *RouteTableAssociation) Find(c *fi.Context) (*RouteTableAssociation, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	routeTableID := e.RouteTable.ID
	subnetID := e.Subnet.ID

	if routeTableID == nil || subnetID == nil {
		return nil, nil
	}

	request := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []*string{routeTableID},
	}

	response, err := cloud.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}
	if response == nil || len(response.RouteTables) == 0 {
		return nil, nil
	}

	if len(response.RouteTables) != 1 {
		return nil, fmt.Errorf("found multiple RouteTables matching tags")
	}
	rt := response.RouteTables[0]
	for _, rta := range rt.Associations {
		if aws.StringValue(rta.SubnetId) != *subnetID {
			continue
		}
		actual := &RouteTableAssociation{
			Name:       e.Name,
			ID:         rta.RouteTableAssociationId,
			RouteTable: &RouteTable{ID: rta.RouteTableId},
			Subnet:     &Subnet{ID: rta.SubnetId},
		}
		klog.V(2).Infof("found matching RouteTableAssociation %q", *actual.ID)
		e.ID = actual.ID

		// Prevent spurious changes
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *RouteTableAssociation) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *RouteTableAssociation) CheckChanges(a, e, changes *RouteTableAssociation) error {
	if a != nil {
		if e.RouteTable == nil {
			return fi.RequiredField("RouteTable")
		}
		if e.Subnet == nil {
			return fi.RequiredField("Subnet")
		}
	}
	if a != nil {
		if changes.RouteTable != nil {
			return fi.CannotChangeField("RouteTable")
		}
		if changes.Subnet != nil {
			return fi.CannotChangeField("Subnet")
		}
	}
	return nil
}

func findExistingRouteTableForSubnet(cloud awsup.AWSCloud, subnet *Subnet) (*ec2.RouteTable, error) {
	if subnet == nil {
		return nil, fmt.Errorf("subnet not set")
	}
	if subnet.ID == nil {
		return nil, fmt.Errorf("subnet ID not set")
	}

	subnetID := fi.StringValue(subnet.ID)

	request := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{awsup.NewEC2Filter("association.subnet-id", subnetID)},
	}
	response, err := cloud.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables for subnet %q: %v", subnetID, err)
	}
	if response == nil || len(response.RouteTables) == 0 {
		return nil, nil
	}

	if len(response.RouteTables) != 1 {
		return nil, fmt.Errorf("found multiple RouteTables attached to subnet")
	}
	rt := response.RouteTables[0]
	return rt, nil
}

func (_ *RouteTableAssociation) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *RouteTableAssociation) error {
	if a == nil {
		// TODO: We might do better just to make the subnet the primary key here

		klog.V(2).Infof("Checking for existing RouteTableAssociation to subnet")
		existing, err := findExistingRouteTableForSubnet(t.Cloud, e.Subnet)
		if err != nil {
			return fmt.Errorf("error checking for existing RouteTableAssociation: %v", err)
		}

		if existing != nil {
			for _, a := range existing.Associations {
				if aws.StringValue(a.SubnetId) != aws.StringValue(e.Subnet.ID) {
					continue
				}
				klog.V(2).Infof("Creating RouteTableAssociation")
				request := &ec2.DisassociateRouteTableInput{
					AssociationId: a.RouteTableAssociationId,
				}

				_, err := t.Cloud.EC2().DisassociateRouteTable(request)
				if err != nil {
					return fmt.Errorf("error disassociating existing RouteTable from subnet: %v", err)
				}
			}
		}

		klog.V(2).Infof("Creating RouteTableAssociation")
		request := &ec2.AssociateRouteTableInput{
			SubnetId:     e.Subnet.ID,
			RouteTableId: e.RouteTable.ID,
		}

		response, err := t.Cloud.EC2().AssociateRouteTable(request)
		if err != nil {
			return fmt.Errorf("error creating RouteTableAssociation: %v", err)
		}

		e.ID = response.AssociationId
	}

	return nil // no tags
}

type terraformRouteTableAssociation struct {
	SubnetID     *terraform.Literal `json:"subnet_id" cty:"subnet_id"`
	RouteTableID *terraform.Literal `json:"route_table_id" cty:"route_table_id"`
}

func (_ *RouteTableAssociation) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RouteTableAssociation) error {
	tf := &terraformRouteTableAssociation{
		SubnetID:     e.Subnet.TerraformLink(),
		RouteTableID: e.RouteTable.TerraformLink(),
	}

	return t.RenderResource("aws_route_table_association", *e.Name, tf)
}

func (e *RouteTableAssociation) TerraformLink() *terraform.Literal {
	return terraform.LiteralSelfLink("aws_route_table_association", *e.Name)
}

type cloudformationRouteTableAssociation struct {
	SubnetID     *cloudformation.Literal `json:"SubnetId,omitempty"`
	RouteTableID *cloudformation.Literal `json:"RouteTableId,omitempty"`
}

func (_ *RouteTableAssociation) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *RouteTableAssociation) error {
	cf := &cloudformationRouteTableAssociation{
		SubnetID:     e.Subnet.CloudformationLink(),
		RouteTableID: e.RouteTable.CloudformationLink(),
	}

	return t.RenderResource("AWS::EC2::SubnetRouteTableAssociation", *e.Name, cf)
}

func (e *RouteTableAssociation) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::EC2::SubnetRouteTableAssociation", *e.Name)
}
