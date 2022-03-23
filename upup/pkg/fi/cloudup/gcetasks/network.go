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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type Network struct {
	Name      *string
	Project   *string
	Lifecycle fi.Lifecycle
	Mode      string

	CIDR *string

	Shared *bool
}

var _ fi.CompareWithID = &Network{}

func (e *Network) CompareWithID() *string {
	return e.Name
}

func (e *Network) Find(c *fi.Context) (*Network, error) {
	cloud := c.Cloud.(gce.GCECloud)
	project := cloud.Project()
	if e.Project != nil {
		project = *e.Project
	}

	r, err := cloud.Compute().Networks().Get(project, *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Networks: %v", err)
	}

	actual := &Network{}
	if r.IPv4Range != "" {
		actual.Mode = "legacy"
		actual.CIDR = &r.IPv4Range
	} else if r.AutoCreateSubnetworks {
		actual.Mode = "auto"
	} else {
		actual.Mode = "custom"
	}
	actual.Project = &project

	if r.SelfLink != e.URL(project) {
		klog.Warningf("SelfLink did not match URL: %q vs %q", r.SelfLink, e.URL(project))
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared
	actual.Name = e.Name

	// Match unspecified values
	if e.Mode == "" {
		e.Mode = actual.Mode
	}

	return actual, nil
}

func (e *Network) URL(project string) string {
	u := gce.GoogleCloudURL{
		Version: "v1",
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
		// Known

	case "":
		// Treated as "keep existing", only allowed for shared mode
		if !fi.BoolValue(e.Shared) {
			return fmt.Errorf("must specify mode for (non-shared) Network")
		}

	default:
		return fmt.Errorf("unknown mode %q for Network", e.Mode)
	}

	return nil
}

func (_ *Network) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Network) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the network was found
		if a == nil {
			return fmt.Errorf("Network with name %q not found", fi.StringValue(e.Name))
		}
	}

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
			// The boolean default value of "false" is omitted when the struct
			// is serialized, which results in the network being created with
			// the auto-create subnetworks default of "true". Explicitly send
			// the default value.
			network.ForceSendFields = []string{"AutoCreateSubnetworks"}

		default:
			return fmt.Errorf("unhandled mode %q", e.Mode)
		}

		op, err := t.Cloud.Compute().Networks().Insert(t.Cloud.Project(), network)
		if err != nil {
			return fmt.Errorf("error creating Network: %v", err)
		}
		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for Network creation to complete: %w", err)
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
	Name                  *string `cty:"name"`
	IPv4Range             *string `cty:"ipv4_range"`
	Project               *string `cty:"project"`
	AutoCreateSubnetworks *bool   `cty:"auto_create_subnetworks"`
}

func (_ *Network) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Network) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed
		return nil
	}

	tf := &terraformNetwork{
		Name:    e.Name,
		Project: e.Project,
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

func (e *Network) TerraformLink() *terraformWriter.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.Name == nil {
			klog.Fatalf("Name must be set, if network is shared: %#v", e)
		}

		klog.V(4).Infof("reusing existing network with name %q", *e.Name)
		name := *e.Name
		if e.Project != nil {
			name = *e.Project + "/" + name
		}
		return terraformWriter.LiteralFromStringValue(name)
	}

	return terraformWriter.LiteralProperty("google_compute_network", *e.Name, "name")
}
