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

const (
	NatType = "Nat"
)

//go:generate fitask -type=EIP
type EIP struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Region     *string
	ID         *string
	IpAddress  *string
	NatGateway *NatGateway
	Available  *bool
}

var _ fi.CompareWithID = &EIP{}

func (e *EIP) CompareWithID() *string {
	return e.ID
}

func (e *EIP) Find(c *fi.Context) (*EIP, error) {
	if e.NatGateway == nil || e.NatGateway.ID == nil {
		klog.V(4).Infof("NatGateway / NatGatewayId not found for %s, skipping Find", fi.StringValue(e.Name))
		return nil, nil
	}

	cloud := c.Cloud.(aliup.ALICloud)
	describeEipAddressesArgs := &ecs.DescribeEipAddressesArgs{
		RegionId:               common.Region(cloud.Region()),
		AssociatedInstanceType: ecs.AssociatedInstanceTypeNat,
		AssociatedInstanceId:   fi.StringValue(e.NatGateway.ID),
	}

	eipAddresses, _, err := cloud.VpcClient().DescribeEipAddresses(describeEipAddressesArgs)
	if err != nil {
		return nil, fmt.Errorf("error finding EIPs: %v", err)
	}
	// Don't exist EIPs with specified NatGateway.
	if len(eipAddresses) == 0 {
		return nil, nil
	}
	if len(eipAddresses) > 1 {
		klog.V(4).Infof("The number of specified EIPs with the same NatGatewayId exceeds 1, eipName:%q", *e.Name)
	}

	klog.V(2).Infof("found matching EIPs: %q", *e.Name)

	actual := &EIP{}
	actual.IpAddress = fi.String(eipAddresses[0].IpAddress)
	actual.ID = fi.String(eipAddresses[0].AllocationId)
	actual.Available = fi.Bool(eipAddresses[0].Status == ecs.EipStatusAvailable)
	if eipAddresses[0].InstanceId != "" {
		actual.NatGateway = &NatGateway{
			ID: fi.String(eipAddresses[0].InstanceId),
		}
		actual.Region = fi.String(cloud.Region())
	}
	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle
	actual.Name = e.Name
	e.ID = actual.ID
	e.Available = actual.Available
	klog.V(4).Infof("found matching EIP %v", actual)
	return actual, nil
}

func (e *EIP) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *EIP) CheckChanges(a, e, changes *EIP) error {
	return nil
}

func (_ *EIP) RenderALI(t *aliup.ALIAPITarget, a, e, changes *EIP) error {

	if a == nil {
		klog.V(2).Infof("Creating new EIP for NatGateway:%q", fi.StringValue(e.NatGateway.Name))

		allocateEipAddressArgs := &ecs.AllocateEipAddressArgs{
			RegionId: common.Region(t.Cloud.Region()),
		}

		eipAddress, allocationId, err := t.Cloud.VpcClient().AllocateEipAddress(allocateEipAddressArgs)
		if err != nil {
			return fmt.Errorf("error creating eip: %v", err)
		}
		e.IpAddress = fi.String(eipAddress)
		e.ID = fi.String(allocationId)
		e.Available = fi.Bool(true)
	}

	associateEipAddressArgs := &ecs.AssociateEipAddressArgs{
		AllocationId: fi.StringValue(e.ID),
		InstanceId:   fi.StringValue(e.NatGateway.ID),
		InstanceType: ecs.Nat,
	}

	if fi.BoolValue(e.Available) {
		err := t.Cloud.VpcClient().NewAssociateEipAddress(associateEipAddressArgs)
		if err != nil {
			return fmt.Errorf("error associating eip to natGateway: %v", err)
		}
	}

	return nil
}

type terraformEip struct {
}

type terraformEipAssociation struct {
	InstanceID   *terraform.Literal `json:"instance_id,omitempty"`
	AllocationID *terraform.Literal `json:"allocation_id,omitempty"`
}

func (_ *EIP) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *EIP) error {
	tf := &terraformEip{}
	err := t.RenderResource("alicloud_eip", *e.Name, tf)
	if err != nil {
		return err
	}

	associationtf := &terraformEipAssociation{
		InstanceID:   e.NatGateway.TerraformLink(),
		AllocationID: e.TerraformLink(),
	}

	return t.RenderResource("alicloud_eip_association", *e.Name+"_asso", associationtf)
}

func (e *EIP) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("alicloud_eip", *e.Name, "id")
}
