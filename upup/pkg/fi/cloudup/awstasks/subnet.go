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
	"net"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/util/subnet"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// +kops:fitask
type Subnet struct {
	Name *string

	// ShortName is a shorter name, for use in terraform outputs
	// ShortName is expected to be unique across all subnets in the cluster,
	// so it is typically set to the name of the Subnet, in the cluster spec.
	ShortName *string

	Lifecycle *fi.Lifecycle

	ID               *string
	VPC              *VPC
	AvailabilityZone *string
	CIDR             *string
	IPv6CIDR         *string
	Shared           *bool

	Tags map[string]string
}

var _ fi.CompareWithID = &Subnet{}

func (e *Subnet) CompareWithID() *string {
	return e.ID
}

// OrderSubnetsById implements sort.Interface for []Subnet, based on ID
type OrderSubnetsById []*Subnet

func (a OrderSubnetsById) Len() int      { return len(a) }
func (a OrderSubnetsById) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderSubnetsById) Less(i, j int) bool {
	return fi.StringValue(a[i].ID) < fi.StringValue(a[j].ID)
}

func (e *Subnet) Find(c *fi.Context) (*Subnet, error) {
	subnet, err := e.findEc2Subnet(c)
	if err != nil {
		return nil, err
	}

	if subnet == nil {
		return nil, nil
	}

	actual := &Subnet{
		ID:               subnet.SubnetId,
		AvailabilityZone: subnet.AvailabilityZone,
		VPC:              &VPC{ID: subnet.VpcId},
		CIDR:             subnet.CidrBlock,
		Name:             findNameTag(subnet.Tags),
		Shared:           e.Shared,
		Tags:             intersectTags(subnet.Tags, e.Tags),
	}

	for _, association := range subnet.Ipv6CidrBlockAssociationSet {
		if association == nil || association.Ipv6CidrBlockState == nil {
			continue
		}

		state := aws.StringValue(association.Ipv6CidrBlockState.State)
		if state != ec2.SubnetCidrBlockStateCodeAssociated && state != ec2.SubnetCidrBlockStateCodeAssociating {
			continue
		}

		actual.IPv6CIDR = association.Ipv6CidrBlock
	}

	klog.V(2).Infof("found matching subnet %q", *actual.ID)
	e.ID = actual.ID

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle // Not materialized in AWS
	actual.ShortName = e.ShortName // Not materialized in AWS
	actual.Name = e.Name           // Name is part of Tags

	if isAutoAssignIPv6CIDR(e) {
		e.IPv6CIDR = actual.IPv6CIDR
	}

	return actual, nil
}

func (e *Subnet) findEc2Subnet(c *fi.Context) (*ec2.Subnet, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	request := &ec2.DescribeSubnetsInput{}
	if e.ID != nil {
		request.SubnetIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2().DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Subnets: %v", err)
	}
	if response == nil || len(response.Subnets) == 0 {
		return nil, nil
	}

	if len(response.Subnets) != 1 {
		klog.Fatalf("found multiple Subnets matching tags")
	}

	subnet := response.Subnets[0]
	return subnet, nil
}

func (e *Subnet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Subnet) CheckChanges(a, e, changes *Subnet) error {
	var errors field.ErrorList
	fieldPath := field.NewPath("Subnet")

	if a == nil {
		if e.VPC == nil {
			errors = append(errors, field.Required(fieldPath.Child("VPC"), "must specify a VPC"))
		}

		if e.CIDR == nil {
			// TODO: Auto-assign CIDR?
			errors = append(errors, field.Required(fieldPath.Child("CIDR"), "must specify a CIDR"))
		}
	}

	if a != nil {
		// TODO: Do we want to destroy & recreate the subnet when these immutable fields change?
		if changes.VPC != nil {
			var aID *string
			if a.VPC != nil {
				aID = a.VPC.ID
			}
			var eID *string
			if e.VPC != nil {
				eID = e.VPC.ID
			}
			errors = append(errors, fi.FieldIsImmutable(eID, aID, fieldPath.Child("VPC")))
		}
		if changes.AvailabilityZone != nil {
			errors = append(errors, fi.FieldIsImmutable(e.AvailabilityZone, a.AvailabilityZone, fieldPath.Child("AvailabilityZone")))
		}
		if changes.CIDR != nil {
			errors = append(errors, fi.FieldIsImmutable(e.CIDR, a.CIDR, fieldPath.Child("CIDR")))
		}
		if changes.IPv6CIDR != nil && a.IPv6CIDR != nil {
			errors = append(errors, fi.FieldIsImmutable(e.IPv6CIDR, a.IPv6CIDR, fieldPath.Child("IPv6CIDR")))
		}
	}

	if len(errors) != 0 {
		return errors[0]
	}

	return nil
}

var autoAssignCIDRLock sync.Mutex

