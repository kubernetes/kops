/*
Copyright 2021 The Kubernetes Authors.

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
	"strings"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

const (
	// NATIPAllocationOptionAutoOnly is specified when NAT IPs are allocated by Google Cloud.
	NATIPAllocationOptionAutoOnly = "AUTO_ONLY"

	// SourceSubnetworkIPRangesAll is specified when all of the IP ranges in every subnetwork are allowed to be NAT-ed.
	SourceSubnetworkIPRangesAll = "ALL_SUBNETWORKS_ALL_IP_RANGES"
	// SourceSubnetworkIPRangesSpecificSubnets is specified when we should NAT only specific listed subnets.
	SourceSubnetworkIPRangesSpecificSubnets = "LIST_OF_SUBNETWORKS"

	// subnetNatAllIPRanges specifies that we should NAT all IP ranges in the subnet.
	subnetNatAllIPRanges = "ALL_IP_RANGES"
)

// +kops:fitask

// Router is a Router task.
type Router struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Network *Network
	Region  *string

	NATIPAllocationOption         *string
	SourceSubnetworkIPRangesToNAT *string

	Subnetworks []*Subnet
}

var _ fi.CompareWithID = &Router{}

// CompareWithID returns the name of the Router.
func (r *Router) CompareWithID() *string {
	return r.Name
}

// Find discovers the Router in the cloud provider.
func (r *Router) Find(c *fi.Context) (*Router, error) {
	cloud := c.Cloud.(gce.GCECloud)

	found, err := cloud.Compute().Routers().Get(cloud.Project(), *r.Region, *r.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Routers: %w", err)
	}

	if len(found.Nats) != 1 {
		return nil, fmt.Errorf("unexpected number of nats found: %+v", found.Nats)
	}
	nat := found.Nats[0]

	if a, e := found.SelfLink, r.url(cloud.Project()); a != e {
		klog.Warningf("SelfLink did not match URL: %q vs %q", a, e)
	}

	actual := &Router{
		Name:                          &found.Name,
		Lifecycle:                     r.Lifecycle,
		Network:                       &Network{Name: fi.String(lastComponent(found.Network))},
		Region:                        fi.String(lastComponent(found.Region)),
		NATIPAllocationOption:         &nat.NatIpAllocateOption,
		SourceSubnetworkIPRangesToNAT: &nat.SourceSubnetworkIpRangesToNat,
	}

	for _, subnet := range nat.Subnetworks {
		if strings.Join(subnet.SourceIpRangesToNat, ",") != subnetNatAllIPRanges {
			klog.Warningf("ignoring NAT router %q with nats.subnetworks.sourceIpRangesToNat != %s", found.SelfLink, subnetNatAllIPRanges)
			return nil, nil
		}

		subnetName := lastComponent(subnet.Name)
		actual.Subnetworks = append(actual.Subnetworks, &Subnet{Name: &subnetName})
	}
	return actual, nil

}

func (r *Router) url(project string) string {
	u := gce.GoogleCloudURL{
		Version: "v1",
		Project: project,
		Name:    *r.Name,
		Type:    "routers",
		Region:  *r.Region,
	}
	return u.BuildURL()
}

// Run implements fi.Task.Run.
func (r *Router) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(r, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*Router) CheckChanges(a, e, changes *Router) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchanegable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}

	// TODO(kenji): Check more fields.

	return nil
}

// RenderGCE creates or updates a Router.
func (*Router) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Router) error {
	cloud := t.Cloud
	project := cloud.Project()
	region := fi.StringValue(e.Region)

	if a == nil {
		klog.V(2).Infof("Creating Cloud NAT Gateway %v", e.Name)
		router := &compute.Router{
			Name:    *e.Name,
			Network: e.Network.URL(project),
			Nats: []*compute.RouterNat{
				{
					Name:                          *e.Name,
					NatIpAllocateOption:           *e.NATIPAllocationOption,
					SourceSubnetworkIpRangesToNat: *e.SourceSubnetworkIPRangesToNAT,
				},
			},
		}

		for _, subnet := range e.Subnetworks {
			router.Nats[0].Subnetworks = append(router.Nats[0].Subnetworks, &compute.RouterNatSubnetworkToNat{
				Name:                subnet.URL(project, region),
				SourceIpRangesToNat: []string{subnetNatAllIPRanges},
			})
		}
		op, err := t.Cloud.Compute().Routers().Insert(project, region, router)
		if err != nil {
			return fmt.Errorf("error creating Router: %w", err)
		}
		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for router creation: %w", err)
		}
	} else {
		if !reflect.DeepEqual(changes, &Router{}) {
			return fmt.Errorf("applying changes to Router is unsupported: %s", *e.Name)
		}
	}

	return nil
}

type terraformRouterNat struct {
	Name                          *string                         `json:"name,omitempty" cty:"name"`
	Region                        *string                         `json:"region,omitempty" cty:"region"`
	Router                        *string                         `json:"router,omitempty" cty:"router"`
	NATIPAllocateOption           *string                         `json:"nat_ip_allocate_option,omitempty" cty:"nat_ip_allocate_option"`
	SourceSubnetworkIPRangesToNat *string                         `json:"source_subnetwork_ip_ranges_to_nat,omitempty" cty:"source_subnetwork_ip_ranges_to_nat"`
	Subnetworks                   []*terraformRouterNatSubnetwork `json:"subnetwork,omitempty" cty:"subnetwork"`
}

type terraformRouterNatSubnetwork struct {
	Name                *terraformWriter.Literal `json:"name,omitempty" cty:"name"`
	SourceIPRangesToNat []string                 `json:"source_ip_ranges_to_nat,omitempty" cty:"source_ip_ranges_to_nat"`
}

type terraformRouter struct {
	Name    *string                  `json:"name,omitempty" cty:"name"`
	Network *terraformWriter.Literal `json:"network,omitempty" cty:"network"`
	Region  *string                  `json:"region,omitempty" cty:"region"`
}

// RenderTerraform renders the Terraform config.
func (*Router) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Router) error {
	tr := &terraformRouter{
		Name: e.Name,
	}
	if e.Network != nil {
		tr.Network = e.Network.TerraformLink()
	}
	err := t.RenderResource("google_compute_router", *e.Name, tr)
	if err != nil {
		return err
	}

	trn := &terraformRouterNat{
		Name:                          e.Name,
		Region:                        e.Region,
		Router:                        e.Name,
		NATIPAllocateOption:           e.NATIPAllocationOption,
		SourceSubnetworkIPRangesToNat: e.SourceSubnetworkIPRangesToNAT,
	}
	for _, subnet := range e.Subnetworks {
		trn.Subnetworks = append(trn.Subnetworks, &terraformRouterNatSubnetwork{
			Name:                subnet.TerraformLink(),
			SourceIPRangesToNat: []string{subnetNatAllIPRanges},
		})
	}
	return t.RenderResource("google_compute_router_nat", *e.Name, trn)
}

// TerraformLink returns an expression for the name.
func (r *Router) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("google_compute_router_nat", *r.Name, "name")
}
