package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=RouteTableAssociation
type RouteTableAssociation struct {
	Name *string

	ID         *string
	RouteTable *RouteTable
	Subnet     *Subnet
}

func (s *RouteTableAssociation) CompareWithID() *string {
	return s.ID
}

func (e *RouteTableAssociation) Find(c *fi.Context) (*RouteTableAssociation, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	routeTableID := e.RouteTable.ID
	subnetID := e.Subnet.ID

	if routeTableID == nil || subnetID == nil {
		return nil, nil
	}

	request := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []*string{routeTableID},
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

func (_ *RouteTableAssociation) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *RouteTableAssociation) error {
	if a == nil {
		glog.V(2).Infof("Creating RouteTableAssociation")

		request := &ec2.AssociateRouteTableInput{
			SubnetId:     e.Subnet.ID,
			RouteTableId: e.RouteTable.ID,
		}

		response, err := t.Cloud.EC2.AssociateRouteTable(request)
		if err != nil {
			return fmt.Errorf("error creating RouteTableAssociation: %v", err)
		}

		e.ID = response.AssociationId
	}

	return nil // no tags
}

type terraformRouteTableAssociation struct {
	SubnetID     *terraform.Literal `json:"subnet_id"`
	RouteTableID *terraform.Literal `json:"route_table_id"`
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
