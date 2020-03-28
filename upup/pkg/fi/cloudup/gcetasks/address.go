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

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Address
type Address struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	IPAddress *string
}

func (e *Address) Find(c *fi.Context) (*Address, error) {
	actual, err := e.find(c.Cloud.(gce.GCECloud))
	if actual != nil && err == nil {
		if e.IPAddress == nil {
			e.IPAddress = actual.IPAddress
		}
	}
	return actual, err
}

func findAddressByIP(cloud gce.GCECloud, ip string) (*Address, error) {
	// Technically this is a regex, but it doesn't matter...
	r, err := cloud.Compute().Addresses.List(cloud.Project(), cloud.Region()).Filter("address eq " + ip).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing IP Addresses: %v", err)
	}

	if len(r.Items) == 0 {
		return nil, nil
	}
	if len(r.Items) > 1 {
		return nil, fmt.Errorf("found multiple Addresses matching %q", ip)
	}

	actual := &Address{}
	actual.IPAddress = &r.Items[0].Address
	actual.Name = &r.Items[0].Name

	return actual, nil
}

func (e *Address) find(cloud gce.GCECloud) (*Address, error) {
	r, err := cloud.Compute().Addresses.Get(cloud.Project(), cloud.Region(), *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing IP Addresses: %v", err)
	}

	actual := &Address{}
	actual.IPAddress = &r.Address
	actual.Name = &r.Name

	return actual, nil
}

var _ fi.HasAddress = &Address{}

func (e *Address) FindIPAddress(context *fi.Context) (*string, error) {
	actual, err := e.find(context.Cloud.(gce.GCECloud))
	if err != nil {
		return nil, fmt.Errorf("error querying for IP Address: %v", err)
	}
	if actual == nil {
		return nil, nil
	}
	return actual.IPAddress, nil
}

func (e *Address) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Address) CheckChanges(a, e, changes *Address) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.IPAddress != nil {
			return fi.CannotChangeField("Address")
		}
	}
	return nil
}

func (_ *Address) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Address) error {
	cloud := t.Cloud
	addr := &compute.Address{
		Name:    *e.Name,
		Address: fi.StringValue(e.IPAddress),
		Region:  cloud.Region(),
	}

	if a == nil {
		klog.Infof("GCE creating address: %q", addr.Name)

		op, err := cloud.Compute().Addresses.Insert(cloud.Project(), cloud.Region(), addr).Do()
		if err != nil {
			return fmt.Errorf("error creating IP Address: %v", err)
		}

		if err := cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for IP Address: %v", err)
		}
	} else {
		return fmt.Errorf("cannot apply changes to IP Address: %v", changes)
	}

	return nil
}

type terraformAddress struct {
	Name *string `json:"name,omitempty" cty:"name"`
}

func (_ *Address) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Address) error {
	tf := &terraformAddress{
		Name: e.Name,
	}
	return t.RenderResource("google_compute_address", *e.Name, tf)
}

func (e *Address) TerraformAddress() *terraform.Literal {
	name := fi.StringValue(e.Name)

	return terraform.LiteralProperty("google_compute_address", name, "address")
}