func autoAssignCIDR(ec2Client ec2iface.EC2API, vpc *VPC, spec string) (*net.IPNet, error) {
	autoAssignCIDRLock.Lock()
	defer autoAssignCIDRLock.Unlock()

	vpcID := fi.StringValue(vpc.ID)
	_, vpcCIDR, err := net.ParseCIDR(fi.StringValue(vpc.IPv6CIDR))
	if err != nil {
		return nil, fmt.Errorf("VPC %q did not have valid ipv6CIDR %q", vpcID, fi.StringValue(vpc.IPv6CIDR))
	}

	request := &ec2.DescribeSubnetsInput{}
	request.Filters = append(request.Filters, &ec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: aws.StringSlice([]string{vpcID}),
	})

	response, err := ec2Client.DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Subnets: %v", err)
	}

	var inUse subnet.CIDRMap

	for _, subnet := range response.Subnets {
		for _, block := range subnet.Ipv6CidrBlockAssociationSet {
			cidrString := aws.StringValue(block.Ipv6CidrBlock)
			if cidrString == "" {
				continue
			}
			if err := inUse.MarkInUse(cidrString); err != nil {
				return nil, fmt.Errorf("invalid CIDR %q found for subnet %q", cidrString, aws.StringValue(subnet.SubnetId))
			}
		}
	}

	_, specCIDR, err := net.ParseCIDR(spec)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CIDR %q", spec)
	}

	allocated, err := inUse.Allocate(vpcCIDR.String(), specCIDR.Mask)
	if err != nil {
		return nil, fmt.Errorf("unable to allocate subnet of size %q in VPC %q: %w", spec, vpcID, err)
	}

	return allocated, nil
}

func isAutoAssignIPv6CIDR(e *Subnet) bool {
	// A hack ... should we just recognize /64 ?
	return fi.StringValue(e.IPv6CIDR) == "::/64"
}

func (_ *Subnet) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Subnet) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the subnet was found
		if a == nil {
			return fmt.Errorf("Subnet with id %q not found", fi.StringValue(e.ID))
		}
	}

	if a == nil {

		ipv6CIDR := e.IPv6CIDR

		// Special case: auto IPAM
		if isAutoAssignIPv6CIDR(e) {
			cidr, err := autoAssignCIDR(t.Cloud.EC2(), e.VPC, fi.StringValue(ipv6CIDR))
			if err != nil {
				return err
			}
			klog.Infof("assigned cidr %v", cidr)
			ipv6CIDR = fi.String(cidr.String())
		}

		klog.V(2).Infof("Creating Subnet with CIDR: %q", *e.CIDR)

		request := &ec2.CreateSubnetInput{
			CidrBlock:         e.CIDR,
			Ipv6CidrBlock:     ipv6CIDR,
			AvailabilityZone:  e.AvailabilityZone,
			VpcId:             e.VPC.ID,
			TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeSubnet, e.Tags),
		}

		klog.Infof("creating subnet %+v", request)

		response, err := t.Cloud.EC2().CreateSubnet(request)
		if err != nil {
			return fmt.Errorf("error creating subnet: %v", err)
		}

		e.ID = response.Subnet.SubnetId
	}

	// We can't support this..
	// else {
	// 	if changes.IPv6CIDR != nil {
	// 		request := &ec2.AssociateSubnetCidrBlockInput{
	// 			Ipv6CidrBlock: ipv6CIDR,
	// 			SubnetId:      e.ID,
	// 		}

	// 		_, err := t.Cloud.EC2().AssociateSubnetCidrBlock(request)
	// 		if err != nil {
	// 			return fmt.Errorf("error associating subnet cidr block: %v", err)
	// 		}
	// 	}
	// }

	return t.AddAWSTags(*e.ID, e.Tags)
}

func subnetSlicesEqualIgnoreOrder(l, r []*Subnet) bool {
	var lIDs []string
	for _, s := range l {
		lIDs = append(lIDs, *s.ID)
	}
	var rIDs []string
	for _, s := range r {
		if s.ID == nil {
			klog.V(4).Infof("Subnet ID not set; returning not-equal: %v", s)
			return false
		}
		rIDs = append(rIDs, *s.ID)
	}
	return utils.StringSlicesEqualIgnoreOrder(lIDs, rIDs)
}

type terraformSubnet struct {
	VPCID            *terraformWriter.Literal `json:"vpc_id" cty:"vpc_id"`
	CIDR             *string                  `json:"cidr_block" cty:"cidr_block"`
	IPv6CIDR         *string                  `json:"ipv6_cidr_block" cty:"ipv6_cidr_block"`
	AvailabilityZone *string                  `json:"availability_zone" cty:"availability_zone"`
	Tags             map[string]string        `json:"tags,omitempty" cty:"tags"`
}

