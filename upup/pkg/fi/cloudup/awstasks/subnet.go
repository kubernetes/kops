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

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/utils"
)

//go:generate fitask -type=Subnet
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

	klog.V(2).Infof("found matching subnet %q", *actual.ID)
	e.ID = actual.ID

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle // Not materialized in AWS
	actual.ShortName = e.ShortName // Not materialized in AWS
	actual.Name = e.Name           // Name is part of Tags

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
			errors = append(errors, fi.FieldIsImmutable(a.AvailabilityZone, e.AvailabilityZone, fieldPath.Child("AvailabilityZone")))
		}
		if changes.CIDR != nil {
			errors = append(errors, fi.FieldIsImmutable(a.CIDR, e.CIDR, fieldPath.Child("CIDR")))
		}
	}

	if len(errors) != 0 {
		return errors[0]
	}

	return nil
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
		klog.V(2).Infof("Creating Subnet with CIDR: %q", *e.CIDR)

		request := &ec2.CreateSubnetInput{
			CidrBlock:        e.CIDR,
			AvailabilityZone: e.AvailabilityZone,
			VpcId:            e.VPC.ID,
		}

		response, err := t.Cloud.EC2().CreateSubnet(request)
		if err != nil {
			return fmt.Errorf("error creating subnet: %v", err)
		}

		e.ID = response.Subnet.SubnetId
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
	VPCID            *terraform.Literal `json:"vpc_id"`
	CIDR             *string            `json:"cidr_block"`
	AvailabilityZone *string            `json:"availability_zone"`
	Tags             map[string]string  `json:"tags,omitempty"`
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
		return t.AddOutputVariableArray("subnet_ids", terraform.LiteralFromStringValue(*e.ID))
	}

	tf := &terraformSubnet{
		VPCID:            e.VPC.TerraformLink(),
		CIDR:             e.CIDR,
		AvailabilityZone: e.AvailabilityZone,
		Tags:             e.Tags,
	}

	return t.RenderResource("aws_subnet", *e.Name, tf)
}

func (e *Subnet) TerraformLink() *terraform.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if subnet is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing subnet with id %q", *e.ID)
		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralProperty("aws_subnet", *e.Name, "id")
}

type cloudformationSubnet struct {
	VPCID            *cloudformation.Literal `json:"VpcId,omitempty"`
	CIDR             *string                 `json:"CidrBlock,omitempty"`
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
