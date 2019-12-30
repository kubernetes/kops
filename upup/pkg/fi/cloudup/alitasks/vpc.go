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

//go:generate fitask -type=VPC
type VPC struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID     *string
	Region *string
	CIDR   *string
	Shared *bool
	Tags   map[string]string
}

var _ fi.CompareWithID = &VPC{}

func (e *VPC) CompareWithID() *string {
	return e.ID
}

func (e *VPC) Find(c *fi.Context) (*VPC, error) {
	cloud := c.Cloud.(aliup.ALICloud)

	request := &ecs.DescribeVpcsArgs{
		RegionId: common.Region(cloud.Region()),
	}

	if e.ID != nil {
		request.VpcId = fi.StringValue(e.ID)
	}

	vpcs, _, err := cloud.EcsClient().DescribeVpcs(request)
	if err != nil {
		return nil, fmt.Errorf("error listing VPCs: %v", err)
	}

	if len(vpcs) == 0 {
		return nil, nil
	}

	var actual *VPC
	for _, vpc := range vpcs {
		if actual != nil {
			return nil, fmt.Errorf("found multiple matching VPCs")
		}
		if vpc.VpcId == fi.StringValue(e.ID) {
			actual = &VPC{
				ID:   fi.String(vpc.VpcId),
				CIDR: fi.String(vpc.CidrBlock),
				Name: fi.String(vpc.VpcName),
			}
			continue
		}
		if vpc.CidrBlock == fi.StringValue(e.CIDR) && vpc.VpcName == fi.StringValue(e.Name) {
			actual = &VPC{
				ID:   fi.String(vpc.VpcId),
				CIDR: fi.String(vpc.CidrBlock),
				Name: fi.String(vpc.VpcName),
			}
		}
	}

	if actual != nil {
		klog.V(2).Infof("found matching VPC %v", actual)

		// Prevent spurious comparison failures
		actual.Shared = e.Shared
		if e.ID == nil {
			e.ID = actual.ID
		}
		actual.Lifecycle = e.Lifecycle
		actual.Name = e.Name
		actual.Region = fi.String(cloud.Region())

		return actual, nil
	}

	return nil, nil
}

func (s *VPC) CheckChanges(a, e, changes *VPC) error {
	if a == nil {
		if e.CIDR == nil {
			return fi.RequiredField("CIDR")
		}
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.CIDR != nil {
			// TODO: Do we want to destroy & recreate the VPC?
			return fi.CannotChangeField("CIDR")
		}
	}

	return nil
}

func (e *VPC) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *VPC) RenderALI(t *aliup.ALIAPITarget, a, e, changes *VPC) error {

	if fi.BoolValue(e.Shared) && a == nil {
		return fmt.Errorf("VPC with id %q not found", fi.StringValue(e.ID))
	}

	if a == nil {
		if e.ID != nil && fi.StringValue(e.ID) != "" {
			klog.V(2).Infof("Shared VPC with VPCID: %q", *e.ID)
			return nil
		}
		klog.V(2).Infof("Creating VPC with CIDR: %q", *e.CIDR)

		request := &ecs.CreateVpcArgs{
			RegionId:  common.Region(t.Cloud.Region()),
			CidrBlock: fi.StringValue(e.CIDR),
			VpcName:   fi.StringValue(e.Name),
		}

		response, err := t.Cloud.EcsClient().CreateVpc(request)
		if err != nil {
			return fmt.Errorf("error creating VPC: %v", err)
		}
		e.ID = fi.String(response.VpcId)
	}
	return nil
}

type terraformVPC struct {
	CIDR *string `json:"cidr_block,omitempty"`
	Name *string `json:"vpc_name,omitempty"`
}

func (_ *VPC) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VPC) error {
	if err := t.AddOutputVariable("id", e.TerraformLink()); err != nil {
		return err
	}

	tf := &terraformVPC{
		CIDR: e.CIDR,
		Name: e.Name,
	}

	return t.RenderResource("alicloud_vpc", *e.Name, tf)
}

func (e *VPC) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_vpc", *e.Name, "id")
}
