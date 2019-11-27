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

//go:generate fitask -type=RouteTable
type RouteTable struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID  *string
	VPC *VPC

	// Shared is set if this is a shared RouteTable
	Shared *bool
	// Tags is a map of aws tags that are added to the RouteTable
	Tags map[string]string
}

var _ fi.CompareWithID = &RouteTable{}

func (e *RouteTable) CompareWithID() *string {
	return e.ID
}

func (e *RouteTable) Find(c *fi.Context) (*RouteTable, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	var rt *ec2.RouteTable
	var err error

	if e.ID != nil {
		rt, err = findRouteTableByID(cloud, *e.ID)
		if err != nil {
			return nil, err
		}
	}

	// Try finding by name
	if rt == nil && e.Tags["Name"] != "" {
		rt, err = findRouteTableByFilters(cloud, cloud.BuildFilters(e.Name))
		if err != nil {
			return nil, err
		}
	}

	// Try finding by shared cluster tag, along with role (so it isn't ambiguous)
	if rt == nil && e.Tags[awsup.TagNameKopsRole] != "" {
		var filters []*ec2.Filter
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("tag-key"),
			Values: aws.StringSlice([]string{"kubernetes.io/cluster/" + c.Cluster.Name}),
		})
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("tag:" + awsup.TagNameKopsRole),
			Values: aws.StringSlice([]string{e.Tags[awsup.TagNameKopsRole]}),
		})

		rt, err = findRouteTableByFilters(cloud, filters)
		if err != nil {
			return nil, err
		}
	}

	if rt == nil {
		return nil, nil
	}

	actual := &RouteTable{
		ID:   rt.RouteTableId,
		VPC:  &VPC{ID: rt.VpcId},
		Name: e.Name,
		Tags: intersectTags(rt.Tags, e.Tags),
	}
	klog.V(2).Infof("found matching RouteTable %q", *actual.ID)
	e.ID = actual.ID

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	return actual, nil
}

func findRouteTableByID(cloud awsup.AWSCloud, id string) (*ec2.RouteTable, error) {
	request := &ec2.DescribeRouteTablesInput{}
	request.RouteTableIds = []*string{&id}

	response, err := cloud.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}
	if response == nil || len(response.RouteTables) == 0 {
		return nil, nil
	}

	if len(response.RouteTables) != 1 {
		return nil, fmt.Errorf("found multiple RouteTables matching ID")
	}
	rt := response.RouteTables[0]

	return rt, nil
}

func findRouteTableByFilters(cloud awsup.AWSCloud, filters []*ec2.Filter) (*ec2.RouteTable, error) {
	request := &ec2.DescribeRouteTablesInput{}
	request.Filters = filters

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

func (_ *RouteTable) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *RouteTable) error {
	if a == nil {
		vpcID := e.VPC.ID
		if vpcID == nil {
			return fi.RequiredField("VPC.ID")
		}

		klog.V(2).Infof("Creating RouteTable with VPC: %q", *vpcID)

		request := &ec2.CreateRouteTableInput{
			VpcId: vpcID,
		}

		response, err := t.Cloud.EC2().CreateRouteTable(request)
		if err != nil {
			return fmt.Errorf("error creating RouteTable: %v", err)
		}

		rt := response.RouteTable
		e.ID = rt.RouteTableId
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

type terraformRouteTable struct {
	VPCID *terraform.Literal `json:"vpc_id"`
	Tags  map[string]string  `json:"tags,omitempty"`
}

func (_ *RouteTable) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RouteTable) error {
	// We use the role tag as a concise and stable identifier
	tag := e.Tags[awsup.TagNameKopsRole]
	if tag != "" {
		if err := t.AddOutputVariable("route_table_"+tag+"_id", e.TerraformLink()); err != nil {
			return err
		}
	}

	tf := &terraformRouteTable{
		VPCID: e.VPC.TerraformLink(),
		Tags:  e.Tags,
	}

	return t.RenderResource("aws_route_table", *e.Name, tf)
}

func (e *RouteTable) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_route_table", *e.Name, "id")
}

type cloudformationRouteTable struct {
	VPCID *cloudformation.Literal `json:"VpcId,omitempty"`
	Tags  []cloudformationTag     `json:"Tags,omitempty"`
}

func (_ *RouteTable) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *RouteTable) error {
	cf := &cloudformationRouteTable{
		VPCID: e.VPC.CloudformationLink(),
		Tags:  buildCloudformationTags(e.Tags),
	}

	return t.RenderResource("AWS::EC2::RouteTable", *e.Name, cf)
}

func (e *RouteTable) CloudformationLink() *cloudformation.Literal {
	return cloudformation.Ref("AWS::EC2::RouteTable", *e.Name)
}
