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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
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
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

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
		glog.V(2).Infof("found matching RouteTableAssociation %q", *actual.ID)
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

func (_ *RouteTableAssociation) Render(t *spotinst.Target, a, e, changes *RouteTableAssociation) error {
	if a == nil {
		// TODO: We might do better just to make the subnet the primary key here

		glog.V(2).Infof("Checking for existing RouteTableAssociation to subnet")
		existing, err := findExistingRouteTableForSubnet(t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud), e.Subnet)
		if err != nil {
			return fmt.Errorf("error checking for existing RouteTableAssociation: %v", err)
		}

		if existing != nil {
			for _, a := range existing.Associations {
				if aws.StringValue(a.SubnetId) != aws.StringValue(e.Subnet.ID) {
					continue
				}
				glog.V(2).Infof("Creating RouteTableAssociation")
				request := &ec2.DisassociateRouteTableInput{
					AssociationId: a.RouteTableAssociationId,
				}

				_, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().DisassociateRouteTable(request)
				if err != nil {
					return fmt.Errorf("error disassociating existing RouteTable from subnet: %v", err)
				}
			}
		}

		glog.V(2).Infof("Creating RouteTableAssociation")
		request := &ec2.AssociateRouteTableInput{
			SubnetId:     e.Subnet.ID,
			RouteTableId: e.RouteTable.ID,
		}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().AssociateRouteTable(request)
		if err != nil {
			return fmt.Errorf("error creating RouteTableAssociation: %v", err)
		}

		e.ID = response.AssociationId
	}

	return nil // no tags
}
