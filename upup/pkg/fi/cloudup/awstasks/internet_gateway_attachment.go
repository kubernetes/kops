package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
)

type InternetGatewayAttachment struct {
	VPC             *VPC
	InternetGateway *InternetGateway
}

func (e *InternetGatewayAttachment) String() string {
	return fi.TaskAsString(e)
}

func (e *InternetGatewayAttachment) Find(c *fi.Context) (*InternetGatewayAttachment, error) {
	if e.InternetGateway == nil {
		return nil, fi.RequiredField("InternetGateway")
	}
	if e.VPC == nil {
		return nil, fi.RequiredField("VPC")
	}

	if e.VPC.ID == nil {
		return nil, nil
	}
	if e.InternetGateway.ID == nil {
		return nil, nil
	}

	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []*string{e.InternetGateway.ID},
	}

	response, err := cloud.EC2.DescribeInternetGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateways: %v", err)
	}
	if response == nil || len(response.InternetGateways) == 0 {
		return nil, nil
	}

	if len(response.InternetGateways) != 1 {
		return nil, fmt.Errorf("found multiple InternetGateways matching ID")
	}
	igw := response.InternetGateways[0]
	for _, attachment := range igw.Attachments {
		if aws.StringValue(attachment.VpcId) == *e.VPC.ID {
			actual := &InternetGatewayAttachment{
				VPC:             &VPC{ID: e.VPC.ID},
				InternetGateway: &InternetGateway{ID: e.InternetGateway.ID},
			}
			glog.V(2).Infof("found matching InternetGateway")
			return actual, nil
		}
	}

	return nil, nil
}

func (e *InternetGatewayAttachment) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *InternetGatewayAttachment) CheckChanges(a, e, changes *InternetGatewayAttachment) error {
	if a != nil {
		// TODO: I think we can change it; we just detach & attach
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}
	return nil
}

func (_ *InternetGatewayAttachment) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InternetGatewayAttachment) error {
	if a == nil {
		glog.V(2).Infof("Creating InternetGatewayAttachment")

		attachRequest := &ec2.AttachInternetGatewayInput{
			VpcId:             e.VPC.ID,
			InternetGatewayId: e.InternetGateway.ID,
		}

		_, err := t.Cloud.EC2.AttachInternetGateway(attachRequest)
		if err != nil {
			return fmt.Errorf("error attaching InternetGatewayAttachment: %v", err)
		}
	}

	return nil // No tags
}
