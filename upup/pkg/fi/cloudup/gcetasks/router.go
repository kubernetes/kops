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
)

//go:generate fitask -type=Router
type Router struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Network *string
	Region  *string
}

var _ fi.CompareWithID = &Router{}

func (e *Router) CompareWithID() *string {
	return e.Name
}

func (e *Router) Find(c *fi.Context) (*Router, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Routers.Get(cloud.Project(), *e.Region, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Routers: %v", err)
	}

	actual := &Router{}
	actual.Name = &r.Name
	actual.Network = &r.Network

	if r.SelfLink != e.URL(cloud.Project()) {
		glog.Warningf("SelfLink did not match URL: %q vs %q", r.SelfLink, e.URL(cloud.Project()))
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *Router) URL(project string) string {
	u := gce.GoogleCloudURL{
		Version: "beta",
		Project: project,
		Name:    *e.Name,
		Type:    "routers",
		Global:  true,
	}
	return u.BuildURL()
}

func (e *Router) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Router) CheckChanges(a, e, changes *Router) error {
	return nil
}

func (_ *Router) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Router) error {
	if a == nil {
		glog.V(2).Infof("Creating Cloud NAT %v", e.Name)

		router := &compute.Router{
			Name:    *e.Name,
			Network: *e.Network,

			Nats: []*compute.RouterNat{
				&compute.RouterNat{
					Name:                          *e.Name,
					NatIpAllocateOption:           "AUTO_ONLY",
					SourceSubnetworkIpRangesToNat: "ALL_SUBNETWORKS_ALL_IP_RANGES",
				},
			},
		}
		_, err := t.Cloud.Compute().Routers.Insert(t.Cloud.Project(), *e.Region, router).Do()
		if err != nil {
			return fmt.Errorf("error creating Cloud NAT: %v", err)
		}

	} else {
		// TODO: what is this?
	}

	return nil
}

/*
type terraformNetwork struct {
	Name                  *string `json:"name"`
	IPv4Range             *string `json:"ipv4_range,omitempty"`
	AutoCreateSubnetworks *bool   `json:"auto_create_subnetworks,omitempty"`
}

func (_ *Network) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Network) error {
	tf := &terraformNetwork{
		Name: e.Name,
	}

	switch e.Mode {
	case "legacy":
		tf.IPv4Range = e.CIDR

	case "auto":
		tf.AutoCreateSubnetworks = fi.Bool(true)

	case "custom":
		tf.AutoCreateSubnetworks = fi.Bool(false)
	}

	return t.RenderResource("google_compute_network", *e.Name, tf)
}

func (i *Network) TerraformName() *terraform.Literal {
	return terraform.LiteralProperty("google_compute_network", *i.Name, "name")
}
*/
