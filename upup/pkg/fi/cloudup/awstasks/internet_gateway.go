package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
)

//go:generate fitask -type=InternetGateway
type InternetGateway struct {
	Name *string
	ID   *string
}

var _ fi.CompareWithID = &InternetGateway{}

func (e *InternetGateway) CompareWithID() *string {
	return e.ID
}

func (e *InternetGateway) Find(c *fi.Context) (*InternetGateway, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeInternetGatewaysInput{}
	if e.ID != nil {
		request.InternetGatewayIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
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

	e.ID = actual.ID

	return actual, nil
}

func (e *InternetGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *InternetGateway) CheckChanges(a, e, changes *InternetGateway) error {
	if a != nil {
	}
	return nil
}

func (_ *InternetGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InternetGateway) error {
	if a == nil {
		glog.V(2).Infof("Creating InternetGateway")

		request := &ec2.CreateInternetGatewayInput{}

		response, err := t.Cloud.EC2.CreateInternetGateway(request)
		if err != nil {
			return fmt.Errorf("error creating InternetGateway: %v", err)
		}

		e.ID = response.InternetGateway.InternetGatewayId
	}

	return t.AddAWSTags(*e.ID, t.Cloud.BuildTags(e.Name, nil))
}
