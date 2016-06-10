package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=RouteTable
type RouteTable struct {
	Name *string
	ID   *string
	VPC  *VPC
}

var _ fi.CompareWithID = &RouteTable{}

func (e *RouteTable) CompareWithID() *string {
	return e.ID
}

func (e *RouteTable) Find(c *fi.Context) (*RouteTable, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeRouteTablesInput{}
	if e.ID != nil {
		request.RouteTableIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2.DescribeRouteTables(request)
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

	actual := &RouteTable{
		ID:   rt.RouteTableId,
		VPC:  &VPC{ID: rt.VpcId},
		Name: e.Name,
	}
	glog.V(2).Infof("found matching RouteTable %q", *actual.ID)
	e.ID = actual.ID

	return actual, nil
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

		glog.V(2).Infof("Creating RouteTable with VPC: %q", *vpcID)

		request := &ec2.CreateRouteTableInput{
			VpcId: vpcID,
		}

		response, err := t.Cloud.EC2.CreateRouteTable(request)
		if err != nil {
			return fmt.Errorf("error creating RouteTable: %v", err)
		}

		rt := response.RouteTable
		e.ID = rt.RouteTableId
	}

	return t.AddAWSTags(*e.ID, t.Cloud.BuildTags(e.Name))
}

type terraformRouteTable struct {
	VPCID *terraform.Literal `json:"vpc_id"`
	Tags  map[string]string  `json:"tags,omitempty"`
}

func (_ *RouteTable) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RouteTable) error {
	cloud := t.Cloud.(*awsup.AWSCloud)

	tf := &terraformRouteTable{
		VPCID: e.VPC.TerraformLink(),
		Tags:  cloud.BuildTags(e.Name),
	}

	return t.RenderResource("aws_route_table", *e.Name, tf)
}

func (e *RouteTable) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_route_table", *e.Name, "id")
}
