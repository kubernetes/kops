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

//go:generate fitask -type=RouteTable
type RouteTable struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID  *string
	VPC *VPC
}

var _ fi.CompareWithID = &RouteTable{}

func (e *RouteTable) CompareWithID() *string {
	return e.ID
}

func (e *RouteTable) Find(c *fi.Context) (*RouteTable, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	rt, err := e.findEc2RouteTable(cloud)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, nil
	}

	actual := &RouteTable{
		ID:   rt.RouteTableId,
		VPC:  &VPC{ID: rt.VpcId},
		Name: e.Name,
	}
	glog.V(2).Infof("found matching RouteTable %q", *actual.ID)
	e.ID = actual.ID

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *RouteTable) findEc2RouteTable(cloud awsup.AWSCloud) (*ec2.RouteTable, error) {
	request := &ec2.DescribeRouteTablesInput{}
	if e.ID != nil {
		request.RouteTableIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
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

	return rt, nil
}

func (e *RouteTable) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *RouteTable) CheckChanges(a, e, changes *RouteTable) error {
	if a == nil {
		if e.VPC == nil {
			return fi.RequiredField("VPC")
		}
	}
	if a != nil {
		if changes.VPC != nil && changes.VPC.ID != nil {
			return fi.CannotChangeField("VPC")
		}
	}
	return nil
}

func (_ *RouteTable) Render(t *spotinst.Target, a, e, changes *RouteTable) error {
	if a == nil {
		vpcID := e.VPC.ID
		if vpcID == nil {
			return fi.RequiredField("VPC.ID")
		}

		glog.V(2).Infof("Creating RouteTable with VPC: %q", *vpcID)

		request := &ec2.CreateRouteTableInput{
			VpcId: vpcID,
		}

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateRouteTable(request)
		if err != nil {
			return fmt.Errorf("error creating RouteTable: %v", err)
		}

		rt := response.RouteTable
		e.ID = rt.RouteTableId
	}

	return t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.ID, t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).BuildTags(e.Name))
}
