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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
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

	Lifecycle fi.Lifecycle

	ID                          *string
	VPC                         *VPC
	AmazonIPv6CIDR              *VPCAmazonIPv6CIDRBlock
	AvailabilityZone            *string
	CIDR                        *string
	IPv6CIDR                    *string
	ResourceBasedNaming         *bool
	AssignIPv6AddressOnCreation *bool
	Shared                      *bool

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
	return fi.ValueOf(a[i].ID) < fi.ValueOf(a[j].ID)
}

func (e *Subnet) Find(c *fi.CloudupContext) (*Subnet, error) {
	e.AssignIPv6AddressOnCreation = fi.PtrTo(e.IPv6CIDR != nil)

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
		break
	}

	actual.AssignIPv6AddressOnCreation = subnet.AssignIpv6AddressOnCreation

	actual.ResourceBasedNaming = fi.PtrTo(aws.StringValue(subnet.PrivateDnsNameOptionsOnLaunch.HostnameType) == ec2.HostnameTypeResourceName)
	if *actual.ResourceBasedNaming {
		if fi.ValueOf(actual.CIDR) != "" && !aws.BoolValue(subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord) {
			actual.ResourceBasedNaming = nil
		}
		if fi.ValueOf(actual.IPv6CIDR) != "" && !aws.BoolValue(subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord) {
			actual.ResourceBasedNaming = nil
		}
	}

	klog.V(2).Infof("found matching subnet %q", *actual.ID)
	e.ID = actual.ID

	// Calculate expected IPv6 CIDR if possible and in the "/64#N" format
	if e.VPC.IPv6CIDR != nil && strings.HasPrefix(aws.StringValue(e.IPv6CIDR), "/") {
		subnetIPv6CIDR, err := calculateSubnetCIDR(e.VPC.IPv6CIDR, e.IPv6CIDR)
		if err != nil {
			return nil, err
		}
		e.IPv6CIDR = subnetIPv6CIDR
	}

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle // Not materialized in AWS
	actual.ShortName = e.ShortName // Not materialized in AWS
	actual.Name = e.Name           // Name is part of Tags
	// Task dependencies
	actual.AmazonIPv6CIDR = e.AmazonIPv6CIDR

	return actual, nil
}

