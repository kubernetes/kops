package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
)

//go:generate fitask -type=Subnet
type Subnet struct {
	Name             *string
	ID               *string
	VPC              *VPC
	AvailabilityZone *string
	CIDR             *string
}

var _ fi.CompareWithID = &Subnet{}

func (e *Subnet) CompareWithID() *string {
	return e.ID
}

func (e *Subnet) Find(c *fi.Context) (*Subnet, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeSubnetsInput{}
	if e.ID != nil {
		request.SubnetIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2.DescribeSubnets(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Subnets: %v", err)
	}
	if response == nil || len(response.Subnets) == 0 {
		return nil, nil
	}

	if len(response.Subnets) != 1 {
		glog.Fatalf("found multiple Subnets matching tags")
	}

	subnet := response.Subnets[0]
	actual := &Subnet{
		ID:               subnet.SubnetId,
		AvailabilityZone: subnet.AvailabilityZone,
		VPC:              &VPC{ID: subnet.VpcId},
		CIDR:             subnet.CidrBlock,
		Name:             findNameTag(subnet.Tags),
	}

	glog.V(2).Infof("found matching subnet %q", *actual.ID)
	e.ID = actual.ID

	return actual, nil
}

func (e *Subnet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Subnet) CheckChanges(a, e, changes *Subnet) error {
	if a == nil {
		if e.VPC == nil {
			return fi.RequiredField("VPC")
		}

		if e.CIDR == nil {
			// TODO: Auto-assign CIDR?
			return fi.RequiredField("CIDR")
		}
	}

	if a != nil {
		if changes.VPC != nil {
			// TODO: Do we want to destroy & recreate the subnet?
			return fi.CannotChangeField("VPC")
		}
		if changes.AvailabilityZone != nil {
			// TODO: Do we want to destroy & recreate the subnet?
			return fi.CannotChangeField("AvailabilityZone")
		}
		if changes.CIDR != nil {
			// TODO: Do we want to destroy & recreate the subnet?
			return fi.CannotChangeField("CIDR")
		}
	}
	return nil
}

func (_ *Subnet) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Subnet) error {
	if a == nil {
		glog.V(2).Infof("Creating Subnet with CIDR: %q", *e.CIDR)

		request := &ec2.CreateSubnetInput{
			CidrBlock:        e.CIDR,
			AvailabilityZone: e.AvailabilityZone,
			VpcId:            e.VPC.ID,
		}

		response, err := t.Cloud.EC2.CreateSubnet(request)
		if err != nil {
			return fmt.Errorf("error creating subnet: %v", err)
		}

		e.ID = response.Subnet.SubnetId
	}

	return t.AddAWSTags(*e.ID, t.Cloud.BuildTags(e.Name))
}

func subnetSlicesEqualIgnoreOrder(l, r []*Subnet) bool {
	var lIDs []string
	for _, s := range l {
		lIDs = append(lIDs, *s.ID)
	}
	var rIDs []string
	for _, s := range r {
		if s.ID == nil {
			glog.V(4).Infof("Subnet ID not set; returning not-equal: %v", s)
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
	cloud := t.Cloud.(*awsup.AWSCloud)

	tf := &terraformSubnet{
		VPCID:            e.VPC.TerraformLink(),
		CIDR:             e.CIDR,
		AvailabilityZone: e.AvailabilityZone,
		Tags:             cloud.BuildTags(e.Name),
	}

	return t.RenderResource("aws_subnet", *e.Name, tf)
}

func (e *Subnet) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_subnet", *e.Name, "id")
}