func (_ *Subnet) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Subnet) error {
	if fi.StringValue(e.ShortName) != "" {
		name := fi.StringValue(e.ShortName)
		if err := t.AddOutputVariable("subnet_"+name+"_id", e.TerraformLink()); err != nil {
			return err
		}
	}

	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed
		// We won't apply changes, but our validation (kops update) will still warn
		//
		// We probably shouldn't output subnet_ids only in this case - we normally output them by role,
		// but removing it now might break people.  We could always output subnet_ids though, if we
		// ever get a request for that.
		return t.AddOutputVariableArray("subnet_ids", terraformWriter.LiteralFromStringValue(*e.ID))
	}

	tf := &terraformSubnet{
		VPCID:            e.VPC.TerraformLink(),
		CIDR:             e.CIDR,
		IPv6CIDR:         e.IPv6CIDR,
		AvailabilityZone: e.AvailabilityZone,
		Tags:             e.Tags,
	}

	return t.RenderResource("aws_subnet", *e.Name, tf)
}

func (e *Subnet) TerraformLink() *terraformWriter.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if subnet is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing subnet with id %q", *e.ID)
		return terraformWriter.LiteralFromStringValue(*e.ID)
	}

	return terraformWriter.LiteralProperty("aws_subnet", *e.Name, "id")
}

type cloudformationSubnet struct {
	VPCID            *cloudformation.Literal `json:"VpcId,omitempty"`
	CIDR             *string                 `json:"CidrBlock,omitempty"`
	IPv6CIDR         *string                 `json:"Ipv6CidrBlock,omitempty"`
	AvailabilityZone *string                 `json:"AvailabilityZone,omitempty"`
	Tags             []cloudformationTag     `json:"Tags,omitempty"`
}

func (_ *Subnet) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *Subnet) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not cloudformation owned / managed
		// We won't apply changes, but our validation (kops update) will still warn
		return nil
	}

	cf := &cloudformationSubnet{
		VPCID:            e.VPC.CloudformationLink(),
		CIDR:             e.CIDR,
		IPv6CIDR:         e.IPv6CIDR,
		AvailabilityZone: e.AvailabilityZone,
		Tags:             buildCloudformationTags(e.Tags),
	}

	return t.RenderResource("AWS::EC2::Subnet", *e.Name, cf)
}

func (e *Subnet) CloudformationLink() *cloudformation.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if subnet is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing subnet with id %q", *e.ID)
		return cloudformation.LiteralString(*e.ID)
	}

	return cloudformation.Ref("AWS::EC2::Subnet", *e.Name)
}

func (e *Subnet) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	if e.ID == nil || aws.BoolValue(e.Shared) {
		return nil, nil
	}

	subnet, err := e.findEc2Subnet(c)
	if err != nil {
		return nil, err
	}

	if subnet == nil {
		return nil, nil
	}

	var removals []fi.Deletion
	for _, association := range subnet.Ipv6CidrBlockAssociationSet {
		// Skip when without state
		if association == nil || association.Ipv6CidrBlockState == nil {
			continue
		}

		// Skip when already disassociated
		state := aws.StringValue(association.Ipv6CidrBlockState.State)
		if state == ec2.SubnetCidrBlockStateCodeDisassociated || state == ec2.SubnetCidrBlockStateCodeDisassociating {
			continue
		}

		// Skip when current IPv6CIDR
		if aws.StringValue(e.IPv6CIDR) == aws.StringValue(association.Ipv6CidrBlock) {
			continue
		}

		if isAutoAssignIPv6CIDR(e) {
			continue
		}

		// Otherwise remove it!
		removals = append(removals, &deleteSubnetIPv6CIDRBlock{
			vpcID:         subnet.VpcId,
			ipv6CidrBlock: association.Ipv6CidrBlock,
			associationID: association.AssociationId,
		})
	}

	return removals, nil
}

type deleteSubnetIPv6CIDRBlock struct {
	vpcID         *string
	ipv6CidrBlock *string
	associationID *string
}

var _ fi.Deletion = &deleteSubnetIPv6CIDRBlock{}

func (d *deleteSubnetIPv6CIDRBlock) Delete(t fi.Target) error {
	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	request := &ec2.DisassociateSubnetCidrBlockInput{
		AssociationId: d.associationID,
	}
	klog.Infof("DisassociateSubnetCidrBlock: %+v", request)

	_, err := awsTarget.Cloud.EC2().DisassociateSubnetCidrBlock(request)
	return err
}

func (d *deleteSubnetIPv6CIDRBlock) TaskName() string {
	return "SubnetIPv6CIDRBlock"
}

func (d *deleteSubnetIPv6CIDRBlock) Item() string {
	return fmt.Sprintf("%v: ipv6cidr=%v", *d.vpcID, *d.ipv6CidrBlock)
}
