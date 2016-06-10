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

//go:generate fitask -type=Route
type Route struct {
	Name *string

	RouteTable      *RouteTable
	InternetGateway *InternetGateway
	Instance        *Instance
	CIDR            *string
}

func (e *Route) Find(c *fi.Context) (*Route, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	if e.RouteTable == nil || e.CIDR == nil {
		// TODO: Move to validate?
		return nil, nil
	}

	if e.RouteTable.ID == nil {
		return nil, nil
	}

	request := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []*string{e.RouteTable.ID},
	}

	response, err := cloud.EC2.DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}
	if response == nil || len(response.RouteTables) == 0 {
		return nil, nil
	} else {
		if len(response.RouteTables) != 1 {
			glog.Fatalf("found multiple RouteTables matching tags")
		}
		rt := response.RouteTables[0]
		for _, r := range rt.Routes {
			if aws.StringValue(r.DestinationCidrBlock) != *e.CIDR {
				continue
			}
			actual := &Route{
				Name:       e.Name,
				RouteTable: &RouteTable{ID: rt.RouteTableId},
				CIDR:       r.DestinationCidrBlock,
			}
			if r.GatewayId != nil {
				actual.InternetGateway = &InternetGateway{ID: r.GatewayId}
			}
			if r.InstanceId != nil {
				actual.Instance = &Instance{ID: r.InstanceId}
			}

			if aws.StringValue(r.State) == "blackhole" {
				glog.V(2).Infof("found route is a blackhole route")
				// These should be nil anyway, but just in case...
				actual.Instance = nil
				actual.InternetGateway = nil
			}

			glog.V(2).Infof("found route matching cidr %s", *e.CIDR)
			return actual, nil
		}
	}

	return nil, nil
}

func (e *Route) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Route) CheckChanges(a, e, changes *Route) error {
	if a == nil {
		// TODO: Create validate method?
		if e.RouteTable == nil {
			return fi.RequiredField("RouteTable")
		}
		if e.CIDR == nil {
			return fi.RequiredField("CIDR")
		}
		targetCount := 0
		if e.InternetGateway != nil {
			targetCount++
		}
		if e.Instance != nil {
			targetCount++
		}
		if targetCount == 0 {
			return fmt.Errorf("InternetGateway or Instance is required")
		}
		if targetCount != 1 {
			return fmt.Errorf("Cannot set both InternetGateway and Instance")
		}
	}

	if a != nil {
		if changes.RouteTable != nil {
			return fi.CannotChangeField("RouteTable")
		}
		if changes.CIDR != nil {
			return fi.CannotChangeField("CIDR")
		}
	}
	return nil
}

func (_ *Route) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Route) error {
	if a == nil {
		request := &ec2.CreateRouteInput{}
		request.RouteTableId = checkNotNil(e.RouteTable.ID)
		request.DestinationCidrBlock = checkNotNil(e.CIDR)

		if e.InternetGateway != nil {
			request.GatewayId = checkNotNil(e.InternetGateway.ID)
		}

		if e.Instance != nil {
			request.InstanceId = checkNotNil(e.Instance.ID)
		}

		glog.V(2).Infof("Creating Route with RouteTable:%q CIDR:%q", *e.RouteTable.ID, *e.CIDR)

		response, err := t.Cloud.EC2.CreateRoute(request)
		if err != nil {
			return fmt.Errorf("error creating Route: %v", err)
		}

		if !aws.BoolValue(response.Return) {
			return fmt.Errorf("create Route request failed: %v", response)
		}
	} else {
		request := &ec2.ReplaceRouteInput{}
		request.RouteTableId = checkNotNil(e.RouteTable.ID)
		request.DestinationCidrBlock = checkNotNil(e.CIDR)

		if e.InternetGateway != nil {
			request.GatewayId = checkNotNil(e.InternetGateway.ID)
		}

		if e.Instance != nil {
			request.InstanceId = checkNotNil(e.Instance.ID)
		}

		glog.V(2).Infof("Updating Route with RouteTable:%q CIDR:%q", *e.RouteTable.ID, *e.CIDR)

		_, err := t.Cloud.EC2.ReplaceRoute(request)
		if err != nil {
			return fmt.Errorf("error updating Route: %v", err)
		}
	}

	return nil
}

func checkNotNil(s *string) *string {
	if s == nil {
		glog.Fatal("string pointer was unexpectedly nil")
	}
	return s
}

type terraformRoute struct {
	RouteTableID      *terraform.Literal `json:"route_table_id"`
	CIDR              *string            `json:"destination_cidr_block,omitempty"`
	InternetGatewayID *terraform.Literal `json:"gateway_id,omitempty"`
	InstanceID        *terraform.Literal `json:"instance_id,omitempty"`
}

func (_ *Route) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Route) error {
	tf := &terraformRoute{
		CIDR:         e.CIDR,
		RouteTableID: e.RouteTable.TerraformLink(),
	}

	if e.InternetGateway != nil {
		tf.InternetGatewayID = e.InternetGateway.TerraformLink()
	}

	if e.Instance != nil {
		tf.InstanceID = e.Instance.TerraformLink()
	}

	return t.RenderResource("aws_route", *e.Name, tf)
}
