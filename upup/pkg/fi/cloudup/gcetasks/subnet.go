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

package gcetasks

import (
	"fmt"
	"reflect"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Subnet
type Subnet struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	GCEName *string
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

	s, err := cloud.Compute().Subnetworks.Get(cloud.Project(), cloud.Region(), *e.GCEName).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Subnets: %v", err)
	}

	actual := &Subnet{}
	actual.Name = e.Name
	actual.GCEName = &s.Name
	actual.Network = &Network{Name: fi.String(lastComponent(s.Network))}
	actual.Region = fi.String(lastComponent(s.Region))
	actual.CIDR = &s.IpCidrRange

	{
		actual.SecondaryIpRanges = make(map[string]string)
		for _, r := range s.SecondaryIpRanges {
			if _, found := e.SecondaryIpRanges[r.RangeName]; found {
				actual.SecondaryIpRanges[r.RangeName] = r.IpCidrRange
			}
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
	cloud := t.Cloud
	project := cloud.Project()

	if a == nil {
		klog.V(2).Infof("Creating Subnet with CIDR: %q", fi.StringValue(e.CIDR))

		subnet := &compute.Subnetwork{
			IpCidrRange: *e.CIDR,
			Name:        *e.GCEName,
			Network:     e.Network.URL(project),
		}

		for k, v := range e.SecondaryIpRanges {
			subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
				RangeName:   k,
				IpCidrRange: v,
			})
		}

		_, err := cloud.Compute().Subnetworks.Insert(t.Cloud.Project(), t.Cloud.Region(), subnet).Do()
		if err != nil {
			return fmt.Errorf("error creating Subnet: %v", err)
		}
	} else {
		subnet, err := cloud.Compute().Subnetworks.Get(cloud.Project(), cloud.Region(), *e.GCEName).Do()
		if err != nil {
			return fmt.Errorf("error fetching subnet for patch: %v", err)
		}

		{
			rangeMap := make(map[string]string)
			for _, r := range subnet.SecondaryIpRanges {
				rangeMap[r.RangeName] = r.IpCidrRange
			}

			// Cannot add and remove ranges in the same call

			patch := true
			for k, v := range e.SecondaryIpRanges {
				if rangeMap[k] != v {
					rangeMap[k] = v
					subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
						RangeName:   k,
						IpCidrRange: v,
					})
					patch = true
				}
			}
			if patch {
				_, err = t.Cloud.Compute().Subnetworks.Patch(t.Cloud.Project(), t.Cloud.Region(), subnet.Name, subnet).Do()
				if err != nil {
					return fmt.Errorf("error patching Subnet: %v", err)
				}
				patch = false
				subnet, err = cloud.Compute().Subnetworks.Get(cloud.Project(), cloud.Region(), *e.GCEName).Do()
				if err != nil {
					return fmt.Errorf("error fetching subnet for patch: %v", err)
				}
				rangeMap = make(map[string]string)
				for _, r := range subnet.SecondaryIpRanges {
					rangeMap[r.RangeName] = r.IpCidrRange
				}
			}

			for k, v := range rangeMap {
				if e.SecondaryIpRanges[k] != v {
					delete(rangeMap, k)
					patch = true
				}
			}

			if patch {
				subnet.SecondaryIpRanges = nil
				for k, v := range rangeMap {
					subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
						RangeName:   k,
						IpCidrRange: v,
					})
				}
				_, err = t.Cloud.Compute().Subnetworks.Patch(t.Cloud.Project(), t.Cloud.Region(), subnet.Name, subnet).Do()
				if err != nil {
					return fmt.Errorf("error patching Subnet: %v", err)
				}
				patch = false
				_, err = cloud.Compute().Subnetworks.Get(cloud.Project(), cloud.Region(), *e.GCEName).Do()
				if err != nil {
					return fmt.Errorf("error fetching subnet for patch: %v", err)
				}

			}

			changes.SecondaryIpRanges = nil
		}

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
		Name:    *e.GCEName,
		Type:    "subnetworks",
		Region:  region,
	}
	return u.BuildURL()
}

type terraformSubnet struct {
	Name    *string            `json:"name"`
	Network *terraform.Literal `json:"network"`
	Region  *string            `json:"region"`
	CIDR    *string            `json:"ip_cidr_range"`

	// SecondaryIPRange defines additional IP ranges
	SecondaryIPRange []terraformSubnetRange `json:"secondary_ip_range,omitempty"`
}

type terraformSubnetRange struct {
	Name string `json:"range_name,omitempty"`
	CIDR string `json:"ip_cidr_range,omitempty"`
}

func (_ *Subnet) RenderSubnet(t *terraform.TerraformTarget, a, e, changes *Subnet) error {
	tf := &terraformSubnet{
		Name:    e.GCEName,
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
