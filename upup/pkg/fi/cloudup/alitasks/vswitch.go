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

	"k8s.io/klog"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=VSwitch
type VSwitch struct {
	Name      *string
	VSwitchId *string

	Lifecycle *fi.Lifecycle
	ZoneId    *string

	CidrBlock *string
	Region    *common.Region
	VPC       *VPC
	Shared    *bool
}

var _ fi.CompareWithID = &VSwitch{}

func (v *VSwitch) CompareWithID() *string {
	return v.VSwitchId
}

func (v *VSwitch) Find(c *fi.Context) (*VSwitch, error) {
	if v.VPC == nil || v.VPC.ID == nil {
		klog.V(4).Infof("VPC / VPCID not found for %s, skipping Find", fi.StringValue(v.Name))
		return nil, nil
	}
	cloud := c.Cloud.(aliup.ALICloud)

	describeVSwitchesArgs := &ecs.DescribeVSwitchesArgs{
		VpcId:    fi.StringValue(v.VPC.ID),
		RegionId: common.Region(cloud.Region()),
		ZoneId:   fi.StringValue(v.ZoneId),
	}

	if v.VSwitchId != nil && fi.StringValue(v.VSwitchId) != "" {
		describeVSwitchesArgs.VSwitchId = fi.StringValue(v.VSwitchId)
	}

	vswitchList, _, err := cloud.EcsClient().DescribeVSwitches(describeVSwitchesArgs)
	if err != nil {
		return nil, fmt.Errorf("error listing VSwitchs: %v", err)
	}

	if len(vswitchList) == 0 {
		return nil, nil
	}

	var actual *VSwitch

	for _, vswitch := range vswitchList {
		if actual != nil {
			return nil, fmt.Errorf("found multiple matching VSwitchs")
		}
		if vswitch.VSwitchId == fi.StringValue(v.VSwitchId) {
			actual = &VSwitch{
				Name:      fi.String(vswitch.VSwitchName),
				VSwitchId: fi.String(vswitch.VSwitchId),
				VPC: &VPC{
					ID: fi.String(vswitch.VpcId),
				},

				ZoneId:    fi.String(vswitch.ZoneId),
				CidrBlock: fi.String(vswitch.CidrBlock),
			}
			continue
		}
		if vswitch.CidrBlock == fi.StringValue(v.CidrBlock) && !fi.BoolValue(v.Shared) {
			klog.V(2).Infof("found matching VSwitch with name: %q", *v.Name)
			actual := &VSwitch{
				Name:      fi.String(vswitch.VSwitchName),
				VSwitchId: fi.String(vswitch.VSwitchId),
				VPC: &VPC{
					ID: fi.String(vswitch.VpcId),
				},

				ZoneId:    fi.String(vswitch.ZoneId),
				CidrBlock: fi.String(vswitch.CidrBlock),
			}
			v.VSwitchId = actual.VSwitchId
		}
	}

	if actual != nil {
		klog.V(2).Infof("found matching subnet %q", *actual.VSwitchId)
		if v.VSwitchId == nil {
			v.VSwitchId = actual.VSwitchId
		}

		// Prevent spurious changes
		actual.Lifecycle = v.Lifecycle
		actual.Name = v.Name
		actual.Shared = v.Shared
		return actual, nil
	}

	return nil, nil
}

func (v *VSwitch) CheckChanges(a, e, changes *VSwitch) error {
	if a == nil {
		if e.CidrBlock == nil {
			return fi.RequiredField("CidrBlock")
		}
		if e.ZoneId == nil {
			return fi.RequiredField("ZoneId")
		}
	} else {
		if changes.ZoneId != nil {
			return fi.CannotChangeField("ZoneId")
		}
	}
	return nil
}

func (v *VSwitch) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *VSwitch) RenderALI(t *aliup.ALIAPITarget, a, e, changes *VSwitch) error {
	if e.VPC == nil || e.VPC.ID == nil {
		return fmt.Errorf("error updating VSwitch, lack of VPCId")
	}
	if a == nil {
		if e.VSwitchId != nil && fi.StringValue(e.VSwitchId) != "" {
			klog.V(2).Infof("Shared VSwitch with VSwitchID: %q", *e.VSwitchId)
			return nil
		}

		klog.V(2).Infof("Creating VSwitch with name: %q", *e.Name)

		createVSwitchArgs := &ecs.CreateVSwitchArgs{
			ZoneId:      fi.StringValue(e.ZoneId),
			CidrBlock:   fi.StringValue(e.CidrBlock),
			VpcId:       fi.StringValue(e.VPC.ID),
			VSwitchName: fi.StringValue(e.Name),
		}

		vswitchId, err := t.Cloud.EcsClient().CreateVSwitch(createVSwitchArgs)
		if err != nil {
			return fmt.Errorf("error creating VSwitch: %v,%v", err, createVSwitchArgs)
		}
		e.VSwitchId = fi.String(vswitchId)
	}

	return nil
}

type terraformVSwitch struct {
	Name      *string            `json:"name,omitempty" cty:"name"`
	CidrBlock *string            `json:"cidr_block,omitempty" cty:"cidr_block"`
	ZoneId    *string            `json:"availability_zone,omitempty" cty:"availability_zone"`
	VPCId     *terraform.Literal `json:"vpc_id,omitempty" cty:"vpc_id"`
}

func (_ *VSwitch) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VSwitch) error {
	tf := &terraformVSwitch{
		Name:      e.Name,
		CidrBlock: e.CidrBlock,
		ZoneId:    e.ZoneId,
		VPCId:     e.VPC.TerraformLink(),
	}

	return t.RenderResource("alicloud_vswitch", *e.Name, tf)
}

func (v *VSwitch) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_vswitch", *v.Name, "id")
}
