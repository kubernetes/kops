package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=InternetGateway
type InternetGateway struct {
	Name   *string
	ID     *string
	VPC    *VPC
	Shared *bool
}

var _ fi.CompareWithID = &InternetGateway{}

func (e *InternetGateway) CompareWithID() *string {
	return e.ID
}

func (e *InternetGateway) Find(c *fi.Context) (*InternetGateway, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeInternetGatewaysInput{}

	shared := fi.BoolValue(e.Shared)
	if shared {
		if fi.StringValue(e.VPC.ID) == "" {
			return nil, fmt.Errorf("VPC ID is required when InternetGateway is shared")
		}

		request.Filters = []*ec2.Filter{awsup.NewEC2Filter("attachment.vpc-id", *e.VPC.ID)}
	} else {
		if e.ID != nil {
			request.InternetGatewayIds = []*string{e.ID}
		} else {
			request.Filters = cloud.BuildFilters(e.Name)
		}
	}

	response, err := cloud.EC2.DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateways: %v", err)
	}
	if response == nil || len(response.InternetGateways) == 0 {
		return nil, nil
	}

	if len(response.InternetGateways) != 1 {
		return nil, fmt.Errorf("found multiple InternetGateways matching tags")
	}
	igw := response.InternetGateways[0]
	actual := &InternetGateway{
		ID:   igw.InternetGatewayId,
		Name: findNameTag(igw.Tags),
	}

	glog.V(2).Infof("found matching InternetGateway %q", *actual.ID)

	for _, attachment := range igw.Attachments {
		actual.VPC = &VPC{ID: attachment.VpcId}
	}

	// Prevent spurious comparison failures
	actual.Shared = e.Shared
	if e.ID == nil {
		e.ID = actual.ID
	}

	return actual, nil
}

func (e *InternetGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *InternetGateway) CheckChanges(a, e, changes *InternetGateway) error {
	if a != nil {
		// TODO: I think we can change it; we just detach & attach
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *InternetGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InternetGateway) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the InternetGateway was found and matches our required settings
		if a == nil {
			return fmt.Errorf("InternetGateway with id %q not found", fi.StringValue(e.ID))
		}

		return nil
	}

	if a == nil {
		glog.V(2).Infof("Creating InternetGateway")

		request := &ec2.CreateInternetGatewayInput{}

		response, err := t.Cloud.EC2.CreateInternetGateway(request)
		if err != nil {
			return fmt.Errorf("error creating InternetGateway: %v", err)
		}

		e.ID = response.InternetGateway.InternetGatewayId
	}

	if a == nil || (changes != nil && changes.VPC != nil) {
		glog.V(2).Infof("Creating InternetGatewayAttachment")

		attachRequest := &ec2.AttachInternetGatewayInput{
			VpcId:             e.VPC.ID,
			InternetGatewayId: e.ID,
		}

		_, err := t.Cloud.EC2.AttachInternetGateway(attachRequest)
		if err != nil {
			return fmt.Errorf("error attaching InternetGateway to VPC: %v", err)
		}
	}

	return t.AddAWSTags(*e.ID, t.Cloud.BuildTags(e.Name))
}

type terraformInternetGateway struct {
	VPCID *terraform.Literal `json:"vpc_id"`
	Tags  map[string]string  `json:"tags,omitempty"`
}

func (_ *InternetGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InternetGateway) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed
		return nil
	}

	cloud := t.Cloud.(*awsup.AWSCloud)

	tf := &terraformInternetGateway{
		VPCID: e.VPC.TerraformLink(),
		Tags:  cloud.BuildTags(e.Name),
	}

	return t.RenderResource("aws_internet_gateway", *e.Name, tf)
}

func (e *InternetGateway) TerraformLink() *terraform.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.ID == nil {
			glog.Fatalf("ID must be set, if InternetGateway is shared: %s", e)
		}

		glog.V(4).Infof("reusing existing InternetGateway with id %q", *e.ID)
		return terraform.LiteralFromStringValue(*e.ID)
	}

	return terraform.LiteralProperty("aws_internet_gateway", *e.Name, "id")
}