func (e *Subnet) findEc2Subnet(c *fi.CloudupContext) (*ec2.Subnet, error) {
	cloud := c.T.Cloud.(awsup.AWSCloud)

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

func (e *Subnet) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (s *Subnet) CheckChanges(a, e, changes *Subnet) error {
	var errors field.ErrorList
	fieldPath := field.NewPath("Subnet")

	if a == nil {
		if e.VPC == nil {
			errors = append(errors, field.Required(fieldPath.Child("VPC"), "must specify a VPC"))
		}

		if e.CIDR == nil && e.IPv6CIDR == nil {
			errors = append(errors, field.Required(fieldPath.Child("CIDR"), "must specify a CIDR or IPv6CIDR"))
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

		if fi.ValueOf(e.Shared) {
			if changes.IPv6CIDR != nil && a.IPv6CIDR == nil {
				errors = append(errors, field.Forbidden(fieldPath.Child("IPv6CIDR"), "field cannot be set on shared subnet"))
			}
		}
	}

	if len(errors) != 0 {
		return errors[0]
	}

	return nil
}

func (_ *Subnet) ShouldCreate(a, e, changes *Subnet) (bool, error) {
	if fi.ValueOf(e.Shared) {
		changes.ResourceBasedNaming = nil
		return changes.Tags != nil, nil
	}
	return true, nil
}

func (_ *Subnet) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Subnet) error {
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Verify the subnet was found
		if a == nil {
			return fmt.Errorf("subnet with id %q not found", fi.ValueOf(e.ID))
		}
	}

	if strings.HasPrefix(aws.StringValue(e.IPv6CIDR), "/") {
		vpcIPv6CIDR := e.VPC.IPv6CIDR
		if vpcIPv6CIDR == nil {
			cidr, err := findVPCIPv6CIDR(t.Cloud, e.VPC.ID)
			if err != nil {
				return err
			}
			if cidr == nil {
				return fi.NewTryAgainLaterError("waiting for the VPC IPv6 CIDR to be assigned")
			}
			vpcIPv6CIDR = cidr
		}

		subnetIPv6CIDR, err := calculateSubnetCIDR(vpcIPv6CIDR, e.IPv6CIDR)
		if err != nil {
			return err
		}
		e.IPv6CIDR = subnetIPv6CIDR
	}

	if a == nil {
		klog.V(2).Infof("Creating Subnet with CIDR: %q IPv6CIDR: %q", fi.ValueOf(e.CIDR), fi.ValueOf(e.IPv6CIDR))

		request := &ec2.CreateSubnetInput{
			CidrBlock:         e.CIDR,
			Ipv6CidrBlock:     e.IPv6CIDR,
			AvailabilityZone:  e.AvailabilityZone,
			VpcId:             e.VPC.ID,
			TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeSubnet, e.Tags),
		}

		if e.CIDR == nil {
			request.Ipv6Native = aws.Bool(true)
		}

		response, err := t.Cloud.EC2().CreateSubnet(request)
		if err != nil {
			return fmt.Errorf("error creating subnet: %v", err)
		}

		e.ID = response.Subnet.SubnetId
	} else {
		if changes.IPv6CIDR != nil {
			request := &ec2.AssociateSubnetCidrBlockInput{
				Ipv6CidrBlock: e.IPv6CIDR,
				SubnetId:      e.ID,
			}

			_, err := t.Cloud.EC2().AssociateSubnetCidrBlock(request)
			if err != nil {
				return fmt.Errorf("error associating subnet cidr block: %v", err)
			}
		}
	}

	if a == nil || changes.AssignIPv6AddressOnCreation != nil {
		request := &ec2.ModifySubnetAttributeInput{
			SubnetId:                    e.ID,
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{Value: e.AssignIPv6AddressOnCreation},
		}
		_, err := t.Cloud.EC2().ModifySubnetAttribute(request)
		if err != nil {
			return fmt.Errorf("error modifying AssignIPv6AddressOnCreation: %w", err)
		}
	}

	if changes.ResourceBasedNaming != nil {
		hostnameType := ec2.HostnameTypeIpName
		if *changes.ResourceBasedNaming {
			hostnameType = ec2.HostnameTypeResourceName
		}
		request := &ec2.ModifySubnetAttributeInput{
			SubnetId:                       e.ID,
			PrivateDnsHostnameTypeOnLaunch: &hostnameType,
		}
		_, err := t.Cloud.EC2().ModifySubnetAttribute(request)
		if err != nil {
			return fmt.Errorf("error modifying hostname type: %w", err)
		}

		if fi.ValueOf(e.CIDR) == "" {
			request = &ec2.ModifySubnetAttributeInput{
				SubnetId:    e.ID,
				EnableDns64: &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
			}
			_, err = t.Cloud.EC2().ModifySubnetAttribute(request)
			if err != nil {
				return fmt.Errorf("error enabling DNS64: %w", err)
			}
		} else {
			request = &ec2.ModifySubnetAttributeInput{
				SubnetId:                             e.ID,
				EnableResourceNameDnsARecordOnLaunch: &ec2.AttributeBooleanValue{Value: changes.ResourceBasedNaming},
			}
			_, err = t.Cloud.EC2().ModifySubnetAttribute(request)
			if err != nil {
				return fmt.Errorf("error modifying A records: %w", err)
			}
		}

		if fi.ValueOf(e.IPv6CIDR) != "" {
			request = &ec2.ModifySubnetAttributeInput{
				SubnetId:                                e.ID,
				EnableResourceNameDnsAAAARecordOnLaunch: &ec2.AttributeBooleanValue{Value: changes.ResourceBasedNaming},
			}
			_, err = t.Cloud.EC2().ModifySubnetAttribute(request)
			if err != nil {
				return fmt.Errorf("error modifying AAAA records: %w", err)
			}
		}
	}

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
	VPCID                                   *terraformWriter.Literal `cty:"vpc_id"`
	CIDR                                    *string                  `cty:"cidr_block"`
	IPv6CIDR                                *terraformWriter.Literal `cty:"ipv6_cidr_block"`
	IPv6Native                              *bool                    `cty:"ipv6_native"`
	AssignIPv6AddressOnCreation             *bool                    `cty:"assign_ipv6_address_on_creation"`
	AvailabilityZone                        *string                  `cty:"availability_zone"`
	EnableDNS64                             *bool                    `cty:"enable_dns64"`
	EnableResourceNameDNSAAAARecordOnLaunch *bool                    `cty:"enable_resource_name_dns_aaaa_record_on_launch"`
	EnableResourceNameDNSARecordOnLaunch    *bool                    `cty:"enable_resource_name_dns_a_record_on_launch"`
	PrivateDNSHostnameTypeOnLaunch          *string                  `cty:"private_dns_hostname_type_on_launch"`
	Tags                                    map[string]string        `cty:"tags"`
}

