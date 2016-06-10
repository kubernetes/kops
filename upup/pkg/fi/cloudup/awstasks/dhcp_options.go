package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"strings"
)

//go:generate fitask -type=DHCPOptions
type DHCPOptions struct {
	Name *string

	ID                *string
	DomainName        *string
	DomainNameServers *string
}

var _ fi.CompareWithID = &DHCPOptions{}

func (e *DHCPOptions) CompareWithID() *string {
	return e.ID
}

func (e *DHCPOptions) Find(c *fi.Context) (*DHCPOptions, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	request := &ec2.DescribeDhcpOptionsInput{}
	if e.ID != nil {
		request.DhcpOptionsIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2.DescribeDhcpOptions(request)
	if err != nil {
		return nil, fmt.Errorf("error listing DHCPOptions: %v", err)
	}

	if response == nil || len(response.DhcpOptions) == 0 {
		return nil, nil
	}

	if len(response.DhcpOptions) != 1 {
		return nil, fmt.Errorf("found multiple DhcpOptions with name: %s", *e.Name)
	}
	glog.V(2).Info("found existing DhcpOptions")
	o := response.DhcpOptions[0]
	actual := &DHCPOptions{
		ID:   o.DhcpOptionsId,
		Name: findNameTag(o.Tags),
	}

	for _, s := range o.DhcpConfigurations {
		k := aws.StringValue(s.Key)
		v := ""
		for _, av := range s.Values {
			if v != "" {
				v = v + ","
			}
			v = v + *av.Value
		}
		switch k {
		case "domain-name":
			actual.DomainName = &v
		case "domain-name-servers":
			actual.DomainNameServers = &v
		default:
			glog.Infof("Skipping over DHCPOption with key=%q value=%q", k, v)
		}
	}

	e.ID = actual.ID

	return actual, nil
}

func (e *DHCPOptions) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *DHCPOptions) CheckChanges(a, e, changes *DHCPOptions) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	if a != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}

		// TODO: Delete & create new DHCPOptions
		// We can't delete the DHCPOptions while it is attached, but we can change the tag (add a timestamp suffix?)
		if changes.DomainName != nil {
			return fi.CannotChangeField("DomainName")
		}
		if changes.DomainNameServers != nil {
			return fi.CannotChangeField("DomainNameServers")
		}
	}
	return nil
}

func (_ *DHCPOptions) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *DHCPOptions) error {
	if a == nil {
		glog.V(2).Infof("Creating DHCPOptions with Name:%q", *e.Name)

		request := &ec2.CreateDhcpOptionsInput{}
		if e.DomainNameServers != nil {
			o := &ec2.NewDhcpConfiguration{
				Key:    aws.String("domain-name-servers"),
				Values: []*string{e.DomainNameServers},
			}
			request.DhcpConfigurations = append(request.DhcpConfigurations, o)
		}
		if e.DomainName != nil {
			o := &ec2.NewDhcpConfiguration{
				Key:    aws.String("domain-name"),
				Values: []*string{e.DomainName},
			}
			request.DhcpConfigurations = append(request.DhcpConfigurations, o)
		}

		response, err := t.Cloud.EC2.CreateDhcpOptions(request)
		if err != nil {
			return fmt.Errorf("error creating DHCPOptions: %v", err)
		}

		e.ID = response.DhcpOptions.DhcpOptionsId
	}

	return t.AddAWSTags(*e.ID, t.Cloud.BuildTags(e.Name))
}

type terraformDHCPOptions struct {
	DomainName        *string           `json:"domain_name,omitempty"`
	DomainNameServers []string          `json:"domain_name_servers,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

func (_ *DHCPOptions) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *DHCPOptions) error {
	cloud := t.Cloud.(*awsup.AWSCloud)

	tf := &terraformDHCPOptions{
		DomainName: e.DomainName,
		Tags:       cloud.BuildTags(e.Name),
	}

	if e.DomainNameServers != nil {
		tf.DomainNameServers = strings.Split(*e.DomainNameServers, ",")
	}

	return t.RenderResource("aws_vpc_dhcp_options", *e.Name, tf)
}

func (e *DHCPOptions) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_vpc_dhcp_options", *e.Name, "id")
}
