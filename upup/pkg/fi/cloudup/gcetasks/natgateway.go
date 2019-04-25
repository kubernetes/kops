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

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=NatGateway
type NatGateway struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Network *string
	Region  *string
}

var _ fi.CompareWithID = &NatGateway{}

func (e *NatGateway) CompareWithID() *string {
	return e.Name
}

func (e *NatGateway) Find(c *fi.Context) (*NatGateway, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Routers.Get(cloud.Project(), *e.Region, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing NatGateways: %v", err)
	}

	actual := &NatGateway{}
	actual.Name = &r.Name
	actual.Network = &r.Network
	actual.Region = &r.Region

	if r.SelfLink != e.URL(cloud.Project()) {
		klog.Warningf("SelfLink did not match URL: %q vs %q", r.SelfLink, e.URL(cloud.Project()))
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *NatGateway) URL(project string) string {
	u := gce.GoogleCloudURL{
		Version: "beta",
		Project: project,
		Name:    *e.Name,
		Type:    "routers",
		Global:  true,
	}
	return u.BuildURL()
}

func (e *NatGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *NatGateway) CheckChanges(a, e, changes *NatGateway) error {
	return nil
}

func (_ *NatGateway) RenderGCE(t *gce.GCEAPITarget, a, e, changes *NatGateway) error {
	if a == nil {
		klog.V(2).Infof("Creating Cloud NAT Gateway %v", e.Name)

		router := &compute.Router{
			Name:    *e.Name,
			Network: *e.Network,

			Nats: []*compute.RouterNat{
				{
					// Nat IPs are allocated by Google Cloud.
					// TODO: support attaching static external IPs?
					NatIpAllocateOption: "AUTO_ONLY",

					// All of the IP ranges in every subnetwork are allowed to Nat.
					SourceSubnetworkIpRangesToNat: "ALL_SUBNETWORKS_ALL_IP_RANGES",
					Name:                          *e.Name,
				},
			},
		}
		_, err := t.Cloud.Compute().Routers.Insert(t.Cloud.Project(), *e.Region, router).Do()
		if err != nil {
			return fmt.Errorf("error creating Cloud NAT Gateway: %v", err)
		}
	} else {
		if !reflect.DeepEqual(changes, &NatGateway{}) {
			return fmt.Errorf("error applying changes to Cloud NAT router: %s", *e.Name)
		}
	}

	return nil
}

type terraformRouterNat struct {
	Name                          *string `json:"name"`
	Region                        *string `json:"region"`
	Router                        *string `json:"router"`
	NatIpAllocateOption           *string `json:"nat_ip_allocate_option"`
	SourceSubnetworkIpRangesToNat *string `json:"source_subnet_ip_ranges_to_nat"`
}

type terraformRouter struct {
	Name    *string `json:"name"`
	Network *string `json:"network"`
	Region  *string `json:"region,omitempty"`
}

func (_ *NatGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NatGateway) error {
	// Terraform 'google_compute_router_nat" requires a separate 'google_computer_router' resource.
	// Default googel_computer_router name to NAT gateway name.
	tr := &terraformRouter{
		Name:    e.Name,
		Network: e.Network,
	}

	trn := &terraformRouterNat{
		Name:                          e.Name,
		Region:                        e.Region,
		Router:                        e.Name,
		NatIpAllocateOption:           fi.String("AUTO_ONLY"),
		SourceSubnetworkIpRangesToNat: fi.String("ALL_SUBNETWORKS_ALL_IP_RANGES"),
	}

	err := t.RenderResource("google_compute_router", *e.Name, tr)
	if err != nil {
		return err
	}

	return t.RenderResource("google_compute_router_nat", *e.Name, trn)
}

func (i *NatGateway) TerraformName() *terraform.Literal {
	return terraform.LiteralProperty("google_compute_router_nat", *i.Name, "name")
}
