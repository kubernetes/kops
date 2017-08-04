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
	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Network
type Network struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	CIDR *string
}

var _ fi.CompareWithID = &Network{}

func (e *Network) CompareWithID() *string {
	return e.Name
}

func (e *Network) Find(c *fi.Context) (*Network, error) {
	cloud := c.Cloud.(*gce.GCECloud)

	r, err := cloud.Compute.Networks.Get(cloud.Project, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Networks: %v", err)
	}

	actual := &Network{}
	actual.Name = &r.Name
	actual.CIDR = &r.IPv4Range

	if r.SelfLink != e.URL(cloud.Project) {
		glog.Warningf("SelfLink did not match URL: %q vs %q", r.SelfLink, e.URL(cloud.Project))
	}

	return actual, nil
}

func (e *Network) URL(project string) string {
	u := gce.GoogleCloudURL{
		Version: "beta",
		Project: project,
		Name:    *e.Name,
		Type:    "networks",
		Global:  true,
	}
	return u.BuildURL()
}

func (e *Network) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Network) CheckChanges(a, e, changes *Network) error {
	return nil
}

func (_ *Network) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Network) error {
	if a == nil {
		glog.V(2).Infof("Creating Network with CIDR: %q", *e.CIDR)

		network := &compute.Network{
			IPv4Range: *e.CIDR,

			//// AutoCreateSubnetworks: When set to true, the network is created in
			//// "auto subnet mode". When set to false, the network is in "custom
			//// subnet mode".
			////
			//// In "auto subnet mode", a newly created network is assigned the
			//// default CIDR of 10.128.0.0/9 and it automatically creates one
			//// subnetwork per region.
			//AutoCreateSubnetworks bool `json:"autoCreateSubnetworks,omitempty"`

			Name: *e.Name,
		}
		_, err := t.Cloud.Compute.Networks.Insert(t.Cloud.Project, network).Do()
		if err != nil {
			return fmt.Errorf("error creating Network: %v", err)
		}
	}

	return nil
}

type terraformNetwork struct {
	Name *string `json:"name"`
	CIDR *string `json:"ipv4_range"`
	//AutoCreateSubnetworks bool `json:"auto_create_subnetworks"`
}

func (_ *Network) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Network) error {
	tf := &terraformNetwork{
		Name: e.Name,
		CIDR: e.CIDR,
		//AutoCreateSubnetworks: false,
	}

	return t.RenderResource("google_compute_network", *e.Name, tf)
}

func (i *Network) TerraformName() *terraform.Literal {
	return terraform.LiteralProperty("google_compute_network", *i.Name, "name")
}
