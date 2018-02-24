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

package gcetasks

import (
	"fmt"
	"reflect"

	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Subnet
type Subnet struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Network *Network
	Region  *string
	CIDR    *string

	SecondaryIpRanges map[string]string
}

var _ fi.CompareWithID = &Subnet{}

func (e *Subnet) CompareWithID() *string {
	return e.Name
}

func (e *Subnet) Find(c *fi.Context) (*Subnet, error) {
	cloud := c.Cloud.(gce.GCECloud)

	s, err := cloud.Compute().Subnetworks.Get(cloud.Project(), cloud.Region(), *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Subnets: %v", err)
	}

	actual := &Subnet{}
	actual.Name = &s.Name
	actual.Network = &Network{Name: fi.String(lastComponent(s.Network))}
	actual.Region = fi.String(lastComponent(s.Region))
	actual.CIDR = &s.IpCidrRange

	if len(s.SecondaryIpRanges) != 0 {
		actual.SecondaryIpRanges = make(map[string]string)
		for _, r := range s.SecondaryIpRanges {
			actual.SecondaryIpRanges[r.RangeName] = r.IpCidrRange
		}
	}

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *Subnet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Subnet) CheckChanges(a, e, changes *Subnet) error {
	return nil
}

func (_ *Subnet) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Subnet) error {
	project := t.Cloud.Project()

	if a == nil {
		glog.V(2).Infof("Creating Subnet with CIDR: %q", *e.CIDR)

		subnet := &compute.Subnetwork{
			IpCidrRange: *e.CIDR,
			Name:        *e.Name,
			Network:     e.Network.URL(project),
		}

		for k, v := range e.SecondaryIpRanges {
			subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
				RangeName:   k,
				IpCidrRange: v,
			})
		}

		_, err := t.Cloud.Compute().Subnetworks.Insert(t.Cloud.Project(), t.Cloud.Region(), subnet).Do()
		if err != nil {
			return fmt.Errorf("error creating Subnet: %v", err)
		}
	} else {
		empty := &Network{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to Subnet: %v", changes)
		}
	}

	return nil
}

func (e *Subnet) URL(project string, region string) string {
	u := gce.GoogleCloudURL{
		Version: "beta",
		Project: project,
		Name:    *e.Name,
		Type:    "subnetworks",
		Region:  region,
	}
	return u.BuildURL()
}

type terraformSubnet struct {
	Name             *string                `json:"name"`
	Network          *terraform.Literal     `json:"network"`
	Region           *string                `json:"region"`
	CIDR             *string                `json:"ip_cidr_range"`
	SecondaryIPRange []terraformSubnetRange `json:"secondary_ip_range,omitempty"`
}

type terraformSubnetRange struct {
	Name string `json:"range_name,omitempty"`
	CIDR string `json:"ip_cidr_range,omitempty"`
}

func (_ *Subnet) RenderSubnet(t *terraform.TerraformTarget, a, e, changes *Subnet) error {
	tf := &terraformSubnet{
		Name:    e.Name,
		Network: e.Network.TerraformName(),
		Region:  e.Region,
		CIDR:    e.CIDR,
	}

	for k, v := range e.SecondaryIpRanges {
		tf.SecondaryIPRange = append(tf.SecondaryIPRange, terraformSubnetRange{
			Name: k,
			CIDR: v,
		})
	}

	return t.RenderResource("google_compute_subnetwork", *e.Name, tf)
}

func (i *Subnet) TerraformName() *terraform.Literal {
	return terraform.LiteralProperty("google_compute_subnetwork", *i.Name, "name")
}
