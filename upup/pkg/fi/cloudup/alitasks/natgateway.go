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

package alitasks

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=NatGateway
type NatGateway struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	VPC    *VPC
	Region *string
	ID     *string
}

var _ fi.CompareWithID = &NatGateway{}

func (e *NatGateway) CompareWithID() *string {
	return e.ID
}

func (e *NatGateway) Find(c *fi.Context) (*NatGateway, error) {
	if e.VPC == nil || e.VPC.ID == nil {
		klog.V(4).Infof("VPC / VPCID not found for %s, skipping Find", fi.StringValue(e.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)
	request := &ecs.DescribeNatGatewaysArgs{
		RegionId: common.Region(cloud.Region()),
		VpcId:    fi.StringValue(e.VPC.ID),
	}

	natGateways, _, err := cloud.VpcClient().DescribeNatGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing NatGateways: %v", err)
	}

	// Don't exist NatGateways with specified VPC.
	if len(natGateways) == 0 {
		return nil, nil
	}
	if len(natGateways) != 1 {
		return nil, fmt.Errorf("found multiple NatGateways for %q", fi.StringValue(e.ID))
	}
	natGateway := natGateways[0]

	actual := &NatGateway{}
	actual.ID = fi.String(natGateway.NatGatewayId)

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle
	actual.Name = e.Name
	actual.Region = e.Region
	actual.VPC = &VPC{ID: &natGateway.VpcId}

	e.ID = actual.ID
	klog.V(4).Infof("found matching NatGateway %v", actual)
	return actual, nil
}

func (s *NatGateway) CheckChanges(a, e, changes *NatGateway) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (e *NatGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *NatGateway) RenderALI(t *aliup.ALIAPITarget, a, e, changes *NatGateway) error {
	if a == nil {
		request := &ecs.CreateNatGatewayArgs{
			RegionId: common.Region(t.Cloud.Region()),
			VpcId:    fi.StringValue(e.VPC.ID),
			Name:     fi.StringValue(e.Name),
		}

		response, err := t.Cloud.VpcClient().CreateNatGateway(request)
		if err != nil {
			return fmt.Errorf("error creating NatGateway: %v", err)
		}
		e.ID = fi.String(response.NatGatewayId)
	}
	return nil
}

type terraformNatGateway struct {
	Name  *string            `json:"name,omitempty"`
	VpcId *terraform.Literal `json:"vpc_id,omitempty"`
}

func (_ *NatGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NatGateway) error {
	tf := &terraformNatGateway{
		Name:  e.Name,
		VpcId: e.VPC.TerraformLink(),
	}

	return t.RenderResource("alicloud_nat_gateway", *e.Name, tf)
}

func (e *NatGateway) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_nat_gateway", *e.Name, "id")
}
