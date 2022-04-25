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
type Subnet struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Network *Network
	Region  *string
	CIDR    *string

	SecondaryIpRanges map[string]string

	Shared *bool
}

var _ fi.CompareWithID = &Subnet{}

func (e *Subnet) CompareWithID() *string {
	return e.Name
}

func (e *Subnet) Find(c *fi.Context) (*Subnet, error) {
	cloud := c.Cloud.(gce.GCECloud)
	_, project, err := gce.ParseNameAndProjectFromNetworkID(c.Cluster.Spec.NetworkID)
	if err != nil {
		return nil, fmt.Errorf("error parsing network name from cluster spec: %w", err)
	} else if project == "" {
		project = cloud.Project()
	}
	s, err := cloud.Compute().Subnetworks().Get(project, cloud.Region(), *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Subnets: %w", err)
	}

	actual := &Subnet{}
	actual.Name = &s.Name
	actual.Network = &Network{Name: fi.String(lastComponent(s.Network))}
	actual.Region = fi.String(lastComponent(s.Region))
	actual.CIDR = &s.IpCidrRange

	shared := fi.BoolValue(e.Shared)
	{
		actual.SecondaryIpRanges = make(map[string]string)
		for _, r := range s.SecondaryIpRanges {
			if shared {
				// In the shared case, only show differences on the ranges we specified
				if _, found := e.SecondaryIpRanges[r.RangeName]; !found {
					continue
				}
			}

			actual.SecondaryIpRanges[r.RangeName] = r.IpCidrRange
		}
	}

	// Prevent spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Name = e.Name
	actual.Shared = e.Shared

	return actual, nil
}

func (e *Subnet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Subnet) CheckChanges(a, e, changes *Subnet) error {
	return nil
}

func (_ *Subnet) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Subnet) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Verify the subnet was found
		if a == nil {
			return fmt.Errorf("Subnet with name %q not found", fi.StringValue(e.Name))
		}
	}

	cloud := t.Cloud
	project := cloud.Project()

	if a == nil {
		klog.V(2).Infof("Creating Subnet with CIDR: %q", fi.StringValue(e.CIDR))

		subnet := &compute.Subnetwork{
			IpCidrRange: fi.StringValue(e.CIDR),
			Name:        *e.Name,
			Network:     e.Network.URL(project),
		}

		for k, v := range e.SecondaryIpRanges {
			subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
				RangeName:   k,
				IpCidrRange: v,
			})
		}

		op, err := cloud.Compute().Subnetworks().Insert(t.Cloud.Project(), t.Cloud.Region(), subnet)
		if err != nil {
			return fmt.Errorf("error creating Subnet: %v", err)
		}
		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for Subnet creation to complete: %w", err)
		}
	} else {
		if changes.SecondaryIpRanges != nil {
			// Update is split into two calls as GCE does not allow us to add and remove ranges in the same call
			if err := updateSecondaryRanges(cloud, "add", e); err != nil {
				return err
			}

			if !shared {
				if err := updateSecondaryRanges(cloud, "remove", e); err != nil {
					return err
				}
			}

			changes.SecondaryIpRanges = nil
		}

		empty := &Subnet{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to Subnet: %v", changes)
		}
	}

	return nil
}

func updateSecondaryRanges(cloud gce.GCECloud, op string, e *Subnet) error {
	// We need to refetch to patch it
	subnet, err := cloud.Compute().Subnetworks().Get(cloud.Project(), cloud.Region(), *e.Name)
	if err != nil {
		return fmt.Errorf("error fetching subnet for patch: %w", err)
	}

	expectedRanges := e.SecondaryIpRanges

	actualRanges := make(map[string]string)
	for _, r := range subnet.SecondaryIpRanges {
		actualRanges[r.RangeName] = r.IpCidrRange
	}

	// Cannot add and remove ranges in the same call
	if op == "add" {
		patch := false
		for k, v := range expectedRanges {
			if actualRanges[k] != v {
				actualRanges[k] = v
				subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
					RangeName:   k,
					IpCidrRange: v,
				})
				patch = true
			}
		}

		if !patch {
			return nil
		}
	} else if op == "remove" {
		patch := false
		if len(actualRanges) != len(expectedRanges) {
			patch = true
		} else {
			for k := range expectedRanges {
				if actualRanges[k] != e.SecondaryIpRanges[k] {
					patch = true
				}
			}
		}

		if !patch {
			return nil
		}

		subnet.SecondaryIpRanges = nil
		for k, v := range expectedRanges {
			subnet.SecondaryIpRanges = append(subnet.SecondaryIpRanges, &compute.SubnetworkSecondaryRange{
				RangeName:   k,
				IpCidrRange: v,
			})
		}
	}

	_, err = cloud.Compute().Subnetworks().Patch(cloud.Project(), cloud.Region(), subnet.Name, subnet)
	if err != nil {
		return fmt.Errorf("error patching Subnet: %w", err)
	}

	return nil
}

func (e *Subnet) URL(project string, region string) string {
	u := gce.GoogleCloudURL{
		Version: "v1",
		Project: project,
		Name:    *e.Name,
		Type:    "subnetworks",
		Region:  region,
	}
	return u.BuildURL()
}

type terraformSubnet struct {
	Name    *string                  `cty:"name"`
	Network *terraformWriter.Literal `cty:"network"`
	Region  *string                  `cty:"region"`
	CIDR    *string                  `cty:"ip_cidr_range"`

	// SecondaryIPRange defines additional IP ranges
	SecondaryIPRange []terraformSubnetRange `cty:"secondary_ip_range"`
}

type terraformSubnetRange struct {
	Name string `cty:"range_name"`
	CIDR string `cty:"ip_cidr_range"`
}

func (_ *Subnet) RenderSubnet(t *terraform.TerraformTarget, a, e, changes *Subnet) error {
	shared := fi.BoolValue(e.Shared)
	if shared {
		// Not terraform owned / managed
		return nil
	}

	tf := &terraformSubnet{
		Name:    e.Name,
		Network: e.Network.TerraformLink(),
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

func (e *Subnet) TerraformLink() *terraformWriter.Literal {
	shared := fi.BoolValue(e.Shared)
	if shared {
		if e.Name == nil {
			klog.Fatalf("GCEName must be set, if subnet is shared: %#v", e)
		}

		name := *e.Name
		if e.Network != nil && e.Network.Project != nil {
			name = *e.Network.Project + "/" + name
		}
		klog.V(4).Infof("reusing existing subnet with name %q", name)
		return terraformWriter.LiteralFromStringValue(name)
	}

	return terraformWriter.LiteralProperty("google_compute_subnetwork", *e.Name, "name")
}
