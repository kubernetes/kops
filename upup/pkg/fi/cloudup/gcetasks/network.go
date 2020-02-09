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

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Network
type Network struct {
	Name      *string
	Lifecycle *fi.Lifecycle
	Mode      string

	CIDR *string
}

var _ fi.CompareWithID = &Network{}

func (e *Network) CompareWithID() *string {
	return e.Name
}

func (e *Network) Find(c *fi.Context) (*Network, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Networks.Get(cloud.Project(), *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Networks: %v", err)
	}

	actual := &Network{}
	actual.Name = &r.Name
	if r.IPv4Range != "" {
		actual.Mode = "legacy"
		actual.CIDR = &r.IPv4Range
	} else if r.AutoCreateSubnetworks {
		actual.Mode = "auto"
	} else {
		actual.Mode = "custom"
	}

	if r.SelfLink != e.URL(cloud.Project()) {
		klog.Warningf("SelfLink did not match URL: %q vs %q", r.SelfLink, e.URL(cloud.Project()))
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

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
	cidr := fi.StringValue(e.CIDR)
	switch e.Mode {
	case "legacy":
		if cidr == "" {
			return fmt.Errorf("CIDR must specified for networks where mode=legacy")
		}
		klog.Warningf("using legacy mode for GCE network %q", fi.StringValue(e.Name))
	default:
		if cidr != "" {
			return fmt.Errorf("CIDR cannot specified for networks where mode=%s", e.Mode)
		}
	}

	switch e.Mode {
	case "auto":
	case "custom":
	case "legacy":

	default:
		return fmt.Errorf("unknown mode %q for Network", e.Mode)
	}

	return nil
}

func (_ *Network) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Network) error {
	if a == nil {
		klog.V(2).Infof("Creating Network with CIDR: %q", fi.StringValue(e.CIDR))

		network := &compute.Network{
			Name: *e.Name,
		}

		switch e.Mode {
		case "legacy":
			network.IPv4Range = fi.StringValue(e.CIDR)

		case "auto":
			network.AutoCreateSubnetworks = true

		case "custom":
			network.AutoCreateSubnetworks = false
		}
		_, err := t.Cloud.Compute().Networks.Insert(t.Cloud.Project(), network).Do()
		if err != nil {
			return fmt.Errorf("error creating Network: %v", err)
		}
	} else {
		if a.Mode == "legacy" {
			return fmt.Errorf("GCE networks in legacy mode are not supported.  Please convert to auto mode or specify a different network.")
		}
		empty := &Network{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to Network: %v", changes)
		}
	}

	return nil
}

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
