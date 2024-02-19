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

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type Address struct {
	Name      *string
	Lifecycle fi.Lifecycle

	IPAddress     *string
	IPAddressType *string
	Purpose       *string
	Region        string

	Subnetwork *Subnet

	// WellKnownServices indicates which services are supported by this resource.
	// This field is internal and is not rendered to the cloud.
	WellKnownServices []wellknownservices.WellKnownService
}

var _ fi.CompareWithID = &ForwardingRule{}

func (e *Address) CompareWithID() *string {
	return e.Name
}

func (e *Address) Find(c *fi.CloudupContext) (*Address, error) {
	actual, err := e.find(c.T.Cloud.(gce.GCECloud), e.Region)
	if actual != nil && err == nil {
		if e.IPAddress == nil {
			e.IPAddress = actual.IPAddress
		}

		// Ignore system fields
		actual.Lifecycle = e.Lifecycle
		actual.WellKnownServices = e.WellKnownServices
	}
	return actual, err
}

func findAddressByIP(cloud gce.GCECloud, ip, region string) (*Address, error) {
	// Technically this is a regex, but it doesn't matter...
	addrs, err := cloud.Compute().Addresses().ListWithFilter(cloud.Project(), region, "address eq "+ip)
	if err != nil {
		return nil, fmt.Errorf("error listing IP Addresses: %v", err)
	}

	if len(addrs) == 0 {
		return nil, nil
	}
	if len(addrs) > 1 {
		return nil, fmt.Errorf("found multiple Addresses matching %q", ip)
	}

	actual := &Address{}
	actual.IPAddress = &addrs[0].Address
	actual.IPAddressType = &addrs[0].AddressType
	actual.Purpose = &addrs[0].Purpose
	actual.Name = &addrs[0].Name

	return actual, nil
}

func (e *Address) find(cloud gce.GCECloud, region string) (*Address, error) {
	r, err := cloud.Compute().Addresses().Get(cloud.Project(), region, *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing IP Addresses: %v", err)
	}

	actual := &Address{}
	actual.IPAddress = &r.Address
	actual.IPAddressType = &r.AddressType
	actual.Purpose = &r.Purpose
	actual.Name = &r.Name
	if e.Subnetwork != nil {
		actual.Subnetwork = &Subnet{
			Name: fi.PtrTo(lastComponent(r.Subnetwork)),
		}
	}

	return actual, nil
}

var _ fi.HasAddress = &Address{}

// GetWellKnownServices implements fi.HasAddress::GetWellKnownServices.
// It indicates which services we support with this address (likely attached to a load balancer).
func (e *Address) GetWellKnownServices() []wellknownservices.WellKnownService {
	return e.WellKnownServices
}

func (e *Address) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	actual, err := e.find(context.T.Cloud.(gce.GCECloud), e.Region)
	if err != nil {
		return nil, fmt.Errorf("error querying for IP Address: %v", err)
	}
	if actual == nil {
		return nil, nil
	}
	return []string{fi.ValueOf(actual.IPAddress)}, nil
}

func (e *Address) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
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
		Name:        *e.Name,
		Address:     fi.ValueOf(e.IPAddress),
		AddressType: fi.ValueOf(e.IPAddressType),
		Purpose:     fi.ValueOf(e.Purpose),
		Region:      e.Region,
	}

	if e.Subnetwork != nil {
		addr.Subnetwork = e.Subnetwork.URL(t.Cloud.Project(), t.Cloud.Region())
	}

	if a == nil {
		klog.V(2).Infof("Creating Address: %q", addr.Name)

		op, err := cloud.Compute().Addresses().Insert(cloud.Project(), e.Region, addr)
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
	Name        *string                  `cty:"name"`
	AddressType *string                  `cty:"address_type"`
	Purpose     *string                  `cty:"purpose"`
	Subnetwork  *terraformWriter.Literal `cty:"subnetwork"`
	Region      string                   `cty:"region"`
}

func (*Address) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Address) error {
	tf := &terraformAddress{
		Name:        e.Name,
		AddressType: e.IPAddressType,
		Purpose:     e.Purpose,
	}
	if e.Subnetwork != nil {
		tf.Subnetwork = e.Subnetwork.TerraformLink()
	}
	if e.Region == "" {
		return t.RenderResource("google_compute_global_address", *e.Name, tf)
	}
	tf.Region = e.Region
	return t.RenderResource("google_compute_address", *e.Name, tf)
}

func (e *Address) TerraformAddress() *terraformWriter.Literal {
	name := fi.ValueOf(e.Name)
	if e.Region == "" {
		return terraformWriter.LiteralProperty("google_compute_global_address", name, "address")
	}
	return terraformWriter.LiteralProperty("google_compute_address", name, "address")
}
