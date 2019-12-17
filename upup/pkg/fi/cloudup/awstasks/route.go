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

//go:generate fitask -type=Route
type Route struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	RouteTable *RouteTable
	Instance   *Instance
	CIDR       *string

	// Either an InternetGateway or a NAT Gateway
	// MUST be provided.
	InternetGateway *InternetGateway
	NatGateway      *NatGateway
}

func (e *Route) Find(c *fi.Context) (*Route, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

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

	response, err := cloud.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}
	if response == nil || len(response.RouteTables) == 0 {
		return nil, nil
	} else {
		if len(response.RouteTables) != 1 {
			klog.Fatalf("found multiple RouteTables matching tags")
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
			if r.NatGatewayId != nil {
				actual.NatGateway = &NatGateway{ID: r.NatGatewayId}
			}
			if r.InstanceId != nil {
				actual.Instance = &Instance{ID: r.InstanceId}
			}

			if aws.StringValue(r.State) == "blackhole" {
				klog.V(2).Infof("found route is a blackhole route")
				// These should be nil anyway, but just in case...
				actual.Instance = nil
				actual.InternetGateway = nil
			}

			// Prevent spurious changes
			actual.Lifecycle = e.Lifecycle

			klog.V(2).Infof("found route matching cidr %s", *e.CIDR)
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
		if e.NatGateway != nil {
			targetCount++
		}
		if targetCount == 0 {
			return fmt.Errorf("InternetGateway or Instance or NatGateway is required")
		}
		if targetCount != 1 {
			return fmt.Errorf("Cannot set more than 1 InternetGateway or Instance or NatGateway")
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

		if e.InternetGateway == nil && e.NatGateway == nil {
			return fmt.Errorf("missing target for route")
		} else if e.InternetGateway != nil {
			request.GatewayId = checkNotNil(e.InternetGateway.ID)
		} else if e.NatGateway != nil {
			if err := e.NatGateway.waitAvailable(t.Cloud); err != nil {
				return err
			}

			request.NatGatewayId = checkNotNil(e.NatGateway.ID)
		}

		if e.Instance != nil {
			request.InstanceId = checkNotNil(e.Instance.ID)
		}

		klog.V(2).Infof("Creating Route with RouteTable:%q CIDR:%q", *e.RouteTable.ID, *e.CIDR)

		response, err := t.Cloud.EC2().CreateRoute(request)
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

		if e.InternetGateway == nil && e.NatGateway == nil {
			return fmt.Errorf("missing target for route")
		} else if e.InternetGateway != nil {
			request.GatewayId = checkNotNil(e.InternetGateway.ID)
		} else if e.NatGateway != nil {
			if err := e.NatGateway.waitAvailable(t.Cloud); err != nil {
				return err
			}

			request.NatGatewayId = checkNotNil(e.NatGateway.ID)
		}

		if e.Instance != nil {
			request.InstanceId = checkNotNil(e.Instance.ID)
		}

		klog.V(2).Infof("Updating Route with RouteTable:%q CIDR:%q", *e.RouteTable.ID, *e.CIDR)

		_, err := t.Cloud.EC2().ReplaceRoute(request)
		if err != nil {
			return fmt.Errorf("error updating Route: %v", err)
		}
	}

	return nil
}

func checkNotNil(s *string) *string {
	if s == nil {
		klog.Fatal("string pointer was unexpectedly nil")
	}
	return s
}

type terraformRoute struct {
	RouteTableID      *terraform.Literal `json:"route_table_id"`
	CIDR              *string            `json:"destination_cidr_block,omitempty"`
	InternetGatewayID *terraform.Literal `json:"gateway_id,omitempty"`
	NATGatewayID      *terraform.Literal `json:"nat_gateway_id,omitempty"`
	InstanceID        *terraform.Literal `json:"instance_id,omitempty"`
}

func (_ *Route) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Route) error {
	tf := &terraformRoute{
		CIDR:         e.CIDR,
		RouteTableID: e.RouteTable.TerraformLink(),
	}

	if e.InternetGateway == nil && e.NatGateway == nil {
		return fmt.Errorf("missing target for route")
	} else if e.InternetGateway != nil {
		tf.InternetGatewayID = e.InternetGateway.TerraformLink()
	} else if e.NatGateway != nil {
		tf.NATGatewayID = e.NatGateway.TerraformLink()
	}

	if e.Instance != nil {
		tf.InstanceID = e.Instance.TerraformLink()
	}

	return t.RenderResource("aws_route", *e.Name, tf)
}

type cloudformationRoute struct {
	RouteTableID      *cloudformation.Literal `json:"RouteTableId"`
	CIDR              *string                 `json:"DestinationCidrBlock,omitempty"`
	InternetGatewayID *cloudformation.Literal `json:"GatewayId,omitempty"`
	NATGatewayID      *cloudformation.Literal `json:"NatGatewayId,omitempty"`
	InstanceID        *cloudformation.Literal `json:"InstanceId,omitempty"`
}

func (_ *Route) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *Route) error {
	tf := &cloudformationRoute{
		CIDR:         e.CIDR,
		RouteTableID: e.RouteTable.CloudformationLink(),
	}

	if e.InternetGateway == nil && e.NatGateway == nil {
		return fmt.Errorf("missing target for route")
	} else if e.InternetGateway != nil {
		tf.InternetGatewayID = e.InternetGateway.CloudformationLink()
	} else if e.NatGateway != nil {
		tf.NATGatewayID = e.NatGateway.CloudformationLink()
	}

	if e.Instance != nil {
		return fmt.Errorf("instance cloudformation routes not yet implemented")
		//tf.InstanceID = e.Instance.CloudformationLink()
	}

	return t.RenderResource("AWS::EC2::Route", *e.Name, tf)
}