func (_ *Subnet) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Subnet) error {
	if fi.ValueOf(e.ShortName) != "" {
		name := fi.ValueOf(e.ShortName)
		if err := t.AddOutputVariable("subnet_"+name+"_id", e.TerraformLink()); err != nil {
			return err
		}
	}

	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed
		// We won't apply changes, but our validation (kops update) will still warn
		//
		// We probably shouldn't output subnet_ids only in this case - we normally output them by role,
		// but removing it now might break people.  We could always output subnet_ids though, if we
		// ever get a request for that.
		return t.AddOutputVariableArray("subnet_ids", terraformWriter.LiteralFromStringValue(*e.ID))
	}

	var ipv6CIDR *terraformWriter.Literal
	if strings.HasPrefix(aws.StringValue(e.IPv6CIDR), "/") {
		newSize, netNum, err := utils.ParseCIDRNotation(*e.IPv6CIDR)
		if err != nil {
			return fmt.Errorf("error parsing CIDR subnet: %v", err)
		}

		newBits := terraformWriter.LiteralBinaryExpression(
			terraformWriter.LiteralFromIntValue(newSize),
			"-",
			terraformWriter.LiteralTokens("local", "vpc_ipv6_cidr_length"),
		)
		ipv6CIDR = terraformWriter.LiteralFunctionExpression("cidrsubnet",
			terraformWriter.LiteralTokens("local", "vpc_ipv6_cidr_block"),
			newBits,
			terraformWriter.LiteralFromIntValue(netNum),
		)
	} else if e.IPv6CIDR != nil {
		ipv6CIDR = terraformWriter.LiteralFromStringValue(aws.StringValue(e.IPv6CIDR))
	}

	tf := &terraformSubnet{
		VPCID:            e.VPC.TerraformLink(),
		CIDR:             e.CIDR,
		IPv6CIDR:         ipv6CIDR,
		AvailabilityZone: e.AvailabilityZone,
		Tags:             e.Tags,
	}
	if fi.ValueOf(e.CIDR) == "" {
		tf.EnableDNS64 = fi.PtrTo(true)
		tf.IPv6Native = fi.PtrTo(true)
	}
	if fi.ValueOf(e.IPv6CIDR) != "" {
		tf.AssignIPv6AddressOnCreation = fi.PtrTo(true)
	}
	if e.ResourceBasedNaming != nil {
		hostnameType := ec2.HostnameTypeIpName
		if *e.ResourceBasedNaming {
			hostnameType = ec2.HostnameTypeResourceName
		}
		tf.PrivateDNSHostnameTypeOnLaunch = fi.PtrTo(hostnameType)
		if fi.ValueOf(e.CIDR) != "" {
			tf.EnableResourceNameDNSARecordOnLaunch = e.ResourceBasedNaming
		}
		if fi.ValueOf(e.IPv6CIDR) != "" {
			tf.EnableResourceNameDNSAAAARecordOnLaunch = e.ResourceBasedNaming
		}
	}

	return t.RenderResource("aws_subnet", *e.Name, tf)
}

func (e *Subnet) TerraformLink() *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if subnet is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing subnet with id %q", *e.ID)
		return terraformWriter.LiteralFromStringValue(*e.ID)
	}

	return terraformWriter.LiteralProperty("aws_subnet", *e.Name, "id")
}

func (e *Subnet) FindDeletions(c *fi.CloudupContext) ([]fi.CloudupDeletion, error) {
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

	var removals []fi.CloudupDeletion
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

var _ fi.CloudupDeletion = &deleteSubnetIPv6CIDRBlock{}

func (d *deleteSubnetIPv6CIDRBlock) Delete(t fi.CloudupTarget) error {
	awsTarget, ok := t.(*awsup.AWSAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	request := &ec2.DisassociateSubnetCidrBlockInput{
		AssociationId: d.associationID,
	}
	_, err := awsTarget.Cloud.EC2().DisassociateSubnetCidrBlock(request)
	return err
}

func (d *deleteSubnetIPv6CIDRBlock) TaskName() string {
	return "SubnetIPv6CIDRBlock"
}

func (d *deleteSubnetIPv6CIDRBlock) Item() string {
	return fmt.Sprintf("%v: ipv6cidr=%v", *d.vpcID, *d.ipv6CidrBlock)
}

func calculateSubnetCIDR(vpcCIDR, subnetCIDR *string) (*string, error) {
	if vpcCIDR == nil {
		return nil, fmt.Errorf("expecting VPC CIDR to not be <nil>")
	}
	if subnetCIDR == nil {
		return nil, fmt.Errorf("expecting subnet CIDR to not be <nil>")
	}
	if !strings.HasPrefix(aws.StringValue(subnetCIDR), "/") {
		return nil, fmt.Errorf("expecting subnet cidr to start with %q: %q", "/", aws.StringValue(subnetCIDR))
	}

	newSize, netNum, err := utils.ParseCIDRNotation(aws.StringValue(subnetCIDR))
	if err != nil {
		return nil, fmt.Errorf("error parsing CIDR subnet: %v", err)
	}

	newCIDR, err := utils.CIDRSubnet(aws.StringValue(vpcCIDR), newSize, netNum)
	if err != nil {
		return nil, fmt.Errorf("error calculating CIDR subnet: %v", err)
	}

	return aws.String(newCIDR), nil
}
