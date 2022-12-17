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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type DHCPOptions struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID                *string
	DomainName        *string
	DomainNameServers *string

	// Shared is set if this is a shared DHCPOptions
	Shared *bool

	// Tags is a map of aws tags that are added to the InternetGateway
	Tags map[string]string
}

var _ fi.CompareWithID = &DHCPOptions{}

func (e *DHCPOptions) CompareWithID() *string {
	return e.ID
}

func (e *DHCPOptions) Find(c *fi.CloudupContext) (*DHCPOptions, error) {
	cloud := c.T.Cloud.(awsup.AWSCloud)

	request := &ec2.DescribeDhcpOptionsInput{}
	if e.ID != nil {
		request.DhcpOptionsIds = []*string{e.ID}
	} else {
		request.Filters = cloud.BuildFilters(e.Name)
	}

	response, err := cloud.EC2().DescribeDhcpOptions(request)
	if err != nil {
		return nil, fmt.Errorf("error listing DHCPOptions: %v", err)
	}

	if response == nil || len(response.DhcpOptions) == 0 {
		return nil, nil
	}

	if len(response.DhcpOptions) != 1 {
		return nil, fmt.Errorf("found multiple DhcpOptions with name: %s", *e.Name)
	}
	klog.V(2).Info("found existing DhcpOptions")
	o := response.DhcpOptions[0]
	actual := &DHCPOptions{
		ID:   o.DhcpOptionsId,
		Name: findNameTag(o.Tags),
		Tags: intersectTags(o.Tags, e.Tags),
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
			klog.Infof("Skipping over DHCPOption with key=%q value=%q", k, v)
		}
	}

	e.ID = actual.ID

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared

	return actual, nil
}

func (e *DHCPOptions) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
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

func (_ *DHCPOptions) RenderAWS(ctx *fi.CloudupContext, t *awsup.AWSAPITarget, a, e, changes *DHCPOptions) error {
	if a == nil {
		klog.V(2).Infof("Creating DHCPOptions with Name:%q", *e.Name)

		request := &ec2.CreateDhcpOptionsInput{
			TagSpecifications: awsup.EC2TagSpecification(ec2.ResourceTypeDhcpOptions, e.Tags),
		}
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

		response, err := t.Cloud.EC2().CreateDhcpOptions(request)
		if err != nil {
			return fmt.Errorf("error creating DHCPOptions: %v", err)
		}

		e.ID = response.DhcpOptions.DhcpOptionsId
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

type terraformDHCPOptions struct {
	DomainName        *string           `cty:"domain_name"`
	DomainNameServers []string          `cty:"domain_name_servers"`
	Tags              map[string]string `cty:"tags"`
}

func (_ *DHCPOptions) RenderTerraform(ctx *fi.CloudupContext, t *terraform.TerraformTarget, a, e, changes *DHCPOptions) error {
	tf := &terraformDHCPOptions{
		DomainName: e.DomainName,
		Tags:       e.Tags,
	}

	if e.DomainNameServers != nil {
		tf.DomainNameServers = strings.Split(*e.DomainNameServers, ",")
	}

	return t.RenderResource("aws_vpc_dhcp_options", *e.Name, tf)
}

func (e *DHCPOptions) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("aws_vpc_dhcp_options", *e.Name, "id")
}
