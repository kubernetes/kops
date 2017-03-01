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

//go:generate fitask -type=IPAddress
type IPAddress struct {
	Name    *string
	Address *string
}

func (e *IPAddress) Find(c *fi.Context) (*IPAddress, error) {
	actual, err := e.find(c.Cloud.(*gce.GCECloud))
	if actual != nil && err == nil {
		if e.Address == nil {
			e.Address = actual.Address
		}
	}
	return actual, err
}

func (e *IPAddress) find(cloud *gce.GCECloud) (*IPAddress, error) {
	r, err := cloud.Compute.Addresses.Get(cloud.Project, cloud.Region, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing IPAddresss: %v", err)
	}

	actual := &IPAddress{}
	actual.Address = &r.Address
	actual.Name = &r.Name

	return actual, nil
}

var _ fi.HasAddress = &IPAddress{}

func (e *IPAddress) FindAddress(context *fi.Context) (*string, error) {
	actual, err := e.find(context.Cloud.(*gce.GCECloud))
	if err != nil {
		return nil, fmt.Errorf("error querying for IPAddress: %v", err)
	}
	if actual == nil {
		return nil, nil
	}
	return actual.Address, nil
}

func (e *IPAddress) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *IPAddress) CheckChanges(a, e, changes *IPAddress) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Address != nil {
			return fi.CannotChangeField("Address")
		}
	}
	return nil
}

func (_ *IPAddress) RenderGCE(t *gce.GCEAPITarget, a, e, changes *IPAddress) error {
	addr := &compute.Address{
		Name:    *e.Name,
		Address: fi.StringValue(e.Address),
		Region:  t.Cloud.Region,
	}

	if a == nil {
		glog.Infof("GCE creating address: %q", addr.Name)

		_, err := t.Cloud.Compute.Addresses.Insert(t.Cloud.Project, t.Cloud.Region, addr).Do()
		if err != nil {
			return fmt.Errorf("error creating IPAddress: %v", err)
		}
	} else {
		return fmt.Errorf("Cannot apply changes to IPAddress: %v", changes)
	}

	return nil
}

type terraformAddress struct {
	Name *string `json:"name"`
}

func (_ *IPAddress) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IPAddress) error {
	tf := &terraformAddress{
		Name: e.Name,
	}
	return t.RenderResource("google_compute_address", *e.Name, tf)
}
