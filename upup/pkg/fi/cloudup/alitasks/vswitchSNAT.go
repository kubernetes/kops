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

//go:generate fitask -type=VSwitchSNAT
type VSwitchSNAT struct {
	Name      *string
	Lifecycle *fi.Lifecycle
	ID        *string

	VSwitch     *VSwitch
	NatGateway  *NatGateway
	EIP         *EIP
	SnatTableId *string
	// Shared is set if this is a shared VSwitch
	Shared *bool
}

var _ fi.CompareWithID = &VSwitchSNAT{}

func (v *VSwitchSNAT) CompareWithID() *string {
	return v.Name
}

func (v *VSwitchSNAT) Find(c *fi.Context) (*VSwitchSNAT, error) {
	if v.VSwitch == nil || v.VSwitch.VSwitchId == nil {
		klog.V(4).Infof("VSwitch / VSwitchId not found for %s, skipping Find", fi.StringValue(v.Name))
		return nil, nil
	}
	if v.NatGateway == nil || v.NatGateway.ID == nil {
		klog.V(4).Infof("NatGateway / NatGatewayId not found for %s, skipping Find", fi.StringValue(v.Name))
		return nil, nil
	}
	if v.EIP == nil || v.EIP.Name == nil {
		klog.V(4).Infof("EIP / EIPName not found for %s, skipping Find", fi.StringValue(v.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)

	describeNatGatewaysArgs := &ecs.DescribeNatGatewaysArgs{
		RegionId:     common.Region(cloud.Region()),
		NatGatewayId: fi.StringValue(v.NatGateway.ID),
	}

	natGateways, _, err := cloud.VpcClient().DescribeNatGateways(describeNatGatewaysArgs)
	if err != nil {
		return nil, fmt.Errorf("error listing NatGateways: %v", err)
	}
	if len(natGateways) == 0 {
		klog.V(4).Infof("NatGateway not found for %s, skipping Find", fi.StringValue(v.Name))
		return nil, nil
	}
	if len(natGateways[0].SnatTableIds.SnatTableId) == 0 {
		return nil, nil
	}

	for _, snatTableId := range natGateways[0].SnatTableIds.SnatTableId {

		describeSnatTableEntriesArgs := &ecs.DescribeSnatTableEntriesArgs{
			RegionId:    common.Region(cloud.Region()),
			SnatTableId: snatTableId,
		}
		snatTableEntries, _, err := cloud.VpcClient().DescribeSnatTableEntries(describeSnatTableEntriesArgs)
		if err != nil {
			return nil, fmt.Errorf("error listing snatTableEntries: %v", err)
		}
		if len(snatTableEntries) == 0 {
			continue
		}

		for _, snatEntry := range snatTableEntries {
			if snatEntry.SourceVSwitchId == fi.StringValue(v.VSwitch.VSwitchId) {
				actual := &VSwitchSNAT{}
				actual.ID = fi.String(snatEntry.SnatEntryId)
				v.ID = actual.ID
				actual.VSwitch = v.VSwitch
				actual.NatGateway = &NatGateway{ID: v.NatGateway.ID}
				actual.SnatTableId = fi.String(snatTableId)
				v.SnatTableId = actual.SnatTableId
				// Prevent spurious changes
				actual.Shared = v.Shared
				actual.Name = v.Name
				actual.Lifecycle = v.Lifecycle

				describeEIPArgs := &ecs.DescribeEipAddressesArgs{
					RegionId:   common.Region(cloud.Region()),
					EipAddress: snatEntry.SnatIp,
				}

				eips, _, err := cloud.VpcClient().DescribeEipAddresses(describeEIPArgs)
				if err != nil {
					return nil, fmt.Errorf("error listing EIP: %v", err)
				}

				if len(eips) == 0 {
					klog.V(4).Infof("EIP not found for %s, skipping Find", snatEntry.SnatIp)
					return nil, nil
				}

				eip := eips[0]

				actual.EIP = &EIP{
					ID:        &eip.AllocationId,
					IpAddress: &eip.IpAddress,
				}

				return actual, nil
			}
		}
	}
	v.SnatTableId = fi.String(natGateways[0].SnatTableIds.SnatTableId[0])
	return nil, nil
}

func (v *VSwitchSNAT) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (v *VSwitchSNAT) CheckChanges(a, e, changes *VSwitchSNAT) error {
	if e.VSwitch == nil {
		return fi.RequiredField("VPC")
	}

	if e.NatGateway == nil {
		return fi.RequiredField("CIDRBlock")
	}

	if a != nil && changes != nil {
		if changes.VSwitch != nil {
			return fi.CannotChangeField("VSwitch")
		}

		if changes.NatGateway != nil {
			return fi.CannotChangeField("NatGateway")
		}
	}
	return nil
}

func (_ *VSwitchSNAT) RenderALI(t *aliup.ALIAPITarget, a, e, changes *VSwitchSNAT) error {

	if a == nil {
		createSnatEntryArgs := &ecs.CreateSnatEntryArgs{
			RegionId:        common.Region(t.Cloud.Region()),
			SnatTableId:     fi.StringValue(e.SnatTableId),
			SourceVSwitchId: fi.StringValue(e.VSwitch.VSwitchId),
			SnatIp:          fi.StringValue(e.EIP.IpAddress),
		}
		resp, err := t.Cloud.VpcClient().CreateSnatEntry(createSnatEntryArgs)
		if err != nil {
			return fmt.Errorf("error creating SnatEntry: %v,%v", err, createSnatEntryArgs)
		}
		e.ID = fi.String(resp.SnatEntryId)
	}
	return nil
}

type terraformVSwitchSNAT struct {
	SnatTableId *string            `json:"snat_table_id,omitempty"`
	VSwitchId   *terraform.Literal `json:"source_vswitch_id,omitempty"`
}

func (_ *VSwitchSNAT) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *VSwitchSNAT) error {
	tf := &terraformVSwitchSNAT{
		SnatTableId: e.SnatTableId,
		VSwitchId:   e.VSwitch.TerraformLink(),
	}

	return t.RenderResource("alicloud_snat_entry", *e.Name, tf)
}
