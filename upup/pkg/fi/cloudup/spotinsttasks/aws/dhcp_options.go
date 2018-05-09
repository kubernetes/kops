/*
Copyright 2016 The Kubernetes Authors.

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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=DHCPOptions
type DHCPOptions struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ID                *string
	DomainName        *string
	DomainNameServers *string
}

var _ fi.CompareWithID = &DHCPOptions{}

func (e *DHCPOptions) CompareWithID() *string {
	return e.ID
}

func (e *DHCPOptions) Find(c *fi.Context) (*DHCPOptions, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

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

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

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

func (_ *DHCPOptions) Render(t *spotinst.Target, a, e, changes *DHCPOptions) error {
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

		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateDhcpOptions(request)
		if err != nil {
			return fmt.Errorf("error creating DHCPOptions: %v", err)
		}

		e.ID = response.DhcpOptions.DhcpOptionsId
	}

	return t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.ID,
		t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).BuildTags(e.Name))
}
